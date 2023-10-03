package services

import (
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/rs/zerolog/log"
)

type dbRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) storages.Repository {
	_, err := db.Exec("create table if not exists metrics (id text primary key, mtype text not null, delta bigint, value double precision, hash text)")
	if err != nil {
		log.Panic().Timestamp().Err(err).Msg("ошибка создания таблицы метрик")
	}

	return &dbRepository{
		db: db,
	}
}

func (dr *dbRepository) Set(metric models.Metric) (*models.Metric, error) {
	key := config.GetKey()
	err := checkHashAndAddDelta(dr.db, &metric, key)
	if err != nil {
		return nil, err
	}

	_, err = dr.db.Exec("insert into metrics(id, mtype, delta, value, hash) values($1, $2, $3, $4, $5) on conflict (id) do update set delta = $3, value = $4, hash = $5",
		metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		log.Error().Err(err).Msgf("ошибка записи метрики в базу данных")
		return nil, err
	}
	return &metric, nil
}

func (dr *dbRepository) Get(mName string, mType string) (*models.Metric, error) {
	rows := dr.db.QueryRow("select id, mtype, delta, value, hash from metrics where id = $1 and mtype = $2", mName, mType)
	var m models.Metric
	err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err != nil {
		log.Error().Err(err).Msg("ошибка сканирования метрики из базы данных")
		return nil, ErrInvalidMetric
	}

	if len(m.Hash) == 0 && len(config.GetKey()) > 0 {
		if m.MType == Gauge {
			m.Hash = utils.Hash(fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value), config.GetKey())
		} else {
			m.Hash = utils.Hash(fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta), config.GetKey())
		}
	}

	return &m, nil
}

func (dr *dbRepository) GetAll() ([]*models.Metric, error) {
	var metrics []*models.Metric

	rows, err := dr.db.Query("select id, mtype, delta, value from metrics order by id")
	if err != nil {
		log.Error().Err(err).Msg("ошибка получения метрик из базы данных")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Metric
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
		if err != nil {
			log.Error().Err(err).Msg("ошибка сканирования метрик из базы данных")
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (dr *dbRepository) Updates(metrics []models.Metric) ([]models.Metric, error) {
	key := config.GetKey()
	for i := range metrics {
		err := checkHashAndAddDelta(dr.db, &metrics[i], key)
		if err != nil {
			return nil, err
		}

		tx, err := dr.db.Begin()
		if err != nil {
			log.Error().Err(err).Msg("ошибка открытия транзакции: %")
			return nil, err
		}
		_, err = tx.Exec("insert into metrics(id, mtype, delta, value, hash) values($1, $2, $3, $4, $5) on conflict (id) do update set delta = $3, value = $4, hash = $5",
			metrics[i].ID, metrics[i].MType, metrics[i].Delta, metrics[i].Value, metrics[i].Hash)
		if err != nil {
			log.Error().Err(err).Msg("ошибка записи метрики в базу данных. Откатываем транзакцию")
			err = tx.Rollback()
			if err != nil {
				log.Error().Err(err).Msg("ошибка отката транзакции")
				return nil, err
			}
			continue
		}

		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Msg("ошибка коммита транзакции")
			return nil, err
		}
	}

	return metrics, nil
}

func (dr *dbRepository) Ping() error {
	return dr.db.Ping()
}

func checkHashAndAddDelta(db *sql.DB, metric *models.Metric, key string) error {
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
			log.Error().Msgf("входящий хэш не совпадает с вычисленным. Метрика %s не будет добавлена", metric.ID)
			return ErrInvalidHash
		}
	}

	if metric.MType == Counter {
		res := db.QueryRow("select delta from metrics where id = $1 and mtype = $2", metric.ID, metric.MType)
		if res != nil {
			var delta int64
			err := res.Scan(&delta)
			if err == nil {
				del := delta + *metric.Delta
				metric.Delta = &del
				if len(key) > 0 {
					s = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
					metric.Hash = utils.Hash(s, key)
				}
			} else if !errors.Is(err, sql.ErrNoRows) {
				log.Error().Err(err).Msg("ошибка сканирования метрики из базы данных")
				return err
			}
		}
	}
	return nil
}
