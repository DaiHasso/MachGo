# MachGo
[![Build Status](https://travis-ci.com/DaiHasso/MachGo.svg?branch=master)](https://travis-ci.com/DaiHasso/MachGo)[![Coverage Status](https://coveralls.io/repos/github/DaiHasso/MachGo/badge.svg?branch=master)](https://coveralls.io/github/DaiHasso/MachGo?branch=master)

MachGo is a ORM-ish database library for go.

Because I am chronically lazy, this readme is a WIP.

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
	database.DefaultDBObject

	PostID int64 `db:"post_id"`

	MimeType string `db:"mime_type"`
	OriginalFileName string `db:"file_name"`
}

func (self *Image) GetTableName() string {
	return "images"
}
```

The use of `DefaultDBObject` from the database pacakge adds some default
behaviour as well as auto generating and managing an ID property for the object.
The auto-generated/managed ID field can be accessed like `objectInstance.ID`.

Next you might want to create a database connection:
``` go
import (
    "github.com/daihasso/machgo/database"
)

var MyDB *database.manager

func init() {
    MyDB, err := database.GetDatabaseManager(
        database.Postgres, "mypostgresuser", "mysecretpassword", "localhost",
        5432, "mycooldatabase",
    )
    if err != nil {
        panic(err)
    }
}
```

And to retrieve single image you might do:
``` go

func getImage(id int64) *Image {
    result := &Image{}
    err := MyDB.GetObject(result, id)
    if err != nil {
        panic(err)
    }

    return result
}
```

And to retrieve all images you might do:
``` go

func getImages() *[]*Image {
    result := &Image{}
    err := MyDb.GetObject(result, id)
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
    MachGo.DefaultCompositeDBObject

    ID *uuid.UUID `db:"post_id"`
    ImageID *uuid.UUID `db:"image_id"`
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

    // Defining the db value as 'foreign' here lets the MachGo know that it
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
func (self *Image) Relationships() []MachGo.Relationship {
    return []MachGo.Relationship{
        MachGo.Relationship{
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
	"github.com/DaiHasso/MachGo/session"
	. "github.com/DaiHasso/MachGo/dsl/dot"
)

func findImagesForPost(postIDs []string) {
    image := &Image{}
    postImage := &PostImage{}

    sess := session.New()
    qs := sess.Query(postImage, image).SelectObject(image).Where(
        Where(
            Eq(ObjectColumn(postImage, "image_id"), ObjectColumn(image, "id")),
            In(ObjectColumn(postImage, "post_id"), Const(postIDs...)),
        ),
    )
}
```

Check out the [godocs!](https://godoc.org/github.com/DaiHasso/MachGo)
