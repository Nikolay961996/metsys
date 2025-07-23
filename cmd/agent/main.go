package main

import (
	"github.com/Nikolay961996/metsys/internal/agent"
)

func main() {
	c := agent.DefaultConfig()
	c.Parse()

	a := agent.InitAgent()
	a.Run(&c)
	defer a.Stop()
}
