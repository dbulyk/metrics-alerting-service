package main

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	cfg config
	mem *stores.MemStorage
)

type config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	mem = stores.NewMemStorage()
}

func main() {
	parseFlagsAndEnvs()

	isAsync := len(cfg.StoreFile) != 0 && cfg.StoreInterval == 0
	r, err := handlers.MetricsRouter(mem, isAsync, cfg.StoreFile)
	if err != nil {
		log.Fatal().Timestamp().Err(err).Msg("ошибка инициализации роутера")
		os.Exit(1)
	}
	log.Info().Timestamp().Msg("роутер инициализирован")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if cfg.Restore {
		err := utils.RestoreMetrics(mem, cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка восстановления метрик")
		}
	}

	if len(cfg.StoreFile) != 0 && cfg.StoreInterval != 0 {
		log.Info().Msgf("запуск записи метрик в файл с интервалом в %s секунд", cfg.StoreInterval)
		writeTicker := time.NewTicker(cfg.StoreInterval)
		go func() {
			for range writeTicker.C {
				err := utils.SaveMetrics(mem, cfg.StoreFile)
				if err != nil {
					log.Error().Timestamp().Err(err).Msg("ошибка сохранения метрик")
				}
			}
		}()
		defer writeTicker.Stop()
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}
	log.Info().Timestamp().Msgf("сервер запускается на %s", cfg.Address)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка работы сервера")
		}
	}()

	<-sigs
	log.Info().Timestamp().Msg("получен сигнал остановки")

	if len(cfg.StoreFile) != 0 {
		err = utils.SaveMetrics(mem, cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка сохранения метрик")
		}
	}

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка остановки сервера")
	}

	log.Info().Timestamp().Msg("сервер остановлен")
	os.Exit(0)
}

func parseFlagsAndEnvs() {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "адрес сервера")
	flag.BoolVar(&cfg.Restore, "r", true, "восстановить метрики из файла")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "интервал сохранения метрик в файл")
	flag.StringVar(&cfg.StoreFile, "f", "tmp/devops-metrics-db.json", "файл для сохранения метрик")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка парсинга конфига")
	}
}
