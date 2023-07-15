package metric

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsumer(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	consumer, err := NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.file.Close()

	assert.NotNil(t, consumer.file)
	assert.NotNil(t, consumer.reader)
	assert.NotNil(t, consumer.decoder)

	_, err = consumer.file.Stat()
	assert.NoError(t, err)
}

func TestConsumer_Read(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	mem := NewFileRepository()
	v := 123.15
	_, err = mem.Set(Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  utils.Hash("testGauge:gauge:123.150000", config.GetKey()),
	})
	assert.NoError(t, err)

	i := int64(123)
	_, err = mem.Set(Metric{
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

	consumer, err := NewConsumer(tmpfile.Name())
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

	consumer, err := NewConsumer(tmpfile.Name())
	require.NoError(t, err)

	err = consumer.Close()
	assert.NoError(t, err)

	_, err = consumer.file.Stat()
	assert.Error(t, err)
}

func TestRestoreMetricsFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	mem := NewFileRepository()

	consumer, err := NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.Close()

	v := 123.15
	_, err = mem.Set(Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  utils.Hash("testGauge:gauge:123.150000", config.GetKey()),
	})
	assert.NoError(t, err)

	i := int64(123)
	_, err = mem.Set(Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: &i,
		Value: nil,
		Hash:  utils.Hash("testCounter:counter:123", config.GetKey()),
	})
	assert.NoError(t, err)

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.file.Close()

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
