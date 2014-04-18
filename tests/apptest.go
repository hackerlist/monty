package tests

import "github.com/revel/revel"

type AppTest struct {
	revel.TestSuite
}

func (t *AppTest) Before() {
	println("Set up")
}

func (t AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t AppTest) TestListNodes() {
	t.Get("/api/nodes")
	t.AssertOk()
}

func (t AppTest) TestShowNode() {
	t.Get("/api/nodes/1")
	t.AssertOk()
}

func (t AppTest) TestNodeStatus() {
	t.Get("/api/status/1")
	t.AssertOk()
}

func (t AppTest) TestListScripts() {
	t.Get("/api/scripts")
	t.AssertOk()
}

func (t AppTest) TestShowScript() {
	t.Get("/api/script/1")
	t.AssertOk()
}

func (t *AppTest) After() {
	println("Tear down")
}
