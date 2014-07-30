package controllers

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

func LuaInit() {
}

// mkstate constructs a new lua state with the appropriate functions binded
// for use in scripts.
// it also binds functions used for actions, which should probably go somewhere else.
func mkstate() *lua.State {
	L := luar.Init()

	luar.Register(L, "", luar.Map{
		"connect": lconnect,
		"http":    lhttp,
	})

	return L
}

// The following functions are executed by way of the scripts table.

// lconnect tries to make a tcp connection to a host.
func lconnect(proto, host, port string) error {
	c, err := net.Dial(proto, host+":"+port)
	if err != nil {
		return err
	}

	defer c.Close()

	return nil
}

// lhttp tries to make a http GET request to the given url.
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

