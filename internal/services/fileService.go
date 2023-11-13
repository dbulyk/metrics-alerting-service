package services

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/fileio"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/rs/zerolog/log"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrInvalidHash       = errors.New("incoming hash does not match the calculated hash")
	ErrInvalidMetric     = errors.New("there is no such metric")
	ErrInvalidMetricType = errors.New("this type of metric doesn't exist")
)

type fileRepository struct {
	sync.Mutex
	metrics       []*models.Metric
	storeInterval time.Duration
	storeFile     string
	key           string
}

func NewFileRepository(storeFile string, storeInterval time.Duration, key string) storages.Repository {
	return &fileRepository{
		metrics:       make([]*models.Metric, 0, 50),
		Mutex:         sync.Mutex{},
		storeFile:     storeFile,
		storeInterval: storeInterval,
		key:           key,
	}
}

func (fr *fileRepository) GetAll(_ context.Context) ([]*models.Metric, error) {
	fr.Lock()
	var listMetrics = make([]*models.Metric, len(fr.metrics))
	copy(listMetrics, fr.metrics)
	fr.Unlock()
	return listMetrics, nil
}

func (fr *fileRepository) Set(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	fr.Lock()
	defer fr.Unlock()

	m, err := addToStorage(&fr.metrics, metric, fr.key)
	if err != nil {
		log.Error().Err(err).Msgf("an error occurred in saving metric %s, it will not be added", metric.ID)
		return nil, err
	}
	fr.metrics = m

	if len(fr.storeFile) != 0 && fr.storeInterval == 0 {
		producer, err := fileio.NewProducer(fr.storeFile)
		if err != nil {
			return nil, err
		}
		err = producer.Save(ctx, fr, fr.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("error saving metrics to file")
			return nil, err
		}
	}

	return &metric, nil
}

func (fr *fileRepository) Get(_ context.Context, id string, mType string) (*models.Metric, error) {
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

func (fr *fileRepository) Updates(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	fr.Lock()
	defer fr.Unlock()

	for _, metric := range metrics {
		_, err := addToStorage(&fr.metrics, metric, fr.key)
		if err != nil {
			log.Error().Err(err).Msgf("an error occurred in saving metric %s, it will not be added", metric.ID)
			continue
		}
	}

	if len(fr.storeFile) != 0 && fr.storeInterval == 0 {
		producer, err := fileio.NewProducer(fr.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("producer creation error")
			return nil, err
		}
		err = producer.Save(ctx, fr, fr.storeFile)
		if err != nil {
			log.Error().Err(err).Msg("error saving metrics to file")
			return nil, err
		}
	}
	return metrics, nil
}

func (fr *fileRepository) Ping() error {
	return nil
}

func addToStorage(metrics *[]*models.Metric, metric models.Metric, key string) ([]*models.Metric, error) {
	if metric.MType != Counter && metric.MType != Gauge {
		log.Error().Msgf("like metric %s doesn't exist", metric.MType)
		return nil, ErrInvalidMetricType
	}

	var mHash, s string
	if len(key) > 0 {
		switch metric.MType {
		case Gauge:
			s = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
		case Counter:
			s = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
		}

		mHash = utils.Hash(s, key)
		if !hmac.Equal([]byte(mHash), []byte(metric.Hash)) {
			log.Error().Msgf("the incoming hash does not match the calculated hash. Metric %s will not be added", metric.ID)
			return nil, ErrInvalidHash
		}
	}

	isNotFound := true
	for _, m := range *metrics {
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
		*metrics = append(*metrics, &metric)
	}

	return *metrics, nil
}
