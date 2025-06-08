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
	fmt.Println("Run on", models.ServerAddress)
	err := http.ListenAndServe(models.ServerAddress, router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func flags() {
	flag.StringVar(&models.ServerAddress, "a", models.ServerAddress, "server address ip:port")
	flag.Parse()
}
