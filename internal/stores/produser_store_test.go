package stores

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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

	mem := NewMemStorage()
	v := 1.05
	_, err = mem.SetMetric("testGauge", "gauge", &v, nil)
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric("testCounter", "counter", nil, &i)
	assert.NoError(t, err)

	metrics, _ := mem.ListMetrics()

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
