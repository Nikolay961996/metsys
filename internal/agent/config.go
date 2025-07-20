package agent

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"os"
	"strings"
	"time"
)

type Config struct {
	SendToServerAddress string
	PollInterval        time.Duration
	ReportInterval      time.Duration
	KeyForSigning       string
}

func DefaultConfig() Config {
	return Config{
		SendToServerAddress: "http://localhost:8080",
		PollInterval:        2 * time.Second,
		ReportInterval:      10 * time.Second,
		KeyForSigning:       "",
	}
}

func (c *Config) Parse() {
	c.flags()
	c.envs()
	fmt.Println("Send to", c.SendToServerAddress)
}

func (c *Config) flags() {
	flag.StringVar(&c.SendToServerAddress, "a", "http://localhost:8080", "Metsys server address ip:port")
	r := flag.Int("r", 10, "ReportInterval in seconds")
	p := flag.Int("p", 2, "PollInterval in seconds")
	k := flag.String("k", "", "Key for signing")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Printf("Unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}

	c.SendToServerAddress = c.fixProtocolPrefixAddress(c.SendToServerAddress)
	c.ReportInterval = time.Duration(*r) * time.Second
	c.PollInterval = time.Duration(*p) * time.Second
	c.KeyForSigning = *k
}

func (c *Config) envs() {
	var configEnv struct {
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
		KeyForSigning  string `env:"KEY"`
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
}

func (c *Config) fixProtocolPrefixAddress(addr string) string {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")

	return addr
}
