package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
)

type config struct {
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"30s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

var (
	mem *stores.MemStorage
	cfg config
)

func MetricsRouter(metrics *stores.MemStorage) (r chi.Router, storeFile *string, err error) {
	r = chi.NewRouter()
	mem = metrics

	fs := flag.NewFlagSet("custom", flag.ContinueOnError)
	storeInterval := fs.Duration("store-interval", 30*time.Second, "Store interval duration (default: 30s)")
	storeFile = fs.String("store-file", "tmp/devops-metrics-db.json", "Store file path (default: tmp/devops-metrics-db.json)")
	restore := fs.Bool("restore", true, "Restore flag (default: true)")

	err = fs.Parse(os.Args[1:])
	if err != nil {
		log.Error().Err(err).Msgf("ошибка парсинга флагов")
	}

	cfg = config{
		StoreInterval: *storeInterval,
		StoreFile:     *storeFile,
		Restore:       *restore,
	}

	err = env.Parse(&cfg)
	if err != nil {
		return nil, nil, err
	}

	if cfg.Restore {
		err := utils.RestoreMetricsFromFile(mem, cfg.StoreFile)
		if err != nil {
			return nil, nil, err
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
			print("Stopping writerTicker\n")
			writerTicker.Stop()
		}()
	}

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetAll)
		r.Get("/value/{type}/{name}", GetWithText)
		r.Post("/value/", GetWithJSON)
		r.Post("/update/{type}/{name}/{value}", UpdateWithText)
		r.Post("/update/", UpdateWithJSON)
	})
	return r, &cfg.StoreFile, nil
}

func UpdateWithJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var m models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msgf("ошибка декодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := mem.SetMetric(m.ID, m.MType, m.Value, m.Delta)
	if err != nil {
		log.Error().Err(err).Msg("ошибка обновления метрики")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(err.Error()))
		return
	}

	if cfg.StoreFile != "" && cfg.StoreInterval == 0 {
		err := utils.SaveMetricsToFile(mem, cfg.StoreFile)
		if err != nil {
			log.Error().Err(err).Msg("ошибка сохранения метрики в файл")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		log.Error().Err(err).Msg("ошибка кодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateWithText(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var (
		mValueFloat *float64
		mValueInt   *int64
	)

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if len(mType) == 0 || len(mName) == 0 || len(chi.URLParam(r, "value")) == 0 {
		log.Error().Msgf("тип, имя или значение метрики не заданы. mType: %s, mName: %s, mValue: %s",
			mType, mName, chi.URLParam(r, "value"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if mType == "gauge" {
		value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			log.Error().Err(err).Msgf("ошибка парсинга значения метрики: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueFloat = &value
	} else if mType == "counter" {
		value, err := strconv.ParseInt(chi.URLParam(r, "value"), 0, 64)
		if err != nil {
			log.Error().Err(err).Msgf("ошибка парсинга значения метрики: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueInt = &value
	} else {
		log.Error().Msgf("пришел несуществующий тип метрики: %s", mType)
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("такого типа метрики не существует"))
		return
	}

	_, err := mem.SetMetric(mName, mType, mValueFloat, mValueInt)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка обновления метрики: %s", mName)
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(err.Error()))
		return
	}

	if cfg.StoreFile != "" && cfg.StoreInterval == 0 {
		err := utils.SaveMetricsToFile(mem, cfg.StoreFile)
		if err != nil {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	metrics, _ := mem.ListMetrics()

	tmpl, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Error().Err(err).Msg("ошибка парсинга шаблона")
		return
	}

	err = tmpl.Execute(w, metrics)
	if err != nil {
		log.Error().Err(err).Msg("ошибка выполнения шаблона")
		return
	}
}

func GetWithJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var m models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msg("ошибка декодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := mem.GetMetric(m.ID, m.MType)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка получения метрики: %s", m.ID)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		log.Error().Err(err).Msg("ошибка кодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetWithText(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if len(mType) == 0 || len(mName) == 0 {
		log.Error().Msgf("тип или имя метрики не заданы. mType: %s, mName: %s", mType, mName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	metric, err := mem.GetMetric(mName, mType)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка получения метрики: %s", mName)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, err.Error())
		return
	}

	if mType == "counter" {
		fmt.Fprint(w, *metric.Delta)
	} else {
		fmt.Fprint(w, *metric.Value)
	}

	w.WriteHeader(http.StatusOK)
}
