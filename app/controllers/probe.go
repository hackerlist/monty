package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/revel/revel"
	"github.com/revel/revel/modules/jobs/app/jobs"

	"github.com/hackerlist/monty/app/models"

	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

type ProbeJob struct {
}

func (pjob ProbeJob) Run() {
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
		probes = nil
		_, err := txn.Select(&probes, "select * from probe where nid=$1 order by id", n.Id)
		if err != nil {
			revel.ERROR.Printf("runchecks: %s", err)
			continue
		}

		for _, p := range probes {
			/* make sure we run the test at the right frequency */
			last := p.LastRun
			if t.Sub(last).Seconds() < p.Frequency {
				revel.INFO.Printf("skipping probe %d - timeout not hit", p.Id)
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

			scr := script.(*models.Script)

			revel.TRACE.Printf("node %d probe %d script (%d) %q args %s", n.Id, p.Id, scr.Id, scr.Code, p.Arguments)
			prunner := NewProbeRunner(&p, scr)

			jobs.Now(prunner)

			res := &models.Result{
				NodeId:  n.Id,
				ProbeId: p.Id,
			}

			err = <-prunner.Error
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

			revel.TRACE.Printf("probe passed? %d %t %q", p.Id, res.Passed, res.StatusMsg)

			// if the probe failed, send the result to the callback url in the node.
			if res.Passed == false {
				revel.WARN.Printf("probe failed, running callback")
				go pjob.Failed(&n, &p, res)
			}
		}
	}

	if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
}

func (p ProbeJob) Failed(node *models.Node, probe *models.Probe, result *models.Result) {
	args := make(map[string]interface{})

	args["node"] = node
	args["probe"] = probe
	args["result"] = result

	postbody := new(bytes.Buffer)
	json.NewEncoder(postbody).Encode(args)

	resp, err := http.Post(node.Callback, "application/json", postbody)
	if err != nil {
		revel.ERROR.Printf("failure to POST failed probe to %q: %s", node.Callback, err)
	}
	defer resp.Body.Close()
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
	var state *lua.State
	var fun *luar.LuaObject
	var res interface{}

	r := p.Error

	args := make(map[string]interface{})

	if err = json.Unmarshal([]byte(p.Probe.Arguments), &args); err != nil {
		goto out
	}

	state = mkstate()
	defer state.Close()

	err = state.DoString(fmt.Sprintf("fun = function(args) %s end", p.Script.Code))
	if err != nil {
		goto out
	}

	state.GetGlobal("fun")
	fun = luar.NewLuaObject(state, -1)

	if res, err = fun.Call(args); err != nil {
		goto out
	} else if res == nil {
		// if nil, that means no error. just go to out.
		goto out
	} else if status, ok := res.(string); !ok {
		// if it's not a string, that's bad. luar seems to convert go errors to strings..
		err = fmt.Errorf("script resulted in non-string return value %q", res)
	} else if status != "" {
		// if the string is not empty that's an error.
		err = fmt.Errorf("probe error: %s", status)
	}

out:
	r <- err
	close(r)
}
