package handlers

import (
	"fmt"
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
		r.Get("/value/{type}/{name}", Get)
		r.Post("/update/{type}/{name}/{value}", Update)
	})
	return r
}

func Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")

	if len(mType) == 0 || len(mName) == 0 || len(chi.URLParam(r, "value")) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mValue, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = mem.SetMetric(mName, mType, mValue)
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
	tmpl, _ := template.ParseFiles("templates/index.gohtml")
	err := tmpl.Execute(w, metrics)
	if err != nil {
		log.Print(err)
		return
	}
	//_, err := fmt.Fprint(w, metrics)
	//if err != nil {
	//	log.Print(err)
	//	return
	//}
}

func Get(w http.ResponseWriter, r *http.Request) {
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

	fmt.Fprint(w, metric.Value)
}
