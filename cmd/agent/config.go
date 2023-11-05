package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentCfg struct {
	Address        string        `env:"ADDRESS" envDescription:"server address"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDescription:"interval for sending metrics to the server"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDescription:"interval for polling metrics"`
	Key            string        `env:"KEY" envDescription:"signature key"`
	RateLimit      int           `env:"RATE_LIMIT" envDescription:"rate limit for requests to the server"`
}

func Get() (*AgentCfg, error) {
	agentCfg := AgentCfg{}
	flag.StringVar(&agentCfg.Address, "a", "localhost:8080", "server address")
	flag.DurationVar(&agentCfg.ReportInterval, "r", 10*time.Second, "interval for sending metrics to the server")
	flag.DurationVar(&agentCfg.PollInterval, "p", 2*time.Second, "interval for polling metrics")
	flag.StringVar(&agentCfg.Key, "k", "", "signature key")
	flag.IntVar(&agentCfg.RateLimit, "l", 3, "rate limit for requests to the server")
	flag.Parse()

	err := env.Parse(&agentCfg)
	if err != nil {
		return nil, err
	}

	return &agentCfg, nil
}
