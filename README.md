# machgo
[![Build Status](https://travis-ci.com//machgo.svg?branch=master)](https://travis-ci.com/daihasso/MachGo)[![Coverage Status](https://coveralls.io/repos/github/DaiHasso/MachGo/badge.svg?branch=master)](https://coveralls.io/github/DaiHasso/MachGo?branch=master)

machgo is a ORM-ish database library for go.

Because I am chronically lazy, this readme is a WIP.

Check out the [godocs!](https://godoc.org/github.com//machgo)

## Example Usage
### Simple usage

Let's say you have the following table:
``` sql
CREATE TABLE images (
  id bigint PRIMARY KEY,
  post_id bigint NOT NULL,
  filename text,
  mime_type text,
  created date DEFAULT now(),
  updated date DEFAULT now()
);
```

Then you might create the following object representation:
``` go
import (
    "github.com/daihasso/machgo/database"
)

type Image struct {
  PostID int64 `db:"post_id"`

  MimeType string `db:"mime_type"`
  OriginalFileName string `db:"file_name"`
}
```

Next you might want to create a database connection:
``` go
import (
    "sync"

    "github.com/daihasso/machgo/pool"
    "github.com/daihasso/machgo/pool/config"
)

var MyDBConn *pool.ConnectionPool

func init() {
    once.Do(func() {
        MyDBConn, err := config.PostgresPool(
            config.Username("mypostgresuser"),
            config.Password("mysecretpassword"),
            config.Host("localhost"),
            config.Port(5432),
            config.DatabaseName("mycooldatabase"),
        )
        if err != nil {
            panic(err)
        }

        // This lets us share a connection across further machgo calls.
        SetGlobalConnectionPool(MyDBConn)
    })
}
```

And to retrieve single image you might do:
``` go

import(
    "github.com/daihasso/machgo/pool/ses"
)

func getImage(id int64) *Image {
    session, err := sess.NewSession()
    if err != nil {
        panic(err)
    }
    result := &Image{}
    err := session.GetObject(result, id)
    if err != nil {
        panic(err)
    }

    return result
}
```

### More advanced usage
Let's say you want many images per post so you expand the previous table by
using a lookup table:
``` sql
CREATE TABLE images (
  id bigint PRIMARY KEY,
  filename text,
  mime_type text,
  created date DEFAULT now(),
  updated date DEFAULT now()
);

CREATE TABLE post_images (
  post_id bigint NOT NULL,
  image_id bigint REFERENCES images (id) ON DELETE CASCADE,
  created timestamp with time zone DEFAULT now(),
  updated timestamp with time zone DEFAULT now(),
  PRIMARY KEY(post_id, image_id)
);
```

Then you might create the following object representation:
``` go
// post_image.go
type PostImage struct {
    machgo.DefaultCompositeDBObject

    ID int64 `db:"post_id"`
    ImageID int64 `db:"image_id"`
}

func (self *PostImage) GetTableName() string {
    return "post_images"
}

// GetColumnNames is required for composite objects to define what columns
// matter.
func (s *PostImage) GetColumnNames() []string {
    return []string{"post_id", "image_id"}
}

// image_object.go
type Image struct {
    database.DefaultDBObject

    // Defining the db value as 'foreign' here lets the machgo know that it
    // needs to pull the column from the table specified in the dbforeign tag
    // value.
    PostID int64 `db:"post_id,foreign" dbforeign:"post_images"`

    MimeType string `db:"mime_type"`
    OriginalFileName string `db:"file_name"`
}

func (self *Image) GetTableName() string {
    return "images"
}

// This function defines what relationships with other tables or objects
func (self *Image) Relationships() []machgo.Relationship {
    return []machgo.Relationship{
        machgo.Relationship{
            SelfObject: self,
            SelfColumn:   "id",
            TargetObject: &ImagePost{},
            TargetColumn: "image_id",
        },
    }
}
```

All the same actions as above will still work the same but in order to get
images for a post (or set of posts).
``` go
import (
  "github.com//machgo/pool/session"
  . "github.com//machgo/dsl/dot"
)

func findImagesForPost(postIDs []string) {
    image := &Image{}
    postImage := &PostImage{}

    session, err := sess.NewSession()
    if err != nil {
        panic(err)
    }
    qs := session.Query(postImage, image).SelectObject(image).Where(
        Eq(ObjectColumn(postImage, "image_id"), ObjectColumn(image, "id")),
        In(ObjectColumn(postImage, "post_id"), Const(postIDs...)),
    )

    results, err := qs.Results()
    if err != nil {
        panic(err)
    }

    images := make([]*Image, 0)
    err = results.WriteAllTo(&images)
    if err != nil {
      panic(err)
    }

    return images
}
```
