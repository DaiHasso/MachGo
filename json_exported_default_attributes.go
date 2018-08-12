package MachGo

import (
	"time"

	"github.com/DaiHasso/MachGo/types"
)

type jsonExportedDefaultAttributes struct {
	Created types.Timestamp `db:"created" json:"created,omitempty"`
	Updated types.Timestamp `db:"updated" json:"updated,omitempty"`
}

func (j *jsonExportedDefaultAttributes) created() types.Timestamp {
	return j.Created
}

func (j *jsonExportedDefaultAttributes) updated() types.Timestamp {
	return j.Updated
}

func (j *jsonExportedDefaultAttributes) update() {
	j.Updated = types.Timestamp{Time: time.Now()}
}

func (j *jsonExportedDefaultAttributes) init() {
	j.Created = types.Timestamp{Time: time.Now()}
	j.update()
}
