package MachGo

import (
	"time"

	"github.com/DaiHasso/MachGo/types"
)

// UnexportedDefaultAttributes is a set of sensible default attributes for an
// object that are not exported when serialized.
type UnexportedDefaultAttributes struct {
	Created types.Timestamp `db:"created" json:"-"`
	Updated types.Timestamp `db:"updated" json:"-"`
}

// Update will update the updated time.
func (j *UnexportedDefaultAttributes) Update() {
	j.Updated = types.Timestamp{Time: time.Now()}
}

// Init will initialize the created time and updated time.
func (j *UnexportedDefaultAttributes) Init() {
	j.Created = types.Timestamp{Time: time.Now()}
	j.Update()
}
