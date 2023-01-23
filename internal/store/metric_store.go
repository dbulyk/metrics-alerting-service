package store

import (
	"errors"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
)

var metrics []*models.Metric

type MetricStore interface {
	SetMetric(mName string, mType string, mValue float64) error
	GetMetric(mName string, mType string) (*models.Metric, error)
	ListMetrics() ([]*models.Metric, error)
}

type MemStorage struct {
	MetricStore
}

func (s *MemStorage) ListMetrics() ([]*models.Metric, error) {
	return metrics, nil
}

func (s *MemStorage) SetMetric(mName string, mType string, mValue float64) error {
	var value interface{}

	switch mType {
	case "gauge":
		value = models.Gauge(mValue)
	case "counter":
		value = models.Counter(mValue)
	default:
		return errors.New("такого типа метрики не существует")
	}

	for _, m := range metrics {
		if m.Name == mName && m.Type == mType {
			m.Value = value
			return nil
		}
	}

	metrics = append(metrics, &models.Metric{
		Name:  mName,
		Type:  mType,
		Value: value,
	})
	return nil
}

func (s *MemStorage) GetMetric(mName string, mType string) (*models.Metric, error) {
	for _, m := range metrics {
		if m.Name == mName && m.Type == mType {
			return m, nil
		}
	}
	return nil, errors.New("метрики с такими параметрами не существует")
}
