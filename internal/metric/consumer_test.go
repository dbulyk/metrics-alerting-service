package metric

//
//import (
//	"encoding/json"
//	"github.com/dbulyk/metrics-alerting-service/internal/metric"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"os"
//	"testing"
//)
//
//func TestNewConsumer(t *testing.T) {
//	tmpDir := t.TempDir()
//
//	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
//	require.NoError(t, err)
//
//	consumer, err := NewConsumer(tmpfile.Name())
//	require.NoError(t, err)
//	defer consumer.file.Close()
//
//	assert.NotNil(t, consumer.file)
//	assert.NotNil(t, consumer.reader)
//	assert.NotNil(t, consumer.decoder)
//
//	_, err = consumer.file.Stat()
//	assert.NoError(t, err)
//}
//
//func TestConsumer_Read(t *testing.T) {
//	tmpDir := t.TempDir()
//
//	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
//	require.NoError(t, err)
//
//	mem := metric.NewRepository()
//	v := 1.05
//	_, err = mem.SetMetric(metric.Metric{
//		ID:    "testGauge",
//		MType: "gauge",
//		Delta: nil,
//		Value: &v,
//		Hash:  "",
//	}, true)
//	assert.NoError(t, err)
//
//	i := int64(2)
//	_, err = mem.SetMetric(metric.Metric{
//		ID:    "testCounter",
//		MType: "counter",
//		Delta: &i,
//		Value: nil,
//		Hash:  "",
//	}, true)
//	assert.NoError(t, err)
//
//	testMetrics, _ := mem.ListMetrics()
//
//	for _, metric := range testMetrics {
//		data, err := json.Marshal(metric)
//		require.NoError(t, err)
//		_, err = tmpfile.Write(data)
//		require.NoError(t, err)
//		_, err = tmpfile.WriteString("\n")
//		require.NoError(t, err)
//	}
//
//	consumer, err := NewConsumer(tmpfile.Name())
//	require.NoError(t, err)
//	defer consumer.Close()
//
//	mem1 := metric.NewRepository()
//	metrics, err := consumer.Read()
//	require.NoError(t, err)
//
//	for _, metric := range metrics {
//		_, err := mem1.SetMetric(metric, true)
//		require.NoError(t, err)
//	}
//
//	metrics1, _ := mem1.ListMetrics()
//	assert.Equal(t, testMetrics, metrics1)
//}
//
//func TestConsumer_Close(t *testing.T) {
//	tmpDir := t.TempDir()
//
//	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
//	require.NoError(t, err)
//
//	consumer, err := NewConsumer(tmpfile.Name())
//	require.NoError(t, err)
//
//	err = consumer.Close()
//	assert.NoError(t, err)
//
//	_, err = consumer.file.Stat()
//	assert.Error(t, err)
//}
//
//func TestRestoreMetricsFromFile(t *testing.T) {
//	tmpDir := t.TempDir()
//
//	tmpfile, err := os.CreateTemp(tmpDir, "*.json")
//	require.NoError(t, err)
//
//	mem := metric.NewRepository()
//
//	consumer, err := NewConsumer(tmpfile.Name())
//	require.NoError(t, err)
//	defer consumer.Close()
//
//	v := 1.05
//	_, err = mem.SetMetric(metric.Metric{
//		ID:    "testGauge",
//		MType: "gauge",
//		Delta: nil,
//		Value: &v,
//		Hash:  "",
//	}, true)
//	assert.NoError(t, err)
//
//	i := int64(2)
//	_, err = mem.SetMetric(metric.Metric{
//		ID:    "testCounter",
//		MType: "counter",
//		Delta: &i,
//		Value: nil,
//		Hash:  "",
//	}, true)
//	assert.NoError(t, err)
//
//	tmpfile, err = os.CreateTemp("", "testfile")
//	require.NoError(t, err)
//	defer os.Remove(tmpfile.Name())
//
//	producer, err := NewProducer(tmpfile.Name())
//	require.NoError(t, err)
//	defer producer.file.Close()
//
//	err = producer.Save(mem, tmpfile.Name())
//	require.NoError(t, err)
//
//	err = consumer.Restore(mem)
//	assert.NoError(t, err)
//
//	metrics, _ := mem.ListMetrics()
//	for _, expectedMetric := range metrics {
//		metric, err := mem.GetMetric(expectedMetric.ID, expectedMetric.MType)
//		assert.NoError(t, err)
//		assert.Equal(t, expectedMetric, metric)
//	}
//}
