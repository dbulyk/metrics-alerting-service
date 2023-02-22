package store

import (
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetMetric(t *testing.T) {
	storage := &MemStorage{}

	err := storage.SetMetric("test_metric_gauge", "gauge", 12.5)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")
	err = storage.SetMetric("test_metric_counter", "counter", 12)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")

	metrics, _ := storage.ListMetrics()
	assert.Lenf(t, metrics, 2, "ожидалось две метрики, получено %d", len(metrics))
	assert.Equalf(t, "test_metric_gauge", metrics[0].Name, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metrics[0].Name)
	assert.Equalf(t, "gauge", metrics[0].Type, "ожидаемый тип метрики 'gauge', получено %s", metrics[0].Type)
	assert.EqualValuesf(t, 12.5, metrics[0].Value.(models.Gauge), "ожидаемое значение метрики 12.5, получено %f", metrics[0].Value.(models.Gauge))
	assert.EqualValuesf(t, 12, metrics[1].Value.(models.Counter), "ожидаемое значение метрики 12, получено %d", metrics[1].Value.(models.Counter))

	err = storage.SetMetric("test_metric_counter", "counter", 12)
	assert.NoErrorf(t, err, "ошибка обновления существующей метрики: %v", err)

	metrics, _ = storage.ListMetrics()
	assert.EqualValuesf(t, 24, metrics[1].Value.(models.Counter), "ожидаемое значение метрики 24, получено %d", metrics[1].Value.(models.Counter))
}

func TestGetMetric(t *testing.T) {
	storage := &MemStorage{}
	storage.SetMetric("test_metric_gauge", "gauge", 12.5)

	metric1, err := storage.GetMetric("test_metric_gauge", "gauge")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_gauge", metric1.Name, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metric1.Name)
	assert.Equalf(t, "gauge", metric1.Type, "ожидаемый тип метрики 'gauge', получено %s", metric1.Type)
	assert.EqualValuesf(t, 12.5, metric1.Value.(models.Gauge), "ожидаемое значение метрики 12.5, получено %f", metric1.Value.(models.Gauge))

	storage.SetMetric("test_metric_counter2", "counter", 12.5)

	metric2, err := storage.GetMetric("test_metric_counter2", "counter")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_counter2", metric2.Name, "ожидаемое имя метрики: 'test_metric_counter2', получено %s", metric2.Name)
	assert.Equalf(t, "counter", metric2.Type, "ожидаемый тип метрики 'counter', получено %s", metric2.Type)
	assert.EqualValuesf(t, 12, metric2.Value.(models.Counter), "ожидаемое значение метрики 12, получено %d", metric2.Value.(models.Counter))

	_, err = storage.GetMetric("non_existing", "gauge")
	assert.Error(t, err, "ожидалась ошибка получения несуществующей метрики")
}
