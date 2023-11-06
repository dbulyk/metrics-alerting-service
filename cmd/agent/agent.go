package main

import (
	"context"
	"github.com/dbulyk/metrics-alerting-service/cmd/agent/config"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbulyk/metrics-alerting-service/cmd/agent/internal/services"
)

func main() {
	agentCfg := &config.AgentCfg{}
	cfg, err := agentCfg.GetAgentConfig()
	if err != nil {
		log.Panicf("config parsing error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	agent := &http.Client{}
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
