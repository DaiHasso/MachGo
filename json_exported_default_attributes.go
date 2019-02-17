package MachGo

import (
	"time"

	"github.com/daihasso/machgo/types"
)

// JSONExportedDefaultAttributes is a set of sensible default attributes that
// are serialized into json.
type JSONExportedDefaultAttributes struct {
	Created types.Timestamp `db:"created" json:"created,omitempty"`
	Updated types.Timestamp `db:"updated" json:"updated,omitempty"`
}

// Update will initialize the updated time.
func (j *JSONExportedDefaultAttributes) Update() {
	j.Updated = types.Timestamp{Time: time.Now()}
}

// Init will initialize the created & updated time.
func (j *JSONExportedDefaultAttributes) Init() {
	j.Created = types.Timestamp{Time: time.Now()}
	j.Update()
}
