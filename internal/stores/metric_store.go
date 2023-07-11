package stores

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/hashes"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/rs/zerolog/log"
)

const (
	counter = "counter"
	gauge   = "gauge"
)

type MetricStore interface {
	SetMetric(id string, mType string, value *float64, delta *int64) (*models.Metrics, error)
	GetMetric(mName string, mType string) (*models.Metrics, error)
	ListMetrics() ([]*models.Metrics, error)
}

type MemStorage struct {
	sync.Mutex
	metrics       []*models.Metrics
	storeInterval time.Duration
	storeFile     string
}

var (
	ErrInvalidHash       = errors.New("входящий хэш не совпадает с вычисленным")
	ErrInvalidMetric     = errors.New("метрики с такими параметрами не существует")
	ErrInvalidMetricType = errors.New("такого типа метрик не существует")
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics:       make([]*models.Metrics, 0, 50),
		Mutex:         sync.Mutex{},
		storeFile:     config.GetStoreFile(),
		storeInterval: config.GetStoreInterval(),
	}
}

func (ms *MemStorage) ListMetrics() ([]*models.Metrics, error) {
	ms.Lock()
	var listMetrics = make([]*models.Metrics, len(ms.metrics))
	copy(listMetrics, ms.metrics)
	ms.Unlock()
	return listMetrics, nil
}

func (ms *MemStorage) SetMetric(id string, mType string, value *float64, delta *int64, hash string) (*models.Metrics, error) {
	ms.Lock()
	defer ms.Unlock()

	if mType != counter && mType != gauge {
		log.Error().Msgf("тип метрики %s не существует", mType)
		return nil, ErrInvalidMetricType
	}

	key := config.GetKey()
	if len(key) != 0 {
		var newHash string
		switch mType {
		case gauge:
			newHash = fmt.Sprintf("%s:gauge:%f", id, *value)
		case counter:
			newHash = fmt.Sprintf("%s:counter:%d", id, *delta)
		}

		if !hashes.ValidHash(newHash, hash, key) {
			log.Error().Msgf("входящий хэш не совпадает с вычисленным. Метрика %s не будет добавлена", id)
			return nil, ErrInvalidHash
		}
	}

	for _, m := range ms.metrics {
		if m.ID == id && m.MType == mType {
			if m.MType == counter {
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
		Hash:  hash,
	}

	ms.metrics = append(ms.metrics, m)

	if len(ms.storeFile) != 0 && ms.storeInterval == 0 {
		producer, err := NewProducer(ms.storeFile)
		if err != nil {
			return nil, err
		}
		err = producer.Save(ms, ms.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("ошибка сохранения метрики в файл")
			return nil, err
		}
	}

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
	return nil, ErrInvalidMetric
}
