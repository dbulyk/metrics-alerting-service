package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/dbulyk/metrics-alerting-service/internal/handlers"

	"github.com/dbulyk/metrics-alerting-service/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

//var (
//	mem *Repository
//)

type handler struct {
	repository Repository
	r          chi.Router
}

func NewRouter(router *chi.Mux, metrics *Repository) (r handlers.Handler) {
	return &handler{
		repository: *metrics,
		r:          router,
	}
}

func (h *handler) Register(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middlewares.GzipMiddleware)

	router.Route("/", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/value/{type}/{name}", h.GetWithText)
		r.Post("/value/", h.GetWithJSON)
		r.Post("/update/{type}/{name}/{value}", h.UpdateWithText)
		r.Post("/update/", h.UpdateWithJSON)
	})
}

func (h *handler) UpdateWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

	var m Metric
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msgf("ошибка декодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.repository.SetMetric(m, false)
	if err != nil {
		if errors.Is(err, ErrInvalidHash) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		log.Error().Err(err).Msg("ошибка кодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) UpdateWithText(w http.ResponseWriter, r *http.Request) {
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
	mHash := chi.URLParam(r, "hash")

	if len(mType) == 0 || len(mName) == 0 || len(chi.URLParam(r, "value")) == 0 {
		log.Error().Msgf("тип, имя или значение метрики не заданы. mType: %s, mName: %s, mValue: %s",
			mType, mName, chi.URLParam(r, "value"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch mType {
	case gauge:
		value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			log.Error().Err(err).Msgf("ошибка парсинга значения метрики: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueFloat = &value
	case counter:
		value, err := strconv.ParseInt(chi.URLParam(r, "value"), 0, 64)
		if err != nil {
			log.Error().Err(err).Msgf("ошибка парсинга значения метрики: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueInt = &value
	default:
		log.Error().Msgf("пришел несуществующий тип метрики: %s", mType)
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("такого типа метрики не существует"))
		return
	}

	metric := Metric{
		ID:    mName,
		MType: mType,
		Value: mValueFloat,
		Delta: mValueInt,
		Hash:  mHash,
	}
	_, err := h.repository.SetMetric(metric, false)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка обновления метрики: %s", mName)
		w.WriteHeader(http.StatusNotImplemented)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)
	metrics, _ := h.repository.ListMetrics()

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

func (h *handler) GetWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("ошибка закрытия тела запроса")
		}
	}(r.Body)

	var m Metric

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msg("ошибка декодирования JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.repository.GetMetric(m.ID, m.MType)
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

func (h *handler) GetWithText(w http.ResponseWriter, r *http.Request) {
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
	metric, err := h.repository.GetMetric(mName, mType)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка получения метрики: %s", mName)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, err.Error())
		return
	}

	if mType == counter {
		fmt.Fprint(w, *metric.Delta)
	} else {
		fmt.Fprint(w, *metric.Value)
	}

	w.WriteHeader(http.StatusOK)
}
