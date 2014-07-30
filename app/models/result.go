package models

import (
	"time"
)

type Result struct {
	Id        int64     `db:"id" json:"id"`
	NodeId    int64     `db:"nid" json:"nid"`
	ProbeId   int64     `db:"pid" json:"pid"`
	Timestamp time.Time `db:"time" json:"time"`
	Passed    bool      `db:"passed" json:"passed"`
	StatusMsg string    `db:"statusmsg" json:"statusmsg"`
}
