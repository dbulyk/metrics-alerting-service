package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	rtm runtime.MemStats
	cfg config
)

type config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

func init() {
	fs := flag.NewFlagSet("custom", flag.ContinueOnError)
	parseFlags(fs, os.Args[1:])

	err := env.Parse(&cfg)
	if err != nil {
		log.Print(err)
	}
}

func main() {
	var (
		metrics = make([]models.Metrics, 0, 50)
		ch      = make(chan []models.Metrics)
		done    = make(chan bool)
		client  = &http.Client{}
	)

	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	collectAndSendMetrics(ch, pollTicker, reportTicker, client, metrics, cfg.Address, done)
	reportTicker.Stop()
	pollTicker.Stop()
	done <- true
}

func collectAndSendMetrics(
	ch chan []models.Metrics,
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
	client *http.Client,
	metrics []models.Metrics,
	address string,
	done chan bool,
) {
	var pollCount atomic.Int64
	pollCount.Store(1)

	for {
		select {
		case <-pollTicker.C:
			go collectMetrics(ch, &pollCount)
			metrics = <-ch
			pollCount.Add(1)
		case <-reportTicker.C:
			isError := false
			for _, m := range metrics {
				request, err := createRequestToMetricsUpdate(m, address)
				if err != nil {
					isError = true
					log.Printf("возникла ошибка при создании запроса. Ошибка: %s", err.Error())
					continue
				}

				response, err := client.Do(request)
				if err != nil {
					isError = true
					log.Printf("возникла ошибка при отправке запроса. Ошибка: %s", err.Error())
					continue
				}

				_, err = io.ReadAll(response.Body)
				if err != nil {
					isError = true
					log.Printf("возникла ошибка при чтении ответа. Ошибка: %s", err.Error())
				}
				response.Body.Close()
			}

			if !isError {
				pollCount.Swap(1)
			}
		case <-done:
			return
		}
	}
}

func createRequestToMetricsUpdate(metrics models.Metrics, address string) (*http.Request, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "http://"+address+"/update/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func collectMetrics(ch chan []models.Metrics, count *atomic.Int64) {
	runtime.ReadMemStats(&rtm)
	randomValue := rand.Float64()
	countValue := count.Load()

	ch <- []models.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.Alloc),
		},
		{
			ID:    "BuckHashSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.BuckHashSys),
		},
		{
			ID:    "Frees",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.Frees),
		},
		{
			ID:    "GCCPUFraction",
			MType: "gauge",
			Value: &rtm.GCCPUFraction,
		},
		{
			ID:    "GCSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.GCSys),
		},
		{
			ID:    "HeapAlloc",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapAlloc),
		},
		{
			ID:    "HeapIdle",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapIdle),
		},
		{
			ID:    "HeapInuse",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapInuse),
		},
		{
			ID:    "HeapObjects",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapObjects),
		},
		{
			ID:    "HeapReleased",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapReleased),
		},
		{
			ID:    "HeapSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.HeapSys),
		},
		{
			ID:    "LastGC",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.LastGC),
		},
		{
			ID:    "Lookups",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.Lookups),
		},
		{
			ID:    "MCacheInuse",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.MCacheInuse),
		},
		{
			ID:    "MCacheSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.MCacheSys),
		},
		{
			ID:    "MSpanInuse",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.MSpanInuse),
		},
		{
			ID:    "MSpanSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.MSpanSys),
		},
		{
			ID:    "Mallocs",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.Mallocs),
		},
		{
			ID:    "NextGC",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.NextGC),
		},
		{
			ID:    "NumForcedGC",
			MType: "gauge",
			Value: convertToPointerToFloat64(uint64(rtm.NumForcedGC)),
		},
		{
			ID:    "NumGC",
			MType: "gauge",
			Value: convertToPointerToFloat64(uint64(rtm.NumGC)),
		},
		{
			ID:    "OtherSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.OtherSys),
		},
		{
			ID:    "PauseTotalNs",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.PauseTotalNs),
		},
		{
			ID:    "StackInuse",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.StackInuse),
		},
		{
			ID:    "StackSys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.StackSys),
		},
		{
			ID:    "Sys",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.Sys),
		},
		{
			ID:    "TotalAlloc",
			MType: "gauge",
			Value: convertToPointerToFloat64(rtm.TotalAlloc),
		},
		{
			ID:    "PollCount",
			MType: "counter",
			Delta: &countValue,
		},
		{
			ID:    "RandomValue",
			MType: "gauge",
			Value: &randomValue,
		},
	}
}

func convertToPointerToFloat64(par uint64) *float64 {
	f := math.Float64frombits(par)
	return &f
}

func parseFlags(fs *flag.FlagSet, args []string) {
	address := fs.String("address", "localhost:8080", "Server address (default: localhost:8080)")
	reportInterval := fs.Duration("report-interval", 10*time.Second, "Report interval duration (default: 10s)")
	pollInterval := fs.Duration("poll-interval", 2*time.Second, "Poll interval duration (default: 2s)")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return
	}

	cfg = config{
		Address:        *address,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}
}
