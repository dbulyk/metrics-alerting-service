package metric

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProducer(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.file.Close()

	assert.NotNil(t, producer.file)
	assert.NotNil(t, producer.encoder)

	_, err = producer.file.Stat()
	assert.NoError(t, err)
}

func TestProducer_Write(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	mem := NewRepository(db)
	v := 1.05
	_, err = mem.SetMetric(Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  "",
	}, true)
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric(Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: &i,
		Value: nil,
		Hash:  "",
	}, true)
	assert.NoError(t, err)

	metrics, _ := mem.ListMetrics()
	assert.Len(t, metrics, 2)

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.file.Close()

	err = producer.Write(metrics)
	assert.NoError(t, err)
}

func TestProducer_Close(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)

	err = producer.Close()
	assert.NoError(t, err)

	_, err = producer.file.Stat()
	assert.Error(t, err)
}

func TestSaveMetricsToFile(t *testing.T) {
	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
	require.NoError(t, err)

	producer, err := NewProducer(tmpfile.Name())
	require.NoError(t, err)
	defer producer.file.Close()

	mem := NewRepository(db)

	v := 1.05
	_, err = mem.SetMetric(Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: &v,
		Hash:  "",
	}, true)
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric(Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: &i,
		Value: nil,
		Hash:  "",
	}, true)
	assert.NoError(t, err)

	tmpfile, err = os.CreateTemp("", "testfile.json")
	require.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		assert.NoError(t, err)
	}(tmpfile.Name())

	err = producer.Save(mem, tmpfile.Name())
	assert.NoError(t, err)
}
