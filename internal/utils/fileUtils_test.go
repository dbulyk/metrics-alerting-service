package utils

import (
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSaveMetricsToFile(t *testing.T) {
	mem := stores.NewMemStorage()

	v := 1.05
	_, err := mem.SetMetric("testGauge", "gauge", &v, nil)
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric("testCounter", "counter", nil, &i)
	assert.NoError(t, err)

	tmpfile, err := os.CreateTemp("", "testfile.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = SaveMetrics(mem, tmpfile.Name())
	assert.NoError(t, err)
}

func TestRestoreMetricsFromFile(t *testing.T) {
	mem := stores.NewMemStorage()

	v := 1.05
	_, err := mem.SetMetric("testGauge", "gauge", &v, nil)
	assert.NoError(t, err)

	i := int64(2)
	_, err = mem.SetMetric("testCounter", "counter", nil, &i)
	assert.NoError(t, err)

	tmpfile, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = SaveMetrics(mem, tmpfile.Name())
	require.NoError(t, err)

	err = RestoreMetrics(mem, tmpfile.Name())
	assert.NoError(t, err)

	metrics, _ := mem.ListMetrics()
	for _, expectedMetric := range metrics {
		metric, err := mem.GetMetric(expectedMetric.ID, expectedMetric.MType)
		assert.NoError(t, err)
		assert.Equal(t, expectedMetric, metric)
	}
}
