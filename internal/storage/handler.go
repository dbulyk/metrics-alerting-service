package storage

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var storage = &MemStorage{
	Alloc:         0,
	BuckHashSys:   0,
	Frees:         0,
	GcCPUFraction: 0,
	GcSys:         0,
	HeapAlloc:     0,
	HeapIdle:      0,
	HeapInuse:     0,
	HeapObjects:   0,
	HeapReleased:  0,
	HeapSys:       0,
	LastGC:        0,
	Lookups:       0,
	MCacheInuse:   0,
	MCacheSys:     0,
	MSpanInuse:    0,
	MSpanSys:      0,
	Mallocs:       0,
	NextGC:        0,
	NumForcedGC:   0,
	NumGC:         0,
	OtherSys:      0,
	PauseTotalNs:  0,
	StackInuse:    0,
	StackSys:      0,
	Sys:           0,
	TotalAlloc:    0,
	RandomValue:   0,
	PollCount:     0,
}

type (
	Handler struct{}
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/update/" && r.Method == http.MethodPost {
		h.Update(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	contents, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	values := strings.Split(string(contents), "/")
	fType := values[0]
	fName := values[1]
	fValue, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	statusCode := storage.SetMetric(fType, fName, fValue)
	fmt.Print(storage)

	w.WriteHeader(statusCode)
}
