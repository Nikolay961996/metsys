// Entry point for agent app.
package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/internal/buildinfo"
	"github.com/Nikolay961996/metsys/models"
)

const (
	addr = ":8081"
)

func main() {
	buildinfo.PrintHello()
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := agent.DefaultConfig()
	c.Parse()

	a := agent.InitAgent()
	go a.Run(&c)

	defer a.Stop()

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
