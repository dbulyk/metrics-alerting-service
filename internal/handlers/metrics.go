package handlers

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

var mem storage.MemStorage

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

	statusCode := mem.Set(mType, mName, mValue)
	fmt.Print(statusCode)
	w.WriteHeader(statusCode)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprint(w, mem.GetAll())
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
	_, err := fmt.Fprint(w, mem.Get(mType, mName))
	if err != nil {
		log.Print(err)
		return
	}
}
