package main

import (
	"context"
	"github.com/dbulyk/metrics-alerting-service/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	mem *stores.MemStorage
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	mem = stores.NewMemStorage()
}

func main() {
	cfg, err := config.NewServerCfg()
	if err != nil {
		log.Fatal().Timestamp().Err(err).Msg("ошибка чтения конфига")
	}

	isAsync := len(cfg.StoreFile) != 0 && cfg.StoreInterval == 0
	r, err := handlers.MetricsRouter(mem, isAsync, cfg.StoreFile)
	if err != nil {
		log.Fatal().Timestamp().Err(err).Msg("ошибка инициализации роутера")
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
