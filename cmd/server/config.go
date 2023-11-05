package main

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
	flag.StringVar(&c.Address, "a", "localhost:8080", "адрес сервера")
	flag.BoolVar(&c.Restore, "r", true, "восстановить метрики из файла")
	flag.DurationVar(&c.StoreInterval, "i", 300*time.Second, "интервал сохранения метрик в файл")
	flag.StringVar(&c.StoreFile, "f", "tmp/devops-metrics-db.json", "файл для сохранения метрик")
	flag.StringVar(&c.Key, "k", "", "ключ подписи")
	flag.StringVar(&c.DatabaseDsn, "d", "", "строка подключения к базе данных")
	// адрес бд -- postgres://postgres:123@localhost:5432/postgres?sslmode=disable
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		log.Panic().Err(err).Msg("ошибка чтения конфига")
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
