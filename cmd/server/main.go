package main

import (
	"context"
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
)

type config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

var (
	cfg config
	mem *stores.MemStorage
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	err := env.Parse(&cfg)
	if err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка парсинга конфига")
	}

	mem = stores.NewMemStorage()
}

func main() {
	r, filename, err := handlers.MetricsRouter(mem)
	if err != nil {
		log.Fatal().Timestamp().Err(err).Msg("ошибка инициализации роутера")
	}
	log.Info().Timestamp().Msg("роутер инициализирован")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}
	log.Info().Timestamp().Msg("сервер запускается")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка работы сервера")
		}
	}()

	<-sigs
	log.Info().Timestamp().Msg("получен сигнал остановки")

	if filename != nil {
		err = utils.SaveMetricsToFile(mem, *filename)
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
