package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetMetric(t *testing.T) {
	storage := &MemStorage{}

	value := 12.5
	_, err := storage.SetMetric("test_metric_gauge", "gauge", &value, nil)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")
	delta := int64(12)
	_, err = storage.SetMetric("test_metric_counter", "counter", nil, &delta)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")

	metrics, _ := storage.ListMetrics()
	assert.Lenf(t, metrics, 2, "ожидалось две метрики, получено %d", len(metrics))
	assert.Equalf(t, "test_metric_gauge", metrics[0].ID, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metrics[0].ID)
	assert.Equalf(t, "gauge", metrics[0].MType, "ожидаемый тип метрики 'gauge', получено %s", metrics[0].MType)
	assert.EqualValuesf(t, &value, metrics[0].Value, "ожидаемое значение метрики 12.5, получено %f", *metrics[0].Value)
	assert.EqualValuesf(t, &delta, metrics[1].Delta, "ожидаемое значение метрики 12, получено %d", *metrics[1].Delta)

	_, err = storage.SetMetric("test_metric_counter", "counter", nil, &delta)
	assert.NoErrorf(t, err, "ошибка обновления существующей метрики: %v", err)

	resultDelta := int64(24)
	metrics, _ = storage.ListMetrics()
	assert.EqualValuesf(t, &resultDelta, metrics[1].Delta, "ожидаемое значение метрики 24, получено %p", metrics[1].Delta)
}

func TestGetMetric(t *testing.T) {
	storage := &MemStorage{}

	value := 12.5
	_, err := storage.SetMetric("test_metric_gauge", "gauge", &value, nil)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")

	metric1, err := storage.GetMetric("test_metric_gauge", "gauge")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_gauge", metric1.ID, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metric1.ID)
	assert.Equalf(t, "gauge", metric1.MType, "ожидаемый тип метрики 'gauge', получено %s", metric1.MType)
	assert.EqualValuesf(t, &value, metric1.Value, "ожидаемое значение метрики 12.5, получено %p", metric1.Value)

	delta := int64(12)
	_, err = storage.SetMetric("test_metric_counter2", "counter", nil, &delta)
	assert.NoError(t, err, "ожидалось отстутствие ошибки")

	metric2, err := storage.GetMetric("test_metric_counter2", "counter")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_counter2", metric2.ID, "ожидаемое имя метрики: 'test_metric_counter2', получено %s", metric2.ID)
	assert.Equalf(t, "counter", metric2.MType, "ожидаемый тип метрики 'counter', получено %s", metric2.MType)
	assert.EqualValuesf(t, &delta, metric2.Delta, "ожидаемое значение метрики 12, получено %d", *metric2.Delta)

	_, err = storage.GetMetric("non_existing", "gauge")
	assert.Error(t, err, "ожидалась ошибка получения несуществующей метрики")
}
