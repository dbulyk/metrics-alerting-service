package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"log"
	"net/http"
)

func main() {
	hStorage := &handlers.Handler{}
	http.Handle("/update/", hStorage)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
