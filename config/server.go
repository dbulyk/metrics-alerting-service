package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
}

var server Server

func NewServerCfg() (*Server, error) {
	flag.StringVar(&server.Address, "a", "localhost:8080", "адрес сервера")
	flag.BoolVar(&server.Restore, "r", true, "восстановить метрики из файла")
	flag.DurationVar(&server.StoreInterval, "i", 20*time.Second, "интервал сохранения метрик в файл")
	flag.StringVar(&server.StoreFile, "f", "tmp/devops-metrics-db.json", "файл для сохранения метрик")
	flag.StringVar(&server.Key, "k", "", "ключ подписи")
	flag.Parse()

	err := env.Parse(&server)
	if err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка парсинга конфига")
		return nil, err
	}
	return &server, nil
}

func GetStoreFile() string {
	return server.StoreFile
}

func GetStoreInterval() time.Duration {
	return server.StoreInterval
}

func GetKey() string {
	return server.Key
}
