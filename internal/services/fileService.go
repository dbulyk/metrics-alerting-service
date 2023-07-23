package services

import (
	"crypto/hmac"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/fileio"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/rs/zerolog/log"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrInvalidHash       = errors.New("входящий хэш не совпадает с вычисленным")
	ErrInvalidMetric     = errors.New("метрики с такими параметрами не существует")
	ErrInvalidMetricType = errors.New("такого типа метрик не существует")
)

type fileRepository struct {
	sync.Mutex
	metrics       []*models.Metric
	storeInterval time.Duration
	storeFile     string
}

func NewFileRepository() storages.Repository {
	return &fileRepository{
		metrics:       make([]*models.Metric, 0, 50),
		Mutex:         sync.Mutex{},
		storeFile:     config.GetStoreFile(),
		storeInterval: config.GetStoreInterval(),
	}
}

func (fr *fileRepository) GetAll() ([]*models.Metric, error) {
	fr.Lock()
	var listMetrics = make([]*models.Metric, len(fr.metrics))
	copy(listMetrics, fr.metrics)
	fr.Unlock()
	return listMetrics, nil
}

func (fr *fileRepository) Set(metric models.Metric) (*models.Metric, error) {
	fr.Lock()
	defer fr.Unlock()

	m, err := addToStorage(fr.metrics, metric)
	if err != nil {
		log.Error().Err(err).Msgf("произошла ошибка сохранения метрики %s, она не будет добавлена.", metric.ID)
		return nil, err
	}
	fr.metrics = m

	if len(fr.storeFile) != 0 && fr.storeInterval == 0 {
		producer, err := fileio.NewProducer(fr.storeFile)
		if err != nil {
			return nil, err
		}
		err = producer.Save(fr, fr.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("ошибка сохранения метрики в файл")
			return nil, err
		}
	}

	return &metric, nil
}

func (fr *fileRepository) Get(id string, mType string) (*models.Metric, error) {
	fr.Lock()
	for _, m := range fr.metrics {
		if m.ID == id && m.MType == mType {
			fr.Unlock()
			return m, nil
		}
	}
	fr.Unlock()
	return nil, ErrInvalidMetric
}

func (fr *fileRepository) Updates(metrics []models.Metric) error {
	fr.Lock()
	defer fr.Unlock()

	for _, metric := range metrics {
		_, err := addToStorage(fr.metrics, metric)
		if err != nil {
			log.Error().Err(err).Msgf("произошла ошибка сохранения метрики %s, она не будет добавлена. "+
				"Ошибка: %s", metric.ID, err)
			continue
		}
	}

	if len(fr.storeFile) != 0 && fr.storeInterval == 0 {
		producer, err := fileio.NewProducer(fr.storeFile)
		if err != nil {
			return err
		}
		err = producer.Save(fr, fr.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("ошибка сохранения метрики в файл")
			return err
		}
	}
	return nil
}

func (fr *fileRepository) Ping() error {
	return nil
}

func addToStorage(metrics []*models.Metric, metric models.Metric) ([]*models.Metric, error) {
	if metric.MType != Counter && metric.MType != Gauge {
		log.Error().Msgf("типа метрики %s не существует", metric.MType)
		return nil, ErrInvalidMetricType
	}

	var mHash, s string
	key := config.GetKey()
	if len(key) > 0 {
		switch metric.MType {
		case Gauge:
			s = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
		case Counter:
			s = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
		}

		mHash = utils.Hash(s, key)
		if !hmac.Equal([]byte(mHash), []byte(metric.Hash)) {
			log.Error().Msgf("входящий хэш не совпадает с вычисленным. Метрика %s не будет добавлена", metric.ID)
			return nil, ErrInvalidHash
		}
	}

	isNotFound := true
	for _, m := range metrics {
		if m.ID == metric.ID && m.MType == metric.MType {
			isNotFound = false
			if m.MType == Counter {
				d := *m.Delta + *metric.Delta
				m.Delta = &d
				if len(key) > 0 {
					s = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta)
					mHash = utils.Hash(s, key)
				}
			} else {
				m.Value = metric.Value
			}
			m.Hash = mHash
		}
	}

	if isNotFound {
		metrics = append(metrics, &metric)
	}

	return metrics, nil
}
