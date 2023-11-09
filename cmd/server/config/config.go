package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type ServerCfg struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DatabaseDsn   string        `env:"DATABASE_DSN"`
}

func (c *ServerCfg) Get() *ServerCfg {
	flag.StringVar(&c.Address, "a", "localhost:8080", "server address")
	flag.BoolVar(&c.Restore, "r", true, "restore metrics from file")
	flag.DurationVar(&c.StoreInterval, "i", 300*time.Second, "save metrics interval")
	flag.StringVar(&c.StoreFile, "f", "tmp/devops-metrics-db.json", "file for saving metrics")
	flag.StringVar(&c.Key, "k", "", "signature key")
	flag.StringVar(&c.DatabaseDsn, "d", "", "database dsn")
	// адрес бд -- postgres://postgres:123@localhost:5432/postgres?sslmode=disable
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		log.Panic().Err(err).Msg("config parsing error")
	}
	return c
}

func (c *ServerCfg) GetStoreFile() string {
	return c.StoreFile
}

func (c *ServerCfg) GetStoreInterval() time.Duration {
	return c.StoreInterval
}

func (c *ServerCfg) GetKey() string {
	return c.Key
}
