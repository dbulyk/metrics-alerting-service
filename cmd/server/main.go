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
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"30s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	fs := flag.NewFlagSet("custom", flag.ContinueOnError)
	address := fs.String("a", "localhost:8080", "адрес сервера")
	storeInterval := fs.Duration("i", 30*time.Second, "Store interval duration (default: 30s)")
	storeFile := fs.String("f", "tmp/devops-metrics-db.json", "Store file path (default: tmp/devops-metrics-db.json)")
	restore := fs.Bool("r", true, "Restore flag (default: true)")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Error().Err(err).Msgf("ошибка парсинга флагов")
	}

	cfg = config{
		Address:       *address,
		StoreInterval: *storeInterval,
		StoreFile:     *storeFile,
		Restore:       *restore,
	}
	flag.Parse()

	err = env.Parse(&cfg)
	if err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка парсинга конфига")
	}

	mem = stores.NewMemStorage()
}

func main() {
	isAsync := cfg.StoreFile != "" && cfg.StoreInterval == 0
	r, err := handlers.MetricsRouter(mem, isAsync, cfg.StoreFile)
	if err != nil {
		log.Fatal().Timestamp().Err(err).Msg("ошибка инициализации роутера")
	}
	log.Info().Timestamp().Msg("роутер инициализирован")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if cfg.Restore {
		err := utils.RestoreMetricsFromFile(mem, cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка восстановления метрик")
		}
	}

	if cfg.StoreFile != "" && cfg.StoreInterval > 0 {
		writerTicker := time.NewTicker(cfg.StoreInterval)

		go func() {
			for range writerTicker.C {
				err := utils.SaveMetricsToFile(mem, cfg.StoreFile)
				if err != nil {
					return
				}
			}
			log.Info().Timestamp().Msg("остановка тикера")
		}()
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
		err = utils.SaveMetricsToFile(mem, cfg.StoreFile)
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
