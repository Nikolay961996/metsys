package main

import (
	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
	_ "net/http/pprof"
)

const (
	addr = ":8081"
)

func main() {
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := agent.DefaultConfig()
	c.Parse()

	a := agent.InitAgent()
	go a.Run(&c)

	defer a.Stop()

	_ = http.ListenAndServe(addr, nil)
}
