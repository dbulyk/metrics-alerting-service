package services

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dbulyk/metrics-alerting-service/internal/fileio"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/stretchr/testify/require"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestSetMetric(t *testing.T) {
	storage := &fileRepository{}

	value := 12.5
	_, err := storage.Set(models.Metric{
		ID:    "test_metric_gauge",
		MType: "gauge",
		Delta: nil,
		Value: &value,
		Hash:  utils.Hash("test_metric_gauge:gauge:12.500000", config.GetKey()),
	})
	assert.NoError(t, err, "ожидалось отсутствие ошибки")
	delta := int64(12)
	_, err = storage.Set(models.Metric{
		ID:    "test_metric_counter",
		MType: "counter",
		Delta: &delta,
		Value: nil,
		Hash:  utils.Hash("test_metric_counter:counter:12", config.GetKey()),
	})
	assert.NoError(t, err, "ожидалось отсутствие ошибки")

	metrics, _ := storage.GetAll()
	assert.Lenf(t, metrics, 2, "ожидалось две метрики, получено %d", len(metrics))
	assert.Equalf(t, "test_metric_gauge", metrics[0].ID, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metrics[0].ID)
	assert.Equalf(t, "gauge", metrics[0].MType, "ожидаемый тип метрики 'gauge', получено %s", metrics[0].MType)
	assert.EqualValuesf(t, &value, metrics[0].Value, "ожидаемое значение метрики 12.5, получено %f", *metrics[0].Value)
	assert.EqualValuesf(t, &delta, metrics[1].Delta, "ожидаемое значение метрики 12, получено %d", *metrics[1].Delta)

	_, err = storage.Set(models.Metric{
		ID:    "test_metric_counter",
		MType: "counter",
		Delta: &delta,
		Value: nil,
		Hash:  "",
	})
	assert.NoErrorf(t, err, "ошибка обновления существующей метрики: %v", err)

	resultDelta := int64(24)
	metrics, _ = storage.GetAll()
	assert.EqualValuesf(t, &resultDelta, metrics[1].Delta, "ожидаемое значение метрики 24, получено %p", metrics[1].Delta)
}

func TestGetMetric(t *testing.T) {
	storage := &fileRepository{}

	value := 12.5
	_, err := storage.Set(models.Metric{
		ID:    "test_metric_gauge",
		MType: "gauge",
		Delta: nil,
		Value: &value,
		Hash:  utils.Hash("test_metric_gauge:gauge:12.500000", config.GetKey()),
	})
	assert.NoError(t, err, "ожидалось отсутствие ошибки")

	metric1, err := storage.Get("test_metric_gauge", "gauge")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_gauge", metric1.ID, "ожидаемое имя метрики: 'test_metric_gauge', получено %s", metric1.ID)
	assert.Equalf(t, "gauge", metric1.MType, "ожидаемый тип метрики 'gauge', получено %s", metric1.MType)
	assert.EqualValuesf(t, &value, metric1.Value, "ожидаемое значение метрики 12.5, получено %d", metric1.Value)

	delta := int64(12)
	_, err = storage.Set(models.Metric{
		ID:    "test_metric_counter2",
		MType: "counter",
		Delta: &delta,
		Value: nil,
		Hash:  utils.Hash("test_metric_counter2:counter:12", config.GetKey()),
	})
	assert.NoError(t, err, "ожидалось отсутствие ошибки")

	metric2, err := storage.Get("test_metric_counter2", "counter")
	assert.NoErrorf(t, err, "ошибка получения метрики: %v", err)
	assert.Equalf(t, "test_metric_counter2", metric2.ID, "ожидаемое имя метрики: 'test_metric_counter2', получено %s", metric2.ID)
	assert.Equalf(t, "counter", metric2.MType, "ожидаемый тип метрики 'counter', получено %s", metric2.MType)
	assert.EqualValuesf(t, &delta, metric2.Delta, "ожидаемое значение метрики 12, получено %d", *metric2.Delta)

	_, err = storage.Get("non_existing", "gauge")
	assert.Error(t, err, "ожидалась ошибка получения несуществующей метрики")
}

func TestNewConsumer(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	consumer, err := fileio.NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.Close()
}

func TestConsumer_Read(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	mem := NewFileRepository()
	v := 123.15
	_, err = mem.Set(models.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  utils.Hash("testGauge:gauge:123.150000", config.GetKey()),
	})
	assert.NoError(t, err)

	i := int64(123)
	_, err = mem.Set(models.Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: &i,
		Value: nil,
		Hash:  utils.Hash("testCounter:counter:123", config.GetKey()),
	})
	assert.NoError(t, err)

	testMetrics, _ := mem.GetAll()

	for _, metric := range testMetrics {
		data, err := json.Marshal(metric)
		require.NoError(t, err)
		_, err = tmpfile.Write(data)
		require.NoError(t, err)
		_, err = tmpfile.WriteString("\n")
		require.NoError(t, err)
	}

	consumer, err := fileio.NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.Close()

	mem1 := NewFileRepository()
	metrics, err := consumer.Read()
	require.NoError(t, err)

	for _, metric := range metrics {
		_, err = mem1.Set(metric)
		require.NoError(t, err)
	}

	metrics1, _ := mem1.GetAll()
	assert.Equal(t, testMetrics, metrics1)
}

func TestConsumer_Close(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	consumer, err := fileio.NewConsumer(tmpfile.Name())
	require.NoError(t, err)

	err = consumer.Close()
	assert.NoError(t, err)
}

func TestRestoreMetricsFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	mem := NewFileRepository()

	consumer, err := fileio.NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.Close()

	v := 123.15
	_, err = mem.Set(models.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  utils.Hash("testGauge:gauge:123.150000", config.GetKey()),
	})
	assert.NoError(t, err)

	i := int64(123)
	_, err = mem.Set(models.Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: &i,
		Value: nil,
		Hash:  utils.Hash("testCounter:counter:123", config.GetKey()),
	})
	assert.NoError(t, err)

	producer, err := fileio.NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.Close()

	err = producer.Save(mem, tmpfile.Name())
	require.NoError(t, err)

	err = consumer.Restore(mem)
	assert.NoError(t, err)

	metrics, _ := mem.GetAll()
	for _, expectedMetric := range metrics {
		metric, err := mem.Get(expectedMetric.ID, expectedMetric.MType)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetric, metric)
	}
}
