package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/middlewares"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

var (
	mem            *stores.MemStorage
	isAsynchronous bool
	storeFile      string
)

func MetricsRouter(metrics *stores.MemStorage, isAsync bool, sFile string) (r chi.Router, err error) {
	r = chi.NewRouter()
	mem = metrics
	isAsynchronous = isAsync
	storeFile = sFile

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetAll)
		r.Get("/value/{type}/{name}", GetWithText)
		r.Post("/value/", GetWithJSON)
		r.Post("/update/{type}/{name}/{value}", UpdateWithText)
		r.Post("/update/", UpdateWithJSON)
	})
	return r, nil
}

func UpdateWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

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

	if isAsynchronous {
		err := utils.SaveMetrics(mem, storeFile)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

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

	if isAsynchronous {
		err := utils.SaveMetrics(mem, storeFile)
		if err != nil {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)
	metrics, _ := mem.ListMetrics()

	tmpl, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Error().Err(err).Msg("ошибка парсинга шаблона")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, metrics)
	if err != nil {
		log.Error().Err(err).Msg("ошибка выполнения шаблона")
		return
	}
}

func GetWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

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
