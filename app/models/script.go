package models

type Script struct {
	Id   int64  `db:"id" json:"id"`
	Desc string `db:"desc" json:"desc"`
	Code string `db:"code" json:"code"`
}
