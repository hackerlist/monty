package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/hackerlist/monty/app/models"
	"github.com/revel/revel"
	"io"
	"time"
)

type Response map[string]interface{}

// Convenience types for loading info from the DB
func getbyid(tx gorp.SqlExecutor, typ interface{}, id int) (interface{}, error) {
	res, err := tx.Get(typ, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type Api struct {
	GorpController
}

func (c Api) Index() revel.Result {
	r := Response{}
	return c.RenderJson(r)
}

func (c Api) NewNode() revel.Result {
	var mid int64
	var sysname string
	var newnode models.Node
	var err error

	r := Response{}

	c.Params.Bind(&mid, "mid")
	c.Params.Bind(&sysname, "sysname")

	if mid == 0 {
		err = fmt.Errorf("bad mid: %d", mid)
		goto out
	}

	if sysname == "" || len(sysname) < 5 {
		err = fmt.Errorf("bad sysname: %s", sysname)
		goto out
	}

	newnode.Mid = mid
	newnode.Sysname = sysname

	c.Txn.Insert(&newnode)

out:
	r["error"] = err
	r["node"] = newnode
	return c.RenderJson(r)
}

func (c Api) ListNodes() revel.Result {
	var nodes []models.Node

	r := Response{}

	_, err := c.Txn.Select(&nodes, "SELECT * FROM node")
	if err != nil {
		r["error"] = err
		goto out
	}

out:
	r["nodes"] = nodes
	return c.RenderJson(r)
}

func (c Api) ShowNode(id int) revel.Result {
	r := Response{}

	node := models.Node{}
	r["node"], r["error"] = getbyid(c.Txn, &node, id)
	return c.RenderJson(r)
}

func (c Api) NodeStatus(mid int) revel.Result {
	r := Response{}

	r["healthy"] = false

	count, err := c.Txn.SelectInt(`SELECT COUNT(*) FROM node n, probe p where n.mid = $1 and n.id = p.nid`, mid)
	if err != nil {
		r["error"] = err
		return c.RenderJson(r)
	}

	r["probes"] = count

	res, err := c.Txn.Select(models.Result{}, `SELECT r.id, r.nid, r.pid, r.time, r.passed, r.statusmsg
												FROM node n, probe p, result r
												WHERE n.mid = $1 and n.id = p.nid and p.rid = r.id
												ORDER BY p.id
												LIMIT $2`, mid, count)

	if err != nil {
		r["error"] = err
		return c.RenderJson(r)
	}

	r["results"] = res

	passing := int64(0)
	for _, re := range res {
		result := re.(*models.Result)
		if result.Passed {
			passing++
		}
	}

	r["passing"] = passing

	if count == passing {
		r["healthy"] = true
	}

	return c.RenderJson(r)
}

func (c Api) NewProbe(nid int, name string, frequency float64, sid int) revel.Result {
	var err interface{}
	var newprobe *models.Probe

	var validjson interface{}
	args := new(bytes.Buffer)

	c.Validation.Required(nid)
	c.Validation.Range(nid, 1, 65535)
	c.Validation.Required(name)
	c.Validation.Required(frequency)
	c.Validation.Required(sid)
	c.Validation.Range(sid, 1, 65535)

	r := Response{}

	if c.Validation.HasErrors() {
		e := c.Validation.Errors[0]
		err = fmt.Sprintf("%s: %s", e.Key, e.Message)
		goto out
	}

	/* get args as post body */
	io.Copy(args, c.Request.Body)

	// validate the json in body
	err = json.Unmarshal(args.Bytes(), &validjson)
	if err != nil {
		goto out
	}

	newprobe = &models.Probe{
		NodeId:       int64(nid),
		Name:         name,
		Frequency:    frequency,
		LastRun:      time.Now().Add(-(int64(frequency) * time.Second)),
		LastResultId: 0,
		ScriptId:     int64(sid),
		Arguments:    args.String(),
	}

	err = c.Txn.Insert(newprobe)

out:
	r["error"] = err
	r["probe"] = newprobe
	return c.RenderJson(r)
}

func (c Api) ListProbes() revel.Result {
	var probes []models.Probe

	r := Response{}

	_, err := c.Txn.Select(&probes, "SELECT * FROM probe")
	if err != nil {
		r["error"] = err
		goto out
	}

out:
	r["probes"] = probes
	return c.RenderJson(r)
}

func (c Api) ShowProbe(id int) revel.Result {
	r := Response{}

	probe := models.Probe{}
	r["probe"], r["error"] = getbyid(c.Txn, &probe, id)
	return c.RenderJson(r)
}

func (c Api) ListResults() revel.Result {
	var results []models.Result

	r := Response{}

	_, err := c.Txn.Select(&results, "SELECT * FROM result ORDER BY id DESC LIMIT 100")
	if err != nil {
		r["error"] = err
		goto out
	}

out:
	r["results"] = results
	return c.RenderJson(r)
}

func (c Api) ShowResult(id int) revel.Result {
	r := Response{}

	result := models.Result{}
	r["result"], r["error"] = getbyid(c.Txn, &result, id)
	return c.RenderJson(r)
}

func (c Api) ListScripts() revel.Result {
	var scripts []models.Script

	r := Response{}

	_, err := c.Txn.Select(&scripts, "SELECT * FROM script")
	if err != nil {
		r["error"] = err
		goto out
	}

out:
	r["scripts"] = scripts
	return c.RenderJson(r)
}

func (c Api) ShowScript(id int) revel.Result {
	r := Response{}

	script := models.Script{}
	r["script"], r["error"] = getbyid(c.Txn, &script, id)
	return c.RenderJson(r)
}
