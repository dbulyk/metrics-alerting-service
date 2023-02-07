package store

import (
	"errors"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"sync"
)

var metrics []*models.Metric

type MetricStore interface {
	SetMetric(mName string, mType string, mValue float64) error
	GetMetric(mName string, mType string) (*models.Metric, error)
	ListMetrics() ([]*models.Metric, error)
}

type MemStorage struct {
	sync.Mutex
	MetricStore
}

func (ms *MemStorage) ListMetrics() ([]*models.Metric, error) {
	return metrics, nil
}

func (ms *MemStorage) SetMetric(mName string, mType string, mValue float64) error {
	ms.Lock()
	defer ms.Unlock()
	for _, m := range metrics {
		if m.Name == mName && m.Type == mType {
			if m.Type == "counter" {
				m.Value = m.Value.(models.Counter) + models.Counter(mValue)
			} else {
				m.Value = models.Gauge(mValue)
			}
			return nil
		}
	}

	var value interface{}

	switch mType {
	case "gauge":
		value = models.Gauge(mValue)
	case "counter":
		value = models.Counter(mValue)
	default:
		return errors.New("такого типа метрики не существует")
	}

	metrics = append(metrics, &models.Metric{
		Name:  mName,
		Type:  mType,
		Value: value,
	})
	return nil
}

func (ms *MemStorage) GetMetric(mName string, mType string) (*models.Metric, error) {
	ms.Lock()
	for _, m := range metrics {
		if m.Name == mName && m.Type == mType {
			ms.Unlock()
			return m, nil
		}
	}
	ms.Unlock()
	return nil, errors.New("метрики с такими параметрами не существует")
}
