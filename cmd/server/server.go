package main

import (
	"context"

	"github.com/dbulyk/metrics-alerting-service/internal/fileio"
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"github.com/dbulyk/metrics-alerting-service/internal/middlewares"
	"github.com/dbulyk/metrics-alerting-service/internal/services"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/rs/zerolog/log"
)

func main() {
	defer os.Exit(0)
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	log.Info().Msg("получение кофигурации сервера")
	cfg := config.GetServerCfg()
	log.Info().Msg("конфигурация сервера получена")

	var (
		metrics storages.Repository
	)
	if len(cfg.DatabaseDsn) > 0 {
		log.Info().Msgf("подключение к базе данных по адресу: %s", cfg.DatabaseDsn)
		db, err := pgxpool.New(context.Background(), cfg.DatabaseDsn)
		if err != nil {
			log.Panic().Timestamp().Err(err).Msg("ошибка открытия соединения с базой данных")
		}
		defer db.Close()

		metrics = services.NewDBRepository(db)
	} else {
		metrics = services.NewFileRepository()
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.GzipMiddleware)
	metricHandler := handlers.NewRouter(router, &metrics)
	metricHandler.Register(router)
	log.Info().Msg("роутер инициализирован")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if len(cfg.DatabaseDsn) == 0 {
		writeTicker := startWriteToFile(cfg, metrics)
		if writeTicker != nil {
			defer func() {
				log.Info().Msgf("останавливаем тикер записи метрик в файл")
				writeTicker.Stop()
			}()
		}
	}

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Info().Msgf("сервер запускается на %s", cfg.Address)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка работы сервера")
		}
	}()

	<-sigs
	shutdown(cfg, srv, metrics)
}

func shutdown(cfg config.Server, srv *http.Server, mem storages.Repository) {
	log.Info().Msg("получен сигнал остановки")

	if len(cfg.StoreFile) > 0 && len(cfg.DatabaseDsn) == 0 {
		producer, err := fileio.NewProducer(cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка инициализации файла")
		} else {
			err = producer.Save(mem, cfg.StoreFile)
			if err != nil {
				log.Error().Timestamp().Err(err).Msg("ошибка сохранения метрики в файл")
			}
		}
	}

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Error().Timestamp().Err(err).Msg("ошибка остановки сервера")
	}

	log.Info().Msg("сервер остановлен")
}

func startWriteToFile(cfg config.Server, metrics storages.Repository) *time.Ticker {
	if cfg.Restore && len(cfg.StoreFile) > 0 {
		consumer, err := fileio.NewConsumer(cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("ошибка инициализации файла")
		} else {
			err = consumer.Restore(metrics)
			if err != nil {
				log.Error().Timestamp().Err(err).Msg("ошибка восстановления метрик")
			}
		}
	}

	if len(cfg.StoreFile) > 0 && cfg.StoreInterval > 0 {
		log.Info().Msgf("запуск записи метрик в файл с интервалом в %s секунд", cfg.StoreInterval)
		writeTicker := time.NewTicker(cfg.StoreInterval)
		go func() {
			for range writeTicker.C {
				producer, err := fileio.NewProducer(cfg.StoreFile)
				if err != nil {
					log.Error().Timestamp().Err(err).Msg("ошибка инициализации файла")
					return
				}

				err = producer.Save(metrics, cfg.StoreFile)
				if err != nil {
					log.Error().Timestamp().Err(err).Msg("ошибка сохранения метрик")
				}
			}
		}()
		return writeTicker
	}
	return nil
}
