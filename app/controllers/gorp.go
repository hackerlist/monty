package controllers

import (
	//	"code.google.com/p/go.crypto/bcrypt"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/hackerlist/monty/app/models"
	_ "github.com/lib/pq"
	r "github.com/revel/revel"
	"github.com/revel/revel/modules/db/app"
)

var (
	Dbm *gorp.DbMap
)

func InitDB() {
	db.Init()
	Dbm = &gorp.DbMap{Db: db.Db, Dialect: gorp.PostgresDialect{}}

	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}

	t := Dbm.AddTableWithName(models.Node{}, "node").SetKeys(true, "Id")
	//		t.ColMap("Password").Transient = true
	setColumnSizes(t, map[string]int{
		"Sysname": 100,
	})

	t = Dbm.AddTableWithName(models.Probe{}, "probe").SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Name":      100,
		"Arguments": 255,
	})

	t = Dbm.AddTableWithName(models.Result{}, "result").SetKeys(true, "Id")
	/*
		t.ColMap("User").Transient = true
		t.ColMap("Hotel").Transient = true
		t.ColMap("CheckInDate").Transient = true
		t.ColMap("CheckOutDate").Transient = true
	*/
	setColumnSizes(t, map[string]int{
		"StatusMsg": 255,
	})

	t = Dbm.AddTableWithName(models.Script{}, "script").SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Desc": 100,
		"Code": 255,
	})

	Dbm.TraceOn("[gorp]", r.TRACE)
	Dbm.CreateTablesIfNotExists()

	/*
		bcryptPassword, _ := bcrypt.GenerateFromPassword(
			[]byte("demo"), bcrypt.DefaultCost)
	*/

	/* insert our default data. return early if it exists. */
	demoNode := &models.Node{1, 1, "hackerlist.net"}
	if hl, _ := Dbm.Get(&models.Node{}, demoNode.Id); hl != nil {
		return
	}

	/* defaults aren't present. set them up. */
	if err := Dbm.Insert(demoNode); err != nil {
		panic(err)
	}

	scripts := []*models.Script{
		&models.Script{0, "HTTP GET check", "local b, err = http(args.Url); return err"},
		&models.Script{0, "TCP/UDP port check", "return connect(args.Proto, args.Host, args.Port)"},
	}

	for _, script := range scripts {
		if err := Dbm.Insert(script); err != nil {
			panic(err)
		}
	}

	probe := &models.Probe{0, demoNode.Id, "http", 60.0, time.Now(), 0, 1, `{ "Url": "https://hackerlist.net/" }`}
	if err := Dbm.Insert(probe); err != nil {
		panic(err)
	}

}

type GorpController struct {
	*r.Controller
	Txn *gorp.Transaction
}

// This should be moved.
func (c *GorpController) CheckToken() r.Result {
	var theirtoken string

	ourtoken, ok := r.Config.String("monty.token")

	c.Params.Bind(&theirtoken, "token")

	if ok && ourtoken != theirtoken {
		return c.RenderError(fmt.Errorf("bad api token"))
	}

	return nil
}

func (c *GorpController) Begin() r.Result {
	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

func (c *GorpController) Commit() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
