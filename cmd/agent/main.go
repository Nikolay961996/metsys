// Entry point for agent app.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/models"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	addr = ":8081"
)

func main() {
	printHello()
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

func printHello() {
	fmt.Printf("Build version: %s\n", ifNA(buildVersion))
	fmt.Printf("Build date: %s\n", ifNA(buildDate))
	fmt.Printf("Build commit: %s\n", ifNA(buildCommit))
}

func ifNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
