package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hackerlist/monty/app/models"
	"github.com/revel/revel"
	"github.com/revel/revel/modules/jobs/app/jobs"
	"time"

	/* lua stuff needs this */
	"bytes"
	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
	"io"
	"net"
	"net/http"
)

type ProbeJob struct {
}

func (p ProbeJob) Run() {
	var nodes []models.Node
	var probes []models.Probe

	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}

	t := time.Now()

	_, err = txn.Select(&nodes, "select * from node order by id")
	if err != nil {
		revel.ERROR.Printf("runchecks: %s", err)
		return
	}

	for _, n := range nodes {
		revel.TRACE.Printf("node %d", n.Id)
		_, err := txn.Select(&probes, "select * from probe where nid=$1 order by id", n.Id)
		if err != nil {
			revel.ERROR.Printf("runchecks: %s", err)
			continue
		}
		for _, p := range probes {
			/* make sure we run the test at the right frequency */
			last := p.LastRun
			if t.Sub(last).Seconds() < p.Frequency {
				/* not enough time elapsed - skip */
				continue
			}

			/* get the script for this probe */
			script, err := txn.Get(&models.Script{}, p.ScriptId)
			if err != nil {
				/* if we can't get the script, the probe can't be run */
				revel.ERROR.Printf("can't run probe: missing script %d", p.ScriptId)
				continue
			}

			if script == nil {
				revel.ERROR.Printf("can't run probe: missing script %d", p.ScriptId)
				continue
			}

			/* set up and run the probe via revel jobs api.
			 * the result comes back through the Error channel. */
			pj := NewProbeRunner(&p, script.(*models.Script))

			jobs.Now(pj)

			res := &models.Result{
				NodeId:  n.Id,
				ProbeId: p.Id,
			}

			err = <-pj.Error
			if err != nil {
				res.Passed = false
				res.StatusMsg = err.Error()
			} else {
				res.Passed = true
				res.StatusMsg = ""
			}
			res.Timestamp = time.Now()
			err = txn.Insert(res)
			if err != nil {
				revel.ERROR.Printf("runchecks: %s", err)
				continue
			}
			p.LastResultId = res.Id
			_, err = txn.Update(&p)
			if err != nil {
				revel.ERROR.Printf("runchecks: %s", err)
				continue
			}
			revel.TRACE.Printf("probe passed? %t %d", res.Passed, p.Id)
		}
	}

	if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}

type ProbeRunner struct {
	Probe  *models.Probe
	Script *models.Script

	Error chan error
}

func NewProbeRunner(p *models.Probe, s *models.Script) *ProbeRunner {
	return &ProbeRunner{
		Probe:  p,
		Script: s,
		Error:  make(chan error),
	}
}

func (p ProbeRunner) Run() {
	var err error
	var fun *luar.LuaObject
	r := p.Error

	args := make(map[string]interface{})

	json.Unmarshal([]byte(p.Probe.Arguments), args)

	l := mkstate()
	defer l.Close()

	err = l.DoString(fmt.Sprintf("fun = function(args) %s end", p.Script.Code))
	if err != nil {
		goto out
	}

	l.GetGlobal("fun")
	fun = luar.NewLuaObject(l, -1)
	if res, err := fun.Call(args); err != nil {
		goto out
	} else if status, ok := res.(string); ok && status != "" {
		err = fmt.Errorf("%s", status)
		goto out
	}

out:
	r <- err
	close(r)
}

/* lua api stuff */
func lconnect(proto, host, port string) error {
	c, err := net.Dial(proto, host+":"+port)
	if err != nil {
		return err
	}

	defer c.Close()

	return nil
}

func lhttp(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)

	return b.String(), nil
}

func mkstate() *lua.State {
	L := luar.Init()

	luar.Register(L, "", luar.Map{
		"connect": lconnect,
		"http":    lhttp,
	})

	return L
}
