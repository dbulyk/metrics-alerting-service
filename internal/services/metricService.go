package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricsService struct {
	sync.Mutex
	ch              chan []models.Metric
	runtimeMetrics  []models.Metric
	advancedMetrics []models.Metric
	pollCount       *atomic.Int64
	reportInterval  time.Duration
	pollInterval    time.Duration
}

func NewMetricsService(reportInterval time.Duration, pollInterval time.Duration, rateLimit int) *MetricsService {
	pollCount := atomic.Int64{}
	pollCount.Store(1)
	runtimeMetrics := make([]models.Metric, 0, 50)
	advancedMetrics := make([]models.Metric, 0, 50)
	ch := make(chan []models.Metric, rateLimit)

	return &MetricsService{
		Mutex:           sync.Mutex{},
		reportInterval:  reportInterval,
		pollInterval:    pollInterval,
		pollCount:       &pollCount,
		runtimeMetrics:  runtimeMetrics,
		advancedMetrics: advancedMetrics,
		ch:              ch,
	}
}

func (ms *MetricsService) CollectRuntime(ctx context.Context) {
	ticker := time.NewTicker(ms.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rtm := runtime.MemStats{}
			runtime.ReadMemStats(&rtm)
			randomValue := rand.Float64()
			countValue := ms.pollCount.Load()
			metrics := []models.Metric{
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

			ms.Lock()
			ms.pollCount.Add(1)
			ms.runtimeMetrics = metrics
			ms.Unlock()
		}
	}
}

func (ms *MetricsService) CollectAdvanced(ctx context.Context) {
	ticker := time.NewTicker(ms.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			memory, err := mem.VirtualMemory()
			if err != nil {
				log.Error().Err(err).Msg("error collecting memory metrics")
				continue
			}

			cpuUtilization, err := cpu.Percent(0, true)
			if err != nil {
				log.Error().Err(err).Msg("error collecting cpu metrics")
				continue
			}

			metrics := []models.Metric{
				{
					ID:    "TotalMemory",
					MType: "gauge",
					Value: convertToPointerToFloat64(memory.Total),
				},
				{
					ID:    "FreeMemory",
					MType: "gauge",
					Value: convertToPointerToFloat64(memory.Available),
				},
			}

			for i := range cpuUtilization {
				metrics = append(metrics, models.Metric{
					ID:    fmt.Sprintf("CPUutilization%d", i+1),
					MType: "gauge",
					Value: &cpuUtilization[i],
				})
			}
			ms.Lock()
			ms.advancedMetrics = metrics
			ms.Unlock()
		}
	}
}

func (ms *MetricsService) MergeAndPushToQueue(ctx context.Context, key string) {
	ticker := time.NewTicker(ms.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(ms.ch)
			return
		case <-ticker.C:
			metrics := make([]models.Metric, 0, 100)

			ms.Lock()
			metrics = append(metrics, ms.runtimeMetrics...)
			metrics = append(metrics, ms.advancedMetrics...)
			ms.Unlock()

			if len(metrics) == 0 {
				log.Warn().Msg("no metrics to send")
				continue
			}

			if len(key) != 0 {
				for i := range metrics {
					switch metrics[i].MType {
					case "gauge":
						metrics[i].Hash = utils.Hash(fmt.Sprintf("%s:%s:%f", metrics[i].ID, metrics[i].MType, *metrics[i].Value), key)
					case "counter":
						metrics[i].Hash = utils.Hash(fmt.Sprintf("%s:%s:%d", metrics[i].ID, metrics[i].MType, *metrics[i].Delta), key)
					}
				}
			}

			ms.ch <- metrics
			ms.Lock()
			ms.pollCount.Swap(1)
			ms.Unlock()
			log.Info().Msg("metrics pushed to queue")
		}
	}
}

func (ms *MetricsService) Send(ctx context.Context, wg *sync.WaitGroup, client http.Client, address string) {
	defer wg.Done()

	for metrics := range ms.ch {
		jsonData, err := json.Marshal(metrics)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling metrics")
			continue
		}

		request, err := http.NewRequestWithContext(ctx,
			http.MethodPost,
			"http://"+address+"/updates/",
			bytes.NewBuffer(jsonData))
		if err != nil {
			log.Error().Err(err).Msg("error creating request")
			continue
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			log.Error().Err(err).Msg("error sending request")
			continue
		}

		_, err = io.ReadAll(response.Body)
		if err != nil {
			log.Error().Err(err).Msg("error reading response")
			continue
		}

		err = response.Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("error closing response body")
			continue
		}
	}
}

func convertToPointerToFloat64(par uint64) *float64 {
	f := math.Float64frombits(par)
	return &f
}
