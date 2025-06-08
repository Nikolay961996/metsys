package main

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
)

func main() {
	flags()
	fmt.Println("Run on", models.RunOnServerAddress)
	err := http.ListenAndServe(models.RunOnServerAddress, router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func flags() {
	flag.StringVar(&models.RunOnServerAddress, "a", "http://localhost:8080", "server address ip:port")
	flag.Parse()
}
