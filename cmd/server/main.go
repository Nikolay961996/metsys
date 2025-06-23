package main

import (
	"github.com/Nikolay961996/metsys/internal/server"
	"github.com/Nikolay961996/metsys/models"
)

func main() {
	err := models.Initialize("info")
	if err != nil {
		panic(err)
	}

	c := server.DefaultConfig()
	c.Parse()
	entity := server.InitServer(&c)
	entity.Run()
	defer entity.Stop()
}
