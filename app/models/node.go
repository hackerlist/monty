package models

type Node struct {
	Id      int64  `db:"id" json:"id"`
	Mid     int64  `db:"mid" json:"mid"`
	Sysname string `db:"sysname" json:"sysname"`
}
