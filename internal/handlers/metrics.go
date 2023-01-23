package handlers

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/store"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

var mem store.MemStorage

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
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain")
	listMetric, _ := mem.ListMetrics()
	_, err := fmt.Fprint(w, listMetric)
	if err != nil {
		log.Print(err)
		return
	}
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
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, metric.Value)
}
