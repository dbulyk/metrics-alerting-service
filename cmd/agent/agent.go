package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbulyk/metrics-alerting-service/cmd/agent/internal/services"
)

func main() {
	agent := &http.Client{}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := Get()
	if err != nil {
		log.Fatalf("config parsing error: %v", err)
	}

	metrics := services.NewMetricService(cfg.ReportInterval, cfg.PollInterval)

	go metrics.CollectRuntime(ctx)
	go metrics.CollectAdvanced(ctx)
	go metrics.Report(ctx, agent, cfg.Address, cfg.Key)

	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	<-shutdownContext.Done()
	log.Print("agent stopped")
}
