package server

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	RunOnServerAddress string
}

func DefaultConfig() Config {
	return Config{
		RunOnServerAddress: "http://localhost:8080",
	}
}

func (c *Config) Parse() {
	c.flags()
	c.envs()

	models.Log.Info("Server run on",
		zap.String("address", c.RunOnServerAddress))
}

func (c *Config) flags() {
	flag.StringVar(&c.RunOnServerAddress, "a", "localhost:8080", "server address ip:port")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Printf("Unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}
}

func (c *Config) envs() {
	var configEnv struct {
		Address string `env:"ADDRESS"`
	}
	err := env.Parse(&configEnv)
	if err != nil {
		panic(err)
	}

	if configEnv.Address != "" {
		c.RunOnServerAddress = configEnv.Address
	}
}
