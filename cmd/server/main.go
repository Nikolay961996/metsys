package main

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/models"
	"github.com/caarlos0/env/v6"
	"net/http"
	"os"
)

func main() {
	flags()
	envs()

	fmt.Println("Run on", models.RunOnServerAddress)
	err := http.ListenAndServe(models.RunOnServerAddress, router.MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func flags() {
	flag.StringVar(&models.RunOnServerAddress, "a", "localhost:8080", "server address ip:port")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Printf("Unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}
}

func envs() {
	var c struct {
		Address string `env:"ADDRESS"`
	}
	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	if c.Address != "" {
		models.RunOnServerAddress = c.Address
	}
}
