package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAll)
		r.Get("/value/{type}/{name}", handlers.Get)
		r.Post("/update/{type}/{name}/{value}", handlers.Update)
	})
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", r))
}
