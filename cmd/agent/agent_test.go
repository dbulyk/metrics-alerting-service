package main

import (
	"github.com/dbulyk/metrics-alerting-service/internal/metric"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateRequestToMetricsUpdate(t *testing.T) {
	val := 8.18
	metric := metric.Metric{
		ID:    "test",
		MType: "gauge",
		Delta: nil,
		Value: &val,
	}
	request, err := createRequestToMetricsUpdate(&metric, "localhost:8080", "test")

	expectedEndpoint := "http://localhost:8080/update/"
	assert.NoErrorf(t, err, "функция не должна была вернуть ошибку, но вернула %s", err)
	assert.Equalf(t, expectedEndpoint, request.URL.String(), "ожидался эндпоинт %s, получен %s", expectedEndpoint, request.URL.String())
	assert.Equalf(t, http.MethodPost, request.Method, "ожидаемый метод отправки: %s, получен %s", http.MethodPost, request.Method)
	assert.Equalf(t, "application/json", request.Header.Get("Content-Type"), "ожидаемый Content-Type: %s, получен %s", "application/json", request.Header.Get("Content-Type"))
}
func TestCollectMetrics(t *testing.T) {
	f := atomic.Int64{}
	metrics := collectMetrics(&f)
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

	done := make(chan bool)
	go func() {
		collectAndSendMetrics(done)
	}()

	time.Sleep(time.Second * 15)
	done <- true

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/update/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}
