package server

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
	"os"
	"time"
)

type Config struct {
	RunOnServerAddress string
	StoreInterval      time.Duration
	FileStoragePath    string
	Restore            bool
}

func DefaultConfig() Config {
	return Config{
		RunOnServerAddress: "localhost:8080",
		StoreInterval:      3000 * time.Second,
		FileStoragePath:    "/metsys.db",
		Restore:            false,
	}
}

func (c *Config) Parse() {
	c.flags()
	c.envs()

	models.Log.Info("Server run on",
		zap.String("address", c.RunOnServerAddress))
}

func (c *Config) flags() {
	flag.StringVar(&c.RunOnServerAddress, "a", c.RunOnServerAddress, "server address ip:port")
	i := flag.Int("i", 300, "period for local saving. 0 - sync save")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "path to file for saves")
	flag.BoolVar(&c.Restore, "r", c.Restore, "restore save on start")

	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Printf("Unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}

	c.StoreInterval = time.Duration(*i) * time.Second
}

func (c *Config) envs() {
	var configEnv struct {
		Address         string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         *bool  `env:"RESTORE"`
	}
	err := env.Parse(&configEnv)
	if err != nil {
		panic(err)
	}

	if configEnv.Address != "" {
		c.RunOnServerAddress = configEnv.Address
	}
	if configEnv.StoreInterval != 0 {
		c.StoreInterval = time.Duration(configEnv.StoreInterval) * time.Second
	}
	if configEnv.FileStoragePath != "" {
		c.FileStoragePath = configEnv.FileStoragePath
	}
	if configEnv.Restore != nil {
		c.Restore = *configEnv.Restore
	}
}
