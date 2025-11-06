// Package server consist server main config entities
package server

import (
	"encoding/json"
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
	RunOnServerAddress string        `json:"address"`      // server address
	GRPCPort           string        `json:"grpc_port"`    // gRPC server port
	FileStoragePath    string        `json:"store_file"`   // file storage path
	DatabaseDSN        string        `json:"database_dsn"` // database connection string
	KeyForSigning      string        // key for sign
	CryptoKey          string        `json:"crypto_key"` // key for decrypt (private key of server)
	ConfigFile         string        // json config
	StoreIntervalStr   string        `json:"store_interval"` // interval for stor
	TrustedSubnet      string        `json:"trusted_subnet"` // trusted subnet in CIDR format
	StoreInterval      time.Duration // interval for stor
	Restore            bool          `json:"restore"` // need restore
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
		ConfigFile:         "",
	}
}

func (c *Config) Parse() {
	c.flags()
	c.envs()
	c.jsonConfig()

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
	flag.StringVar(&c.CryptoKey, "crypto-key", c.CryptoKey, "key for decryption")
	flag.StringVar(&c.ConfigFile, "c", c.ConfigFile, "json config")
	flag.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "trusted subnet in CIDR format")
	flag.StringVar(&c.GRPCPort, "grpc-port", c.GRPCPort, "gRPC server port")

	flag.Parse()

	if flag.NArg() > 0 {
		models.Log.Error(fmt.Sprintf("Unknown flags: %v\n", flag.Args()))
		os.Exit(1)
	}

	c.StoreInterval = time.Duration(*i) * time.Second
}

func (c *Config) envs() {
	var configEnv struct {
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
		Address         string `env:"ADDRESS"`
		KeyForSigning   string `env:"KEY"`
		CryptoKey       string `env:"CRYPTO_KEY"`
		ConfigFile      string `env:"CONFIG"`
		TrustedSubnet   string `env:"TRUSTED_SUBNET"`
		GRPCPort        string `env:"GRPC_PORT"`
		Restore         *bool  `env:"RESTORE"`
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
	if configEnv.ConfigFile != "" {
		c.ConfigFile = configEnv.ConfigFile
	}
	if configEnv.TrustedSubnet != "" {
		c.TrustedSubnet = configEnv.TrustedSubnet
	}
	if configEnv.GRPCPort != "" {
		c.GRPCPort = configEnv.GRPCPort
	}
}

func (c *Config) jsonConfig() {
	if c.ConfigFile == "" {
		return
	}

	d, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		models.Log.Error(fmt.Sprintf("read config file error: %v", err))
		return
	}

	var parsed Config
	err = json.Unmarshal(d, &parsed)
	if err != nil {
		models.Log.Error(fmt.Sprintf("parse config file error: %v", err))
		return
	}

	defConfig := DefaultConfig()
	if c.RunOnServerAddress == defConfig.RunOnServerAddress {
		c.RunOnServerAddress = parsed.RunOnServerAddress
	}
	if c.FileStoragePath == defConfig.FileStoragePath {
		c.FileStoragePath = parsed.FileStoragePath
	}
	if c.DatabaseDSN == defConfig.DatabaseDSN {
		c.DatabaseDSN = parsed.DatabaseDSN
	}
	if c.CryptoKey == defConfig.CryptoKey {
		c.CryptoKey = parsed.CryptoKey
	}
	if c.StoreInterval == defConfig.StoreInterval && parsed.StoreIntervalStr != "" {
		utils.TryParseDuration(&c.StoreInterval, parsed.StoreIntervalStr)
	}
	if c.Restore == defConfig.Restore {
		c.Restore = parsed.Restore
	}
	if c.GRPCPort == "" {
		c.GRPCPort = parsed.GRPCPort
	}
}
