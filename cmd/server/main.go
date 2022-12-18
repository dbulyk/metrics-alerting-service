package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/storage"
	"log"
	"net/http"
)

func main() {
	hStorage := &storage.Handler{}
	http.Handle("/update/", hStorage)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
