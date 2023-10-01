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
	"runtime"
	"sync/atomic"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"

	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/dbulyk/metrics-alerting-service/config"
)

var (
	rtm runtime.MemStats
)

func main() {
	var done = make(chan bool)
	collectAndSendMetrics(done)
	done <- true
}

func collectAndSendMetrics(done chan bool) {
	cfg, err := config.NewAgentCfg()
	if err != nil {
		log.Fatalf("ошибка парсинга конфига: %v", err)
	}

	var (
		metrics   = make([]models.Metric, 0, 50)
		pollCount atomic.Int64
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
			log.Print("Сбор метрик")
			metrics = collectMetrics(&pollCount)
			pollCount.Add(1)
			log.Print("Сбор метрик завершен")
		case <-reportTicker.C:
			log.Print("Отправка метрик")

			err = sendRequestToMetricsUpdate(metrics, cfg.Address, cfg.Key)
			if err != nil {
				log.Printf("возникла ошибка при создании запроса. Ошибка: %s", err.Error())
				continue
			}
			log.Print("Отправка метрик завершена")
			pollCount.Swap(1)
		case <-done:
			return
		}
	}
}

func sendRequestToMetricsUpdate(metrics []models.Metric, address string, key string) error {
	if len(metrics) == 0 {
		log.Print("на вход пришла пустая коллекция")
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
		log.Printf("возникла ошибка при создании запроса. Ошибка: %s", err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("возникла ошибка при отправке запроса. Ошибка: %s", err.Error())
		return err
	}

	_, err = io.ReadAll(response.Body)
	if err != nil {
		log.Printf("возникла ошибка при чтении ответа. Ошибка: %s", err.Error())
		return err
	}

	err = response.Body.Close()
	if err != nil {
		log.Printf("возникла ошибка при закрытии тела ответа. Ошибка: %s", err.Error())
		return err
	}

	return nil
}

func collectMetrics(count *atomic.Int64) (metrics []models.Metric) {
	runtime.ReadMemStats(&rtm)
	randomValue := rand.Float64()
	countValue := count.Load()

	return []models.Metric{
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
