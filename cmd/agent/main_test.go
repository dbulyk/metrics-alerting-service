package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCreateRequestToMetricsUpdate(t *testing.T) {
	builder := strings.Builder{}
	request, _ := createRequestToMetricsUpdate("key", "gauge", 123, builder)

	expectedEndpoint := endpoint + "update/gauge/key/123"
	assert.Equalf(t, expectedEndpoint, request.URL.String(), "олжидался эндпоинт %s, получен %s", expectedEndpoint, request.URL.String())
	assert.Equalf(t, http.MethodPost, request.Method, "ожидаемый метод отправки: %s, получен %s", http.MethodPost, request.Method)
	assert.Equalf(t, "text/plain", request.Header.Get("Content-Type"), "ожидаемый Content-Type: %s, получен %s", "text/plain", request.Header.Get("Content-Type"))
}

func TestCollectMetrics(t *testing.T) {
	ch := make(chan []models.Metric)
	go collectMetrics(ch, 1)

	metrics := <-ch
	assert.NotEqual(t, len(metrics), 0, "ожидался набор метрик, но получен пустой ответ")

	for _, m := range metrics {
		if m.Name == "" || m.Type == "" || m.Value == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
}

func TestCollectAndSendMetrics(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "=~^http://localhost:8080/update/(\\w+)/(\\w+)/(\\d+)",
		httpmock.NewStringResponder(200, ""))

	ch := make(chan []models.Metric)
	pollTicker := time.NewTicker(time.Millisecond * 10)
	reportTicker := time.NewTicker(time.Millisecond * 100)
	client := &http.Client{}
	var metrics []models.Metric
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		collectAndSendMetrics(ch, pollTicker, reportTicker, client, metrics)
	}()

	time.Sleep(time.Millisecond * 200)
	pollTicker.Stop()
	reportTicker.Stop()
	wg.Done()
}
