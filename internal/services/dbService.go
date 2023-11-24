package services

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/rs/zerolog/log"
)

type dbRepository struct {
	db  *sql.DB
	key string
}

// NewDBRepository creates a new repository for working with the database and returns a pointer to it.
// It also performs database migration.
func NewDBRepository(db *sql.DB, dsn string, key string) storages.IRepository {
	m, err := migrate.New(
		"file://db/migrations",
		dsn)
	if err != nil {
		log.Panic().Err(err).Msg("migrate creation error")
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Panic().Err(err).Msg("migrate up error")
	}

	return &dbRepository{
		db:  db,
		key: key,
	}
}

// Set adds a new metric to the database or updates an existing one, check hash and add delta to existing counter.
func (dr *dbRepository) Set(ctx context.Context, metric models.Metric) (*models.Metric, error) {
	err := checkHashAndAddDelta(ctx, dr.db, &metric, dr.key)
	if err != nil {
		log.Error().Err(err).Msg("error of hash verification and delta addition")
		return nil, err
	}

	_, err = dr.db.ExecContext(ctx,
		"insert into metrics(id, mtype, delta, value, hash) values($1, $2, $3, $4, $5) "+
			"on conflict (id) do update set delta = $3, value = $4, hash = $5",
		metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		log.Error().Err(err).Msg("error of writing metrics to the database")
		return nil, err
	}
	return &metric, nil
}

// Get returns a metric from the database by name and type and check hash.
func (dr *dbRepository) Get(ctx context.Context, mName string, mType string) (*models.Metric, error) {
	rows := dr.db.QueryRowContext(ctx, "select id, mtype, delta, value, hash from metrics "+
		"where id = $1 and mtype = $2", mName, mType)
	var m models.Metric
	err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err != nil {
		log.Error().Err(err).Msg("metric scanning error from database")
		return nil, ErrInvalidMetric
	}

	if len(dr.key) > 0 {
		if m.MType == Gauge {
			m.Hash = utils.Hash(fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value), dr.key)
		} else {
			m.Hash = utils.Hash(fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta), dr.key)
		}
	}

	return &m, nil
}

// GetAll returns all metrics from the database.
func (dr *dbRepository) GetAll(ctx context.Context) ([]*models.Metric, error) {
	var metrics []*models.Metric

	rows, err := dr.db.QueryContext(ctx, "select id, mtype, delta, value from metrics order by id")
	if err != nil {
		log.Error().Err(err).Msg("error of getting metrics from the database")
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Error().Err(err).Msg("error of closing rows")
		}
	}(rows)

	for rows.Next() {
		var m models.Metric
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
		if err != nil {
			log.Error().Err(err).Msg("error of scanning metrics from the database")
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("error of getting metrics from the database")
		return nil, err
	}

	return metrics, nil
}

// Updates adds a slice of metrics to the database or updates existing ones, check hash
// and add delta to existing counter.
func (dr *dbRepository) Updates(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	key := dr.key
	for i := range metrics {
		err := checkHashAndAddDelta(ctx, dr.db, &metrics[i], key)
		if err != nil {
			log.Error().Err(err).Msg("error of hash verification and delta addition")
			return nil, err
		}

		tx, err := dr.db.Begin()
		if err != nil {
			log.Error().Err(err).Msg("transaction opening error")
			return nil, err
		}
		_, err = tx.ExecContext(ctx, "insert into metrics(id, mtype, delta, value, hash) values($1, $2, $3, $4, $5) "+
			"on conflict (id) do update set delta = $3, value = $4, hash = $5",
			metrics[i].ID, metrics[i].MType, metrics[i].Delta, metrics[i].Value, metrics[i].Hash)
		if err != nil {
			log.Error().Err(err).Msg("error of writing the metric to the database. Roll back the transaction")
			err = tx.Rollback()
			if err != nil {
				log.Error().Err(err).Msg("transaction rollback error")
				return nil, err
			}
			continue
		}

		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Msg("transaction commit error")
			return nil, err
		}
	}

	return metrics, nil
}

// Ping checks the connection to the database.
func (dr *dbRepository) Ping() error {
	return dr.db.Ping()
}

func checkHashAndAddDelta(ctx context.Context, db *sql.DB, metric *models.Metric, key string) error {
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
			log.Error().Msgf("the incoming hash does not match the calculated hash. Metric %s will not be added",
				metric.ID)
			return ErrInvalidHash
		}
	}

	if metric.MType == Counter {
		res := db.QueryRowContext(ctx, "select delta from metrics where id = $1 and mtype = $2",
			metric.ID, metric.MType)
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
				log.Error().Err(err).Msg("metric scanning error from database")
				return err
			}
		}
	}
	return nil
}
