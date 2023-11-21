package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/services"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

//	@Title			Metrics Alerting Service API
//	@Description	This is a metrics alerting service API.
//	@Version		1
//	@BasePath		/

type handler struct {
	repository storages.IRepository
	r          chi.Router
}

// NewRouter creates a new handler and returns a pointer to it.
func NewRouter(router *chi.Mux, rep *storages.IRepository) (r Handler) {
	return &handler{
		repository: *rep,
		r:          router,
	}
}

// Register registers all metric handlers.
func (h *handler) Register(router *chi.Mux) {
	router.Route("/", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/value/{type}/{name}", h.GetWithText)
		r.Post("/value/", h.GetWithJSON)
		r.Post("/update/{type}/{name}/{value}", h.UpdateWithText)
		r.Post("/update/{type}/{name}/{value}/{hash}", h.UpdateWithText)
		r.Post("/update/", h.UpdateWithJSON)
		r.Post("/updates/", h.Updates)
		r.Get("/ping", h.Ping)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	})
}

// UpdateWithJSON updates metrics with JSON.
//
//	@Description	Updates metrics with JSON.
//	@Accept			json
//	@Produce		json
//	@Param			metric	body		models.Metric	true	"metric"
//	@Success		200		{object}	models.Metric
//	@Failure		400		{string}	string
//	@Failure		501		{string}	string
//	@Router			/update/ [post]
func (h *handler) UpdateWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
		}
	}(r.Body)

	var m models.Metric
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msgf("JSON decoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	metric, err := h.repository.Set(ctx, m)
	if err != nil {
		if errors.Is(err, services.ErrInvalidHash) {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Error().Err(err).Msg("response writing error")
			}
			return
		}
		w.WriteHeader(http.StatusNotImplemented)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		log.Error().Err(err).Msg("JSON encoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
			return
		}
		return
	}
}

// UpdateWithText updates metrics with text.
//
//	@Description	Updates metrics with text.
//	@Param			type	path		string	true	"metric type"
//	@Param			name	path		string	true	"metric name"
//	@Param			value	path		string	true	"metric value"
//	@Success		200		{object}	models.Metric
//	@Failure		400		{string}	string
//	@Failure		501		{string}	string
//	@Failure		404		{string}	string
//	@Router			/update/{type}/{name}/{value} [post]
func (h *handler) UpdateWithText(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
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
		log.Error().Msgf("metric type, name, or value is not specified. mType: %s, mName: %s, mValue: %s",
			mType, mName, chi.URLParam(r, "value"))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch mType {
	case services.Gauge:
		value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			log.Error().Err(err).Msgf("metric value parsing error: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Error().Err(err).Msg("response writing error")
			}
			return
		}
		mValueFloat = &value
	case services.Counter:
		value, err := strconv.ParseInt(chi.URLParam(r, "value"), 0, 64)
		if err != nil {
			log.Error().Err(err).Msgf("metric value parsing error: %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Error().Err(err).Msg("response writing error")
			}
			return
		}
		mValueInt = &value
	}

	metric := models.Metric{
		ID:    mName,
		MType: mType,
		Value: mValueFloat,
		Delta: mValueInt,
		Hash:  mHash,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := h.repository.Set(ctx, metric)
	if err != nil {
		log.Error().Err(err).Msgf("metric %s update error ", mName)
		w.WriteHeader(http.StatusNotImplemented)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}
}

// GetAll returns all metrics.
//
//	@Description	Returns all metrics.
//	@Produce		html
//	@Success		200	{array}		models.Metric
//	@Failure		500	{string}	string
//	@Router			/ [get]
func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
		}
	}(r.Body)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	metrics, _ := h.repository.GetAll(ctx)

	tmpl, err := template.ParseFiles(filepath.Join("internal", "templates", "index.gohtml"))
	if err != nil {
		log.Error().Err(err).Msg("template parsing error")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, metrics)
	if err != nil {
		log.Error().Err(err).Msg("template execution error")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}
}

// GetWithJSON returns a metric in application/json content-type.
//
//	@Description	Returns a metric in application/json content-type.
//	@Accept			json
//	@Produce		json
//	@Param			metric	body		models.Metric	true	"metric"
//	@Success		200		{string}	string
//	@Failure		400		{string}	string
//	@Failure		404		{string}	string
//	@Router			/value/ [post]
func (h *handler) GetWithJSON(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
		}
	}(r.Body)

	var m models.Metric

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error().Err(err).Msg("JSON decoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	metric, err := h.repository.Get(ctx, m.ID, m.MType)
	if err != nil {
		log.Error().Err(err).Msgf("metric %s retrieval error", m.ID)
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(metric); err != nil {
		log.Error().Err(err).Msg("JSON encoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}
}

// GetWithText returns a metric in text/plain content type.
//
//	@Description	Returns a metric in text/plain content type.
//	@Param			type	path		string	true	"metric type"
//	@Param			name	path		string	true	"metric name"
//	@Success		200		{string}	string
//	@Failure		400		{string}	string
//	@Failure		404		{string}	string
//	@Router			/value/{type}/{name} [get]
func (h *handler) GetWithText(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
			return
		}
	}(r.Body)

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if len(mType) == 0 || len(mName) == 0 {
		log.Error().Msgf("metric type or name is not specified. mType: %s, mName: %s", mType, mName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	metric, err := h.repository.Get(ctx, mName, mType)
	if err != nil {
		log.Error().Err(err).Msgf("metric %s retrieval error", mName)
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if mType == services.Counter {
		fmt.Fprint(w, *metric.Delta)
	} else {
		fmt.Fprint(w, *metric.Value)
	}
}

// Updates handles HTTP requests to update metrics.
// It decodes the JSON request body into metrics, updates them using a 5-second context,
// and sends the updated metrics back as a JSON response.
//
//	@Description	Handles HTTP requests to update metrics.
//	@Accept			json
//	@Produce		json
//	@Param			metrics	body		[]models.Metric	true	"metrics"
//	@Success		200		{object}	models.Metric
//	@Failure		400		{string}	string
//	@Router			/updates/ [post]
func (h *handler) Updates(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
		}
	}(r.Body)

	var metrics []models.Metric

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		log.Error().Err(err).Msg("JSON decoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	var metricsResp []models.Metric

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	metricsResp, err := h.repository.Updates(ctx, metrics)
	if err != nil {
		log.Error().Err(err).Msg("metrics update error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(metricsResp); err != nil {
		log.Error().Err(err).Msg("JSON encoding error")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
		}
		return
	}
}

// Ping check connection to db or always return 200 if we have a file repository.
//
//	@Description	Check connection to db or always return 200 if we have a file repository.
//	@Success		200	{string}	string
//	@Failure		500	{string}	string
//	@Router			/ping [get]
func (h *handler) Ping(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("request body closing error")
		}
	}(r.Body)

	err := h.repository.Ping()
	if err != nil {
		log.Error().Err(err).Msg("ping error")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("response writing error")
			return
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
