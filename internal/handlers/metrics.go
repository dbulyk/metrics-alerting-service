package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var mem store.MemStorage

func MetricsRouter() chi.Router {
	r := chi.NewRouter()

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
	return r
}

func UpdateWithJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var m models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := mem.SetMetric(m.ID, m.MType, m.Value, m.Delta)
	if err != nil {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if mType == "gauge" {
		value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueFloat = &value
	} else if mType == "counter" {
		value, err := strconv.ParseInt(chi.URLParam(r, "value"), 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValueInt = &value
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("такого типа метрики не существует"))
		return
	}

	_, err := mem.SetMetric(mName, mType, mValueFloat, mValueInt)
	if err != nil {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	metrics, _ := mem.ListMetrics()

	tmpl, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Print(err)
		return
	}

	err = tmpl.Execute(w, metrics)
	if err != nil {
		log.Print(err)
		return
	}
}

func GetWithJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var m models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := mem.GetMetric(m.ID, m.MType)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func GetWithText(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if len(mType) == 0 || len(mName) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	metric, err := mem.GetMetric(mName, mType)
	if err != nil {
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
