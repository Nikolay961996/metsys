// Entry point for server app.
package main

import (
	"fmt"

	"github.com/Nikolay961996/metsys/internal/server"
	"github.com/Nikolay961996/metsys/models"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printHello()
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := server.DefaultConfig()
	c.Parse()
	entity := server.InitServer(&c)
	entity.Run(c.RunOnServerAddress, c.KeyForSigning)
	defer entity.Stop()
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
