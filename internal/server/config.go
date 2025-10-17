// Package server consist server main config entities
package server

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Nikolay961996/metsys/utils"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"

	"github.com/Nikolay961996/metsys/models"
)

type Config struct {
	RunOnServerAddress string        // server address
	FileStoragePath    string        // file storage path
	DatabaseDSN        string        // database connection string
	KeyForSigning      string        // key for sign
	CryptoKey          string        // key for decrypt (private key of server)
	StoreInterval      time.Duration // interval for stor
	Restore            bool          // need restore
}

func DefaultConfig() Config {
	return Config{
		RunOnServerAddress: "localhost:8080",
		StoreInterval:      300 * time.Second,
		FileStoragePath:    "",
		Restore:            false,
		DatabaseDSN:        "",
		KeyForSigning:      "",
		CryptoKey:          "",
	}
}

func (c *Config) Parse() {
	c.flags()
	c.envs()

	if !utils.FileExists(c.CryptoKey) {
		panic(errors.New("CryptoKey file not found"))
	}

	models.Log.Info("Server run on",
		zap.String("address", c.RunOnServerAddress))
}

func (c *Config) flags() {
	flag.StringVar(&c.RunOnServerAddress, "a", c.RunOnServerAddress, "server address ip:port")
	i := flag.Int("i", 300, "period for local saving. 0 - sync save")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "path to file for saves")
	flag.BoolVar(&c.Restore, "r", c.Restore, "restore save on start")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "database connection string")
	flag.StringVar(&c.KeyForSigning, "k", c.KeyForSigning, "key for signing")
	flag.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "Key for decryption")

	flag.Parse()

	if flag.NArg() > 0 {
		models.Log.Error(fmt.Sprintf("Unknown flags: %v\n", flag.Args()))
		os.Exit(1)
	}

	c.StoreInterval = time.Duration(*i) * time.Second
}

func (c *Config) envs() {
	var configEnv struct {
		Restore         *bool  `env:"RESTORE"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
		Address         string `env:"ADDRESS"`
		KeyForSigning   string `env:"KEY"`
		CryptoKey       string `env:"CRYPTO_KEY"`
		StoreInterval   int32  `env:"STORE_INTERVAL"`
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
	if configEnv.DatabaseDSN != "" {
		c.DatabaseDSN = configEnv.DatabaseDSN
	}
	if configEnv.KeyForSigning != "" {
		c.KeyForSigning = configEnv.KeyForSigning
	}
	if configEnv.CryptoKey != "" {
		c.CryptoKey = configEnv.CryptoKey
	}
}
