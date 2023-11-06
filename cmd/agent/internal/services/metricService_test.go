package services

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/jarcoal/httpmock"
)

func TestMetricService_Report(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, ""))

	metrics := NewMetricService(time.Second*2, time.Second*1)

	metrics.rtmMetrics = append(metrics.rtmMetrics, models.Metric{
		ID:    "test",
		MType: "gauge",
		Value: new(float64),
		Delta: new(int64),
	},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{}
	go metrics.Report(ctx, client, "localhost:8080", "test_key")

	time.Sleep(time.Second * 5)

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/updates/"] == 0 {
		t.Error("Ожидался запрос на сервер, но он не был получен")
	}
}

func TestMetricService_CollectRuntime(t *testing.T) {
	metrics := NewMetricService(time.Second*2, time.Second*1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go metrics.CollectRuntime(ctx)
	time.Sleep(time.Second * 5)

	assert.NotEqual(t, len(metrics.rtmMetrics), 0, "ожидался набор метрик, но получен пустой ответ")
	for _, m := range metrics.rtmMetrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
}

func TestMetricService_CollectAdvancedMetrics(t *testing.T) {
	metrics := NewMetricService(time.Second*2, time.Second*1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go metrics.CollectRuntime(ctx)
	time.Sleep(time.Second * 5)

	assert.NotEqual(t, len(metrics.rtmMetrics), 0, "ожидался набор метрик, но получен пустой ответ")
	for _, m := range metrics.rtmMetrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
}
