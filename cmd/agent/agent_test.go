package main

import (
	"sync/atomic"
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

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, ""))

	done := make(chan bool)
	go func() {
		collectAndSendMetrics(done)
	}()

	time.Sleep(time.Second * 15)
	done <- true

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/updates/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}
