package main

import (
	"flag"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
)

func main() {
	flags()
	err := http.ListenAndServe(":8080", router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func flags() {
	flag.StringVar(&models.ServerAddress, "a", models.ServerAddress, "server address ip:port")
	flag.Parse()
}
