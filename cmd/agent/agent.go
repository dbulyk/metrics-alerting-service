package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dbulyk/metrics-alerting-service/cmd/agent/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dbulyk/metrics-alerting-service/cmd/agent/internal/services"
)

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	agentCfg := &config.AgentCfg{}
	cfg, err := agentCfg.GetAgentConfig()
	if err != nil {
		log.Panic().Err(err).Msg("config parsing error")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	agent := &http.Client{}
	metrics := services.NewMetricsService(cfg.ReportInterval, cfg.PollInterval, cfg.RateLimit)

	go metrics.CollectRuntime(ctx)
	go metrics.CollectAdvanced(ctx)
	go metrics.MergeAndPushToQueue(ctx, cfg.Key)

	wg := &sync.WaitGroup{}
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go metrics.Send(ctx, wg, agent, cfg.Address)
	}

	<-ctx.Done()
	wg.Wait()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	<-shutdownContext.Done()
	log.Info().Msg("agent shutdown")
}
