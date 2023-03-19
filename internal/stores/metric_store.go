package stores

import (
	"errors"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"sync"
)

type MetricStore interface {
	SetMetric(id string, mType string, value *float64, delta *int64) (*models.Metrics, error)
	GetMetric(mName string, mType string) (*models.Metrics, error)
	ListMetrics() ([]*models.Metrics, error)
}

type MemStorage struct {
	sync.Mutex
	metrics []*models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make([]*models.Metrics, 0, 50),
		Mutex:   sync.Mutex{},
	}
}

func (ms *MemStorage) ListMetrics() ([]*models.Metrics, error) {
	ms.Lock()
	var listMetrics = make([]*models.Metrics, len(ms.metrics))
	copy(listMetrics, ms.metrics)
	ms.Unlock()
	return listMetrics, nil
}

func (ms *MemStorage) SetMetric(id string, mType string, value *float64, delta *int64) (*models.Metrics, error) {
	ms.Lock()
	defer ms.Unlock()

	if mType != "counter" && mType != "gauge" {
		return nil, errors.New("такого типа метрик не существует")
	}

	for _, m := range ms.metrics {
		if m.ID == id && m.MType == mType {
			if m.MType == "counter" {
				d := *m.Delta + *delta
				m.Delta = &d
			} else {
				m.Value = value
			}
			return m, nil
		}
	}

	m := &models.Metrics{
		ID:    id,
		MType: mType,
		Value: value,
		Delta: delta,
	}

	ms.metrics = append(ms.metrics, m)
	return m, nil
}

func (ms *MemStorage) GetMetric(id string, mType string) (*models.Metrics, error) {
	ms.Lock()
	for _, m := range ms.metrics {
		if m.ID == id && m.MType == mType {
			ms.Unlock()
			return m, nil
		}
	}
	ms.Unlock()
	return nil, errors.New("метрики с такими параметрами не существует")
}
