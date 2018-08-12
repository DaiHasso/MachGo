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

func (j *UnexportedDefaultAttributes) created() types.Timestamp {
	return j.Created
}

func (j *UnexportedDefaultAttributes) updated() types.Timestamp {
	return j.Updated
}

func (j *UnexportedDefaultAttributes) update() {
	j.Updated = types.Timestamp{Time: time.Now()}
}

func (j *UnexportedDefaultAttributes) init() {
	j.Created = types.Timestamp{Time: time.Now()}
	j.update()
}
