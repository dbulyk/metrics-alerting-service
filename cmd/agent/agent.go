package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	rtm runtime.MemStats
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	collectAndSendMetrics(sigs)
}

func collectAndSendMetrics(sigs chan os.Signal) {
	cfg, err := Get()
	if err != nil {
		log.Fatalf("config parsing error: %v", err)
	}

	var (
		metrics   = make([]models.Metric, 0, 100)
		metricsCh = make(chan []models.Metric)
		pollCount atomic.Int64
		mutex     sync.Mutex
	)

	pollCount.Store(1)
	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()
	}()

	for {
		select {
		case <-pollTicker.C:
			go collectRuntimeMetrics(&pollCount, metricsCh)
			go collectAdvancedMetrics(metricsCh)

			metricsBuffer := <-metricsCh
			metricsBuffer = append(metricsBuffer, <-metricsCh...)

			mutex.Lock()
			metrics = metricsBuffer
			mutex.Unlock()
			pollCount.Add(1)
			log.Print("metrics collection is complete")
		case <-reportTicker.C:
			mutex.Lock()
			err = sendRequestToMetricsUpdate(metrics, cfg.Address, cfg.Key)
			mutex.Unlock()
			if err != nil {
				log.Printf("An error occurred while creating a query. Error: %s", err.Error())
				continue
			}
			log.Print("metrics submission complete")
			pollCount.Swap(1)
		case <-sigs:
			log.Print("a shutdown signal is received")
			return
		}
	}
}

func sendRequestToMetricsUpdate(metrics []models.Metric, address string, key string) error {
	if len(metrics) == 0 {
		log.Print("an empty collection came in")
		return nil
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
		return err
	}

	request, err := http.NewRequest(http.MethodPost, "http://"+address+"/updates/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

func collectRuntimeMetrics(count *atomic.Int64, metrics chan<- []models.Metric) {
	runtime.ReadMemStats(&rtm)
	randomValue := rand.Float64()
	countValue := count.Load()

	metrics <- []models.Metric{
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

func collectAdvancedMetrics(metrics chan<- []models.Metric) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("An error occurred while collecting metrics. Error: %s", err.Error())
		return
	}

	cpuUtilization, err := cpu.Percent(0, true)
	if err != nil {
		log.Printf("An error occurred while collecting metrics. Error: %s", err.Error())
		return
	}

	m := []models.Metric{
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
		m = append(m, models.Metric{
			ID:    fmt.Sprintf("CPUutilization%d", i+1),
			MType: "gauge",
			Value: &cpuUtilization[i],
		})
	}
	metrics <- m
}

func convertToPointerToFloat64(par uint64) *float64 {
	f := math.Float64frombits(par)
	return &f
}
