package models

import (
	"time"
)

type Result struct {
	Id        int64     `db:"id"`
	NodeId    int64     `db:"nid"`
	ProbeId   int64     `db:"pid"`
	Timestamp time.Time `db:"time"`
	Passed    bool      `db:"passed"`
	StatusMsg string    `db:"statusmsg"`
}
