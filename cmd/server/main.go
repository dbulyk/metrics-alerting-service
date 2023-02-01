package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"log"
	"net/http"
)

func main() {
	r := handlers.MetricsRouter()
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", r))
}
