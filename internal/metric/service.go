package metric

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/rs/zerolog/log"
)

const (
	counter = "counter"
	gauge   = "gauge"
)

var (
	ErrInvalidHash       = errors.New("входящий хэш не совпадает с вычисленным")
	ErrInvalidMetric     = errors.New("метрики с такими параметрами не существует")
	ErrInvalidMetricType = errors.New("такого типа метрик не существует")
)

type repository struct {
	sync.Mutex
	metrics       []*Metric
	storeInterval time.Duration
	storeFile     string
	db            *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		metrics:       make([]*Metric, 0, 50),
		Mutex:         sync.Mutex{},
		storeFile:     config.GetStoreFile(),
		storeInterval: config.GetStoreInterval(),
		db:            db,
	}
}

func (ms *repository) ListMetrics() ([]*Metric, error) {
	ms.Lock()
	var listMetrics = make([]*Metric, len(ms.metrics))
	copy(listMetrics, ms.metrics)
	ms.Unlock()
	return listMetrics, nil
}

func (ms *repository) SetMetric(metric Metric, restore bool) (*Metric, error) {
	ms.Lock()
	defer ms.Unlock()

	if metric.MType != counter && metric.MType != gauge {
		log.Error().Msgf("типа метрики %s не существует", metric.MType)
		return nil, ErrInvalidMetricType
	}

	var mHash, s string
	key := config.GetKey()
	if len(key) != 0 && !restore {
		switch metric.MType {
		case gauge:
			log.Info().Msgf("получено значение метрики %s:%s:%f", metric.ID, metric.MType, *metric.Value)
			s = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
		case counter:
			log.Info().Msgf("получено значение метрики %s:%s:%d", metric.ID, metric.MType, *metric.Delta)
			s = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
		}

		mHash = utils.Hash(s, key)
		if !hmac.Equal([]byte(mHash), []byte(metric.Hash)) {
			log.Error().Msgf("входящий хэш не совпадает с вычисленным. Метрика %s не будет добавлена", metric.ID)
			return nil, ErrInvalidHash
		}
	}

	for _, m := range ms.metrics {
		if m.ID == metric.ID && m.MType == metric.MType {
			if m.MType == counter {
				d := *m.Delta + *metric.Delta
				m.Delta = &d
				s = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta) //TODO: подумать над тем, чтобы не дублировать код
				mHash = utils.Hash(s, key)
			} else {
				m.Value = metric.Value
			}
			m.Hash = mHash
			return m, nil
		}
	}

	ms.metrics = append(ms.metrics, &metric)

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

	return &metric, nil
}

func (ms *repository) GetMetric(id string, mType string) (*Metric, error) {
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

func (ms *repository) Ping() error {
	return ms.db.Ping(context.Background())
}