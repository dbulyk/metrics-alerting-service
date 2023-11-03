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
	DatabaseDsn   string        `env:"DATABASE_DSN"`
}

var server Server

func GetServerCfg() Server {
	flag.StringVar(&server.Address, "a", "localhost:8080", "адрес сервера")
	flag.BoolVar(&server.Restore, "r", true, "восстановить метрики из файла")
	flag.DurationVar(&server.StoreInterval, "i", 300*time.Second, "интервал сохранения метрик в файл")
	flag.StringVar(&server.StoreFile, "f", "tmp/devops-metrics-db.json", "файл для сохранения метрик")
	flag.StringVar(&server.Key, "k", "", "ключ подписи")
	flag.StringVar(&server.DatabaseDsn, "d", "", "строка подключения к базе данных")
	// адрес бд -- postgres://postgres:123@localhost:5432/postgres?sslmode=disable
	flag.Parse()

	err := env.Parse(&server)
	if err != nil {
		log.Panic().Err(err).Msg("ошибка чтения конфига")
	}
	return server
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
