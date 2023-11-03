package main

import (
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSendRequestToMetricsUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, ""))

	metrics := []models.Metric{
		{
			ID:    "test",
			MType: "gauge",
			Value: new(float64),
			Delta: new(int64),
		},
	}

	err := sendRequestToMetricsUpdate(metrics, "localhost:8080", "test_key")
	assert.NoError(t, err)

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/updates/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}

func TestCollectMetrics(t *testing.T) {
	f := atomic.Int64{}
	metricsCh := make(chan []models.Metric)
	go collectRuntimeMetrics(&f, metricsCh)
	metrics := <-metricsCh
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

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, ""))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		collectAndSendMetrics(sigs)
	}()

	time.Sleep(time.Second * 15)
	sigs <- syscall.SIGINT

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/updates/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}
