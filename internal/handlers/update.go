package handlers

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

var mem storage.MemStorage

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.Update(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	values := strings.Split(r.URL.Path, "/")
	if len(values) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mType := values[2]
	mName := values[3]
	mValue, err := strconv.ParseFloat(values[4], 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	statusCode := mem.SetMetric(mType, mName, mValue)
	fmt.Print(mem.Metrics["PollCount"])

	w.WriteHeader(statusCode)
}
