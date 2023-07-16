package metric

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dbulyk/metrics-alerting-service/config"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbRepository struct {
	db *pgxpool.Pool
}

func NewDBRepository(db *pgxpool.Pool) Repository {
	_, err := db.Exec(context.Background(), "create table if not exists metrics (id text primary key, mtype text not null, delta bigint, value double precision, hash text)")
	if err != nil {
		log.Panic().Timestamp().Err(err).Msg("ошибка создания таблицы метрик")
	}

	return &dbRepository{
		db: db,
	}
}

func (d *dbRepository) Set(metric Metric) (*Metric, error) {
	log.Info().Msgf("добавление метрики %s", metric.ID)
	var mHash, key, s string
	key = config.GetKey()
	if len(key) > 0 {
		switch metric.MType {
		case gauge:
			s = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
		case counter:
			s = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
		}

		mHash = utils.Hash(s, key)
		if !hmac.Equal([]byte(mHash), []byte(metric.Hash)) {
			log.Error().Msgf("входящий хэш не совпадает с вычисленным. Метрика %s не будет добавлена", metric.ID)
			return nil, ErrInvalidHash
		}
	}

	if metric.MType == counter {
		res := d.db.QueryRow(context.Background(), "select delta from metrics where id = $1 and mtype = $2", metric.ID, metric.MType)
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
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		}
	}

	_, err := d.db.Exec(context.Background(),
		"insert into metrics(id, mtype, delta, value, hash) values($1, $2, $3, $4, $5) on conflict (id) do update set delta = $3, value = $4, hash = $5",
		metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		log.Info().Msgf("ошибка записи метрики в базу данных: %s", err)
		return nil, err
	}
	return &metric, nil
}

func (d *dbRepository) Get(mName string, mType string) (*Metric, error) {
	rows := d.db.QueryRow(context.Background(),
		"select id, mtype, delta, value, hash from metrics where id = $1 and mtype = $2", mName, mType)
	var m Metric
	err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err != nil {
		log.Info().Msgf("ошибка сканирования метрики из базы данных: %s", err)
		return nil, ErrInvalidMetric
	}
	return &m, nil
}

func (d *dbRepository) GetAll() ([]*Metric, error) {
	var metrics []*Metric

	rows, err := d.db.Query(context.Background(), "select id, mtype, delta, value from metrics order by id")
	if err != nil {
		log.Info().Msgf("ошибка получения метрик из базы данных: %s", err)
		return nil, err
	}
	for rows.Next() {
		var m Metric
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
		if err != nil {
			log.Info().Msgf("ошибка сканирования метрик из базы данных: %s", err)
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (d *dbRepository) Ping() error {
	return d.db.Ping(context.Background())
}
