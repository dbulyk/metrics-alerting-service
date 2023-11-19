package services

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestMetricService_CollectRuntime(t *testing.T) {
	metrics := NewMetricsService(time.Second*2, time.Second*1, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go metrics.CollectRuntime(ctx)
	time.Sleep(time.Second * 5)

	metrics.Lock()
	assert.NotEqual(t, len(metrics.runtimeMetrics), 0, "a set of metrics was expected, but an empty response was received")
	for _, m := range metrics.runtimeMetrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
	metrics.Unlock()
}

func TestMetricService_CollectAdvancedMetrics(t *testing.T) {
	metrics := NewMetricsService(time.Second*2, time.Second*1, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go metrics.CollectAdvanced(ctx)
	time.Sleep(time.Second * 5)

	metrics.Lock()
	assert.NotEqual(t, len(metrics.advancedMetrics), 0, "a set of metrics was expected, but an empty response was received")
	for _, m := range metrics.advancedMetrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("ожидалось что все метрики будут иметь имя, тип и значение, но получено %v", m)
		}
	}
	metrics.Unlock()
}

func TestMetricService_MergeAndPushToQueue(t *testing.T) {
	metrics := NewMetricsService(time.Second*3, time.Second*1, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go metrics.CollectRuntime(ctx)
	metrics.MergeAndPushToQueue(ctx, "test")

	metrics.Lock()
	assert.NotNil(t, metrics.ch, "channel was expected, but nil was received")
	mRtm, ok := <-metrics.ch
	assert.Truef(t, ok, "channel was expected to be open, but it was closed")
	assert.NotEqual(t, len(mRtm), 0, "a set of metrics was expected, but an empty response was received")
	for _, m := range metrics.runtimeMetrics {
		if m.ID == "" || m.MType == "" || m.Value == nil && m.Delta == nil {
			t.Errorf("it was expected that all metrics would have name, type and value, but %v was received.", m)
		}
	}
	metrics.Unlock()
}

func TestMetricService_Send(t *testing.T) {
	metrics := NewMetricsService(time.Second*3, time.Second*1, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, ""))

	go metrics.CollectRuntime(ctx)

	agent := &http.Client{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	metrics.MergeAndPushToQueue(ctx, "test")
	metrics.Send(ctx, wg, *agent, "localhost:8080")

	assert.NotNil(t, metrics.ch, "channel was expected, but nil was received")

	time.Sleep(time.Second * 5)
	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8080/updates/"] == 0 {
		t.Error("a request to the server was expected, but it was not received")
	}
}
