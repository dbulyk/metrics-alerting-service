package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"log"
	"net/http"
)

func main() {
	http.Handle("/update/", new(handlers.UpdateHandler))
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
