package models

import (
	"time"
)

type Probe struct {
	Id           int64     `db:"id" json:"id"`
	NodeId       int64     `db:"nid" json:"nid"`
	Name         string    `db:"name" json:"name"`
	Frequency    float64   `db:"frequency" json:"frequency"` // in minutes
	LastRun      time.Time `db:"last" json:"last"`
	LastResultId int64     `db:"rid" json:"rid"`
	ScriptId     int64     `db:"sid" json:"sid"`
	Arguments    string    `db:"args" json:"args"`
}
