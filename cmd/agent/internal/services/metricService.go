package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricService struct {
	sync.Mutex
	rtmMetrics      []models.Metric
	advancedMetrics []models.Metric
	pollCount       *atomic.Int64
	reportInterval  time.Duration
	pollInterval    time.Duration
}

func NewMetricService(reportInterval time.Duration, pollInterval time.Duration) *MetricService {
	pollCount := atomic.Int64{}
	pollCount.Store(1)

	return &MetricService{
		Mutex:           sync.Mutex{},
		rtmMetrics:      make([]models.Metric, 0, 100),
		advancedMetrics: make([]models.Metric, 0, 50),
		reportInterval:  reportInterval,
		pollInterval:    pollInterval,
		pollCount:       &pollCount,
	}
}

func (ms *MetricService) CollectRuntime(ctx context.Context) {
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
			metrics := make([]models.Metric, 0, 100)
			metrics = []models.Metric{
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
			ms.rtmMetrics = metrics
			ms.Unlock()
			log.Print("runtime metrics collected")
		}
	}
}

func (ms *MetricService) CollectAdvanced(ctx context.Context) {
	ticker := time.NewTicker(ms.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			memory, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("An error occurred while collecting metrics. Error: %s", err.Error())
				continue
			}

			cpuUtilization, err := cpu.Percent(0, true)
			if err != nil {
				log.Printf("An error occurred while collecting metrics. Error: %s", err.Error())
				continue
			}

			metrics := make([]models.Metric, 0, 50)
			metrics = []models.Metric{
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
			log.Print("advanced metrics collected")
		}
	}
}

func (ms *MetricService) Report(ctx context.Context, client *http.Client, address string, key string) {
	ticker := time.NewTicker(ms.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := make([]models.Metric, 0, 150)
			ms.Lock()
			metrics = append(metrics, ms.rtmMetrics...)
			metrics = append(metrics, ms.advancedMetrics...)
			ms.Unlock()

			if len(metrics) == 0 {
				log.Print("an empty collection came in")
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

			jsonData, err := json.Marshal(metrics)
			if err != nil {
				log.Printf("An error occurred while marshalling metrics. Error: %s", err.Error())
				continue
			}

			request, err := http.NewRequest(http.MethodPost, "http://"+address+"/updates/", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("An error occurred while creating request. Error: %s", err.Error())
				continue
			}
			request.Header.Set("Content-Type", "application/json")

			response, err := client.Do(request)
			if err != nil {
				log.Printf("An error occurred while sending request. Error: %s", err.Error())
				continue
			}

			_, err = io.ReadAll(response.Body)
			if err != nil {
				log.Printf("An error occurred while reading response. Error: %s", err.Error())
				continue
			}

			err = response.Body.Close()
			if err != nil {
				log.Printf("An error occurred while closing response body. Error: %s", err.Error())
				continue
			}
			ms.pollCount.Swap(1)
			log.Print("metrics sent")
		}
	}
}

func convertToPointerToFloat64(par uint64) *float64 {
	f := math.Float64frombits(par)
	return &f
}
