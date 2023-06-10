package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateRequestToMetricsUpdate(t *testing.T) {
	val := 8.18
	metric := models.Metrics{
		ID:    "test",
		MType: "gauge",
		Delta: nil,
		Value: &val,
	}
	request, err := createRequestToMetricsUpdate(metric, "localhost:8080")

	expectedEndpoint := "http://localhost:8080/update/"
	assert.NoErrorf(t, err, "функция не должна была вернуть ошибку, но вернула %s", err)
	assert.Equalf(t, expectedEndpoint, request.URL.String(), "ожидался эндпоинт %s, получен %s", expectedEndpoint, request.URL.String())
	assert.Equalf(t, http.MethodPost, request.Method, "ожидаемый метод отправки: %s, получен %s", http.MethodPost, request.Method)
	assert.Equalf(t, "application/json", request.Header.Get("Content-Type"), "ожидаемый Content-Type: %s, получен %s", "application/json", request.Header.Get("Content-Type"))
}

func TestCollectMetrics(t *testing.T) {
	ch := make(chan []models.Metrics)
	f := atomic.Int64{}
	go collectMetrics(ch, &f)

	metrics := <-ch
	assert.NotEqual(t, len(metrics), 0, "ожидался набор метрик, но получен пустой ответ")

	for _, m := range metrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
}

func TestCollectAndSendMetrics(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/update/",
		httpmock.NewStringResponder(200, ""))

	ch := make(chan []models.Metrics)
	pollTicker := time.NewTicker(time.Millisecond * 10)
	reportTicker := time.NewTicker(time.Millisecond * 100)
	client := &http.Client{}
	var metrics []models.Metrics
	done := make(chan bool)

	go func() {
		collectAndSendMetrics(ch, pollTicker, reportTicker, client, metrics, "localhost:8080", done)
	}()

	time.Sleep(time.Millisecond * 200)
	pollTicker.Stop()
	reportTicker.Stop()
	done <- true

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/update/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}
