// Package agent Configuration agent
package agent

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nikolay961996/metsys/utils"
	"github.com/caarlos0/env/v6"

	"github.com/Nikolay961996/metsys/models"
)

// Config for agent
type Config struct {
	SendToServerAddress  string        `json:"address"` // server address for reporting
	GRPCServerAddress    string        `json:"grpc_address"` // gRPC server address for reporting
	CryptoKey            string        `json:"crypto_key"` // key for encrypt (public key of server)
	ConfigFile           string        // json config
	KeyForSigning        string        // private key for signing
	ReportIntervalStr    string        `json:"report_interval"`
	PollIntervalStr      string        `json:"poll_interval"`
	PollInterval         time.Duration // poll time period
	ReportInterval       time.Duration // report time period
	SendMetricsRateLimit int           // send metrics rate limit
}

// DefaultConfig default config
func DefaultConfig() Config {
	return Config{
		PollInterval:         2 * time.Second,         // updating device data interval
		ReportInterval:       10 * time.Second,        // report to server interval
		KeyForSigning:        "",                      // private key for singing
		SendMetricsRateLimit: 1,                       // rate limit for parallel sending to server
		CryptoKey:            "",                      // key for encrypt (public key of server)
		ConfigFile:           "",                      // json config
	}
}

// Parse from all sources
func (c *Config) Parse() {
	c.flags()
	c.envs()
	c.jsonConfig()

	models.Log.Info(fmt.Sprintf("Send to %s", c.SendToServerAddress))
	if !utils.FileExists(c.CryptoKey) {
		panic(errors.New("CryptoKey file not found"))
	}
}

func (c *Config) flags() {
	flag.StringVar(&c.SendToServerAddress, "a", "http://localhost:8080", "metsys server address ip:port")
	r := flag.Int("r", 10, "reportInterval in seconds")
	p := flag.Int("p", 2, "pollInterval in seconds")
	flag.StringVar(&c.KeyForSigning, "k", "", "key for signing")
	flag.IntVar(&c.SendMetricsRateLimit, "l", 1, "rate limit to sending server")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "key for encryption")
	flag.StringVar(&c.ConfigFile, "c", c.ConfigFile, "json config")
	flag.StringVar(&c.GRPCServerAddress, "grpc-address", "", "gRPC server address")

	flag.Parse()

	if flag.NArg() > 0 {
		models.Log.Error(fmt.Sprintf("Unknown flags: %v\n", flag.Args()))
		os.Exit(1)
	}

	c.SendToServerAddress = c.fixProtocolPrefixAddress(c.SendToServerAddress)
	c.ReportInterval = time.Duration(*r) * time.Second
	c.PollInterval = time.Duration(*p) * time.Second
}

func (c *Config) envs() {
	var configEnv struct {
		Address        string `env:"ADDRESS"`
		KeyForSigning  string `env:"KEY"`
		CryptoKey      string `env:"CRYPTO_KEY"`
		ConfigFile     string `env:"CONFIG"`
		GRPCServerAddress string `env:"GRPC_ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
		SendRateLimit  int    `env:"RATE_LIMIT"`
	}
	err := env.Parse(&configEnv)
	if err != nil {
		panic(err)
	}

	if configEnv.Address != "" {
		c.SendToServerAddress = c.fixProtocolPrefixAddress(configEnv.Address)
	}
	if configEnv.ReportInterval != 0 {
		c.ReportInterval = time.Duration(configEnv.ReportInterval) * time.Second
	}
	if configEnv.PollInterval != 0 {
		c.PollInterval = time.Duration(configEnv.PollInterval) * time.Second
	}
	if configEnv.KeyForSigning != "" {
		c.KeyForSigning = configEnv.KeyForSigning
	}
	if configEnv.SendRateLimit != 0 {
		c.SendMetricsRateLimit = configEnv.SendRateLimit
	}
	if configEnv.CryptoKey != "" {
		c.CryptoKey = configEnv.CryptoKey
	}
	if configEnv.ConfigFile != "" {
		c.ConfigFile = configEnv.ConfigFile
	}
}

func (c *Config) fixProtocolPrefixAddress(addr string) string {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")

	return addr
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
	if c.SendToServerAddress == defConfig.SendToServerAddress {
		c.SendToServerAddress = parsed.SendToServerAddress
	}
	if c.CryptoKey == defConfig.CryptoKey {
		c.CryptoKey = parsed.CryptoKey
	}
	if c.CryptoKey == defConfig.CryptoKey {
		c.CryptoKey = parsed.CryptoKey
	}
	if c.ReportInterval == defConfig.ReportInterval && parsed.ReportIntervalStr != "" {
		utils.TryParseDuration(&c.ReportInterval, parsed.ReportIntervalStr)
	}
	if c.PollInterval == defConfig.PollInterval && parsed.PollIntervalStr != "" {
		utils.TryParseDuration(&c.PollInterval, parsed.PollIntervalStr)
	}
	if c.GRPCServerAddress == "" {
		c.GRPCServerAddress = parsed.GRPCServerAddress
	}
}
