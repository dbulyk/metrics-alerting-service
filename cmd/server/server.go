package main

import (
	"context"
	"database/sql"

	"github.com/dbulyk/metrics-alerting-service/internal/fileio"
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"github.com/dbulyk/metrics-alerting-service/internal/middlewares"
	"github.com/dbulyk/metrics-alerting-service/internal/services"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"
	"github.com/go-chi/chi/v5/middleware"
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

	cfg := config.GetServerCfg()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var metrics storages.Repository

	if len(cfg.DatabaseDsn) > 0 {
		db, err := sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			log.Panic().Timestamp().Err(err).Msg("database connection opening error")
		}

		defer func(db *sql.DB) {
			err = db.Close()
			if err != nil {
				log.Error().Timestamp().Err(err).Msg("database connection closing error")
			}
		}(db)

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
	log.Info().Msg("router initialized")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if len(cfg.DatabaseDsn) == 0 {
		writeTicker := startWriteToFile(ctx, cfg, metrics)
		if writeTicker != nil {
			defer func() {
				log.Info().Msgf("stop the ticker writing metrics to the file")
				writeTicker.Stop()
			}()
		}
	}

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Info().Msgf("the server starts at %s", cfg.Address)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Timestamp().Err(err).Msg("server error")
		}
	}()

	<-sigs
	shutdown(ctx, cfg, srv, metrics)
}

func shutdown(ctx context.Context, cfg config.Server, srv *http.Server, mem storages.Repository) {
	if len(cfg.StoreFile) > 0 && len(cfg.DatabaseDsn) == 0 {
		producer, err := fileio.NewProducer(cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("file initialization error")
		} else {
			err = producer.Save(ctx, mem, cfg.StoreFile)
			if err != nil {
				log.Error().Timestamp().Err(err).Msg("error saving metrics to file")
			}
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Timestamp().Err(err).Msg("server stop error")
	}
}

func startWriteToFile(ctx context.Context, cfg config.Server, metrics storages.Repository) *time.Ticker {
	if cfg.Restore && len(cfg.StoreFile) > 0 {
		consumer, err := fileio.NewConsumer(cfg.StoreFile)
		if err != nil {
			log.Error().Timestamp().Err(err).Msg("file initialization error")
		} else {
			err = consumer.Restore(ctx, metrics)
			if err != nil {
				log.Error().Timestamp().Err(err).Msg("metrics recovery error")
			}
		}
	}

	if len(cfg.StoreFile) > 0 && cfg.StoreInterval > 0 {
		log.Info().Msgf("start recording metrics to a file at %s seconds intervals", cfg.StoreInterval)
		writeTicker := time.NewTicker(cfg.StoreInterval)
		go func() {
			for range writeTicker.C {
				producer, err := fileio.NewProducer(cfg.StoreFile)
				if err != nil {
					log.Error().Timestamp().Err(err).Msg("file initialization error")
					return
				}

				err = producer.Save(ctx, metrics, cfg.StoreFile)
				if err != nil {
					log.Error().Timestamp().Err(err).Msg("metrics saving error")
				}
			}
		}()
		return writeTicker
	}
	return nil
}
