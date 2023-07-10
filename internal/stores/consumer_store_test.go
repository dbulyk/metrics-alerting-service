package stores

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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

	mem := NewMemStorage()
	v := 1.05
	_, err = mem.SetMetric("testGauge", "gauge", &v, nil, "")
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric("testCounter", "counter", nil, &i, "")
	assert.NoError(t, err)

	testMetrics, _ := mem.ListMetrics()

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

	mem1 := NewMemStorage()
	metrics, err := consumer.Read()
	require.NoError(t, err)

	for _, metric := range metrics {
		_, err := mem1.SetMetric(metric.ID, metric.MType, metric.Value, metric.Delta, "")
		require.NoError(t, err)
	}

	metrics1, _ := mem1.ListMetrics()
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

	mem := NewMemStorage()

	consumer, err := NewConsumer(tmpfile.Name())
	require.NoError(t, err)
	defer consumer.Close()

	v := 1.05
	_, err = mem.SetMetric("testGauge", "gauge", &v, nil, "")
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric("testCounter", "counter", nil, &i, "")
	assert.NoError(t, err)

	tmpfile, err = os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.file.Close()

	err = producer.Save(mem, tmpfile.Name())
	require.NoError(t, err)

	err = consumer.Restore(mem)
	assert.NoError(t, err)

	metrics, _ := mem.ListMetrics()
	for _, expectedMetric := range metrics {
		metric, err := mem.GetMetric(expectedMetric.ID, expectedMetric.MType)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetric, metric)
	}
}
