package main

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
	endpoint       = "http://localhost:8080/"
)

var (
	rtm runtime.MemStats
)

func main() {
	var metrics = make([]models.Metric, 0, 50)
	ch := make(chan []models.Metric)
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	client := &http.Client{}

	collectAndSendMetrics(ch, pollTicker, reportTicker, client, metrics)
	reportTicker.Stop()
	pollTicker.Stop()
}

func collectAndSendMetrics(
	ch chan []models.Metric,
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
	client *http.Client,
	metrics []models.Metric,
) {
	var (
		pollCount int64 = 1
		builder   strings.Builder
	)
	for {
		select {
		case <-pollTicker.C:
			go collectMetrics(ch, pollCount)
			metrics = <-ch
			pollCount += 1
		case <-reportTicker.C:
			isError := false
			for _, m := range metrics {
				request, err := createRequestToMetricsUpdate(m.Name, m.Type, m.Value, builder)
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
				pollCount = 1
			}
		}
	}
}

func createRequestToMetricsUpdate(key string, mType string, value interface{}, builder strings.Builder) (*http.Request, error) {
	fmt.Fprintf(&builder, "%s/%s/%v",
		mType, key, value)

	request, err := http.NewRequest(http.MethodPost, endpoint+"update/"+builder.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "text/plain")
	builder.Reset()
	return request, nil
}

func collectMetrics(ch chan []models.Metric, count int64) {
	runtime.ReadMemStats(&rtm)

	ch <- []models.Metric{
		{
			Name:  "Alloc",
			Type:  "gauge",
			Value: models.Gauge(rtm.Alloc),
		},
		{
			Name:  "BuckHashSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.BuckHashSys),
		},
		{
			Name:  "Frees",
			Type:  "gauge",
			Value: models.Gauge(rtm.Frees),
		},
		{
			Name:  "GcCPUFraction",
			Type:  "gauge",
			Value: models.Gauge(rtm.GCCPUFraction),
		},
		{
			Name:  "GcSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.GCSys),
		},
		{
			Name:  "HeapAlloc",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapAlloc),
		},
		{
			Name:  "HeapIdle",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapIdle),
		},
		{
			Name:  "HeapInuse",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapInuse),
		},
		{
			Name:  "HeapObjects",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapObjects),
		},
		{
			Name:  "HeapReleased",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapReleased),
		},
		{
			Name:  "HeapSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.HeapSys),
		},
		{
			Name:  "LastGC",
			Type:  "gauge",
			Value: models.Gauge(rtm.LastGC),
		},
		{
			Name:  "Lookups",
			Type:  "gauge",
			Value: models.Gauge(rtm.Lookups),
		},
		{
			Name:  "MCacheInuse",
			Type:  "gauge",
			Value: models.Gauge(rtm.MCacheInuse),
		},
		{
			Name:  "MCacheSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.MCacheSys),
		},
		{
			Name:  "MSpanInuse",
			Type:  "gauge",
			Value: models.Gauge(rtm.MSpanInuse),
		},
		{
			Name:  "MSpanSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.MSpanSys),
		},
		{
			Name:  "Mallocs",
			Type:  "gauge",
			Value: models.Gauge(rtm.Mallocs),
		},
		{
			Name:  "NextGC",
			Type:  "gauge",
			Value: models.Gauge(rtm.NextGC),
		},
		{
			Name:  "NumForcedGC",
			Type:  "gauge",
			Value: models.Gauge(rtm.NumForcedGC),
		},
		{
			Name:  "NumGC",
			Type:  "gauge",
			Value: models.Gauge(rtm.NumGC),
		},
		{
			Name:  "OtherSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.OtherSys),
		},
		{
			Name:  "PauseTotalNs",
			Type:  "gauge",
			Value: models.Gauge(rtm.PauseTotalNs),
		},
		{
			Name:  "StackInuse",
			Type:  "gauge",
			Value: models.Gauge(rtm.StackInuse),
		},
		{
			Name:  "StackSys",
			Type:  "gauge",
			Value: models.Gauge(rtm.StackSys),
		},
		{
			Name:  "Sys",
			Type:  "gauge",
			Value: models.Gauge(rtm.Sys),
		},
		{
			Name:  "TotalAlloc",
			Type:  "gauge",
			Value: models.Gauge(rtm.TotalAlloc),
		},
		{
			Name:  "PollCount",
			Type:  "counter",
			Value: models.Counter(count),
		},
		{
			Name:  "RandomValue",
			Type:  "gauge",
			Value: models.Gauge(rand.Float64()),
		},
	}
}
