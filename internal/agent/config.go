// Package agent Configuration agent
package agent

import (
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
	SendToServerAddress  string        // server address for reporting
	KeyForSigning        string        // private key for signing
	CryptoKey            string        // key for encrypt (public key of server)
	PollInterval         time.Duration // poll time period
	ReportInterval       time.Duration // report time period
	SendMetricsRateLimit int           // send metrics rate limit
}

// DefaultConfig default config
func DefaultConfig() Config {
	return Config{
		SendToServerAddress:  "http://localhost:8080", // server address. Like http://localhost:8080
		PollInterval:         2 * time.Second,         // updating device data interval
		ReportInterval:       10 * time.Second,        // report to server interval
		KeyForSigning:        "",                      // private key for singing
		SendMetricsRateLimit: 1,                       // rate limit for parallel sending to server
		CryptoKey:            "",                      // key for encrypt (public key of server)
	}
}

// Parse from all sources
func (c *Config) Parse() {
	c.flags()
	c.envs()
	models.Log.Info(fmt.Sprintf("Send to %s", c.SendToServerAddress))
	if !utils.FileExists(c.CryptoKey) {
		panic(errors.New("CryptoKey file not found"))
	}
}

func (c *Config) flags() {
	flag.StringVar(&c.SendToServerAddress, "a", "http://localhost:8080", "Metsys server address ip:port")
	r := flag.Int("r", 10, "ReportInterval in seconds")
	p := flag.Int("p", 2, "PollInterval in seconds")
	k := flag.String("k", "", "Key for signing")
	l := flag.Int("l", 1, "Rate limit to sending server")
	cryptoKey := flag.String("crypto-key", "", "Key for encryption")
	flag.Parse()

	if flag.NArg() > 0 {
		models.Log.Error(fmt.Sprintf("Unknown flags: %v\n", flag.Args()))
		os.Exit(1)
	}

	c.SendToServerAddress = c.fixProtocolPrefixAddress(c.SendToServerAddress)
	c.ReportInterval = time.Duration(*r) * time.Second
	c.PollInterval = time.Duration(*p) * time.Second
	c.KeyForSigning = *k
	c.SendMetricsRateLimit = *l
	c.CryptoKey = *cryptoKey
}

func (c *Config) envs() {
	var configEnv struct {
		Address        string `env:"ADDRESS"`
		KeyForSigning  string `env:"KEY"`
		CryptoKey      string `env:"CRYPTO_KEY"`
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
}

func (c *Config) fixProtocolPrefixAddress(addr string) string {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")

	return addr
}
