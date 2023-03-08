package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/dbulyk/metrics-alerting-service/internal/handlers"
	"log"
	"net/http"
)

type config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := handlers.MetricsRouter()
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
