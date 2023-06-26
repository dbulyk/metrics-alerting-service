package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
)

type Agent struct {
	Address        string        `env:"ADDRESS" envDescription:"адрес сервера"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDescription:"интервал отправки метрик на сервер"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDescription:"интервал опроса метрик"`
}

var agent Agent

func NewAgentCfg() (*Agent, error) {
	flag.StringVar(&agent.Address, "a", "localhost:8080", "адрес сервера")
	flag.DurationVar(&agent.ReportInterval, "r", 10*time.Second, "интервал отправки метрик на сервер")
	flag.DurationVar(&agent.PollInterval, "p", 2*time.Second, "интервал опроса метрик")
	flag.Parse()

	err := env.Parse(&agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}
