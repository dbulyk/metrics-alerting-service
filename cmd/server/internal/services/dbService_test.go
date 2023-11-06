package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	del := int64(1)
	metric := models.Metric{ID: "testCounter", MType: Counter, Delta: &del}
	dr := dbRepository{db: db}

	rows := sqlmock.
		NewRows([]string{"delta"}).
		AddRow(0)

	mock.ExpectQuery("select (.+)").WithArgs(metric.ID, metric.MType).WillReturnRows(rows)

	mock.ExpectExec("insert (.+)").
		WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	metricResp, err := dr.Set(ctx, metric)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), *metricResp.Delta)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCheckHashAndAddDelta(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	delta := int64(5)
	metric := &models.Metric{
		MType: Counter,
		ID:    "test",
		Delta: &delta,
		Hash:  "",
	}

	rows := sqlmock.
		NewRows([]string{"delta"}).
		AddRow(0)

	mock.ExpectQuery("select (.+)").WithArgs(metric.ID, metric.MType).WillReturnRows(rows)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = checkHashAndAddDelta(ctx, db, metric, "")
	assert.Nil(t, err)

	assert.Equal(t, int64(5), *metric.Delta)

	delta = int64(10)
	metric = &models.Metric{
		MType: Counter,
		ID:    "test",
		Delta: &delta,
		Hash:  "",
	}

	rows = sqlmock.
		NewRows([]string{"delta"}).
		AddRow(5)

	mock.ExpectQuery("select (.+)").WithArgs(metric.ID, metric.MType).WillReturnRows(rows)

	err = checkHashAndAddDelta(ctx, db, metric, "")
	assert.NoError(t, err)

	assert.Equal(t, int64(15), *metric.Delta)

	metric.Hash = utils.Hash("test:counter:15", "test")
	mock.ExpectQuery("select (.+)").WithArgs(metric.ID, metric.MType).WillReturnRows(rows)
	err = checkHashAndAddDelta(ctx, db, metric, "test")
	assert.NoError(t, err)

	err = checkHashAndAddDelta(ctx, db, metric, "test1")
	assert.Error(t, err, ErrInvalidHash)
}

func TestGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	dr := &dbRepository{db: db}

	del := int64(123)
	mockMetric := &models.Metric{
		ID:    "testCounter",
		MType: Counter,
		Delta: &del,
		Value: nil,
		Hash:  "",
	}

	rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value", "hash"}).
		AddRow(mockMetric.ID, mockMetric.MType, mockMetric.Delta, mockMetric.Value, mockMetric.Hash)
	mock.ExpectQuery("^select (.+) from metrics where (.+)").
		WithArgs(mockMetric.ID, mockMetric.MType).
		WillReturnRows(rows)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m, err := dr.Get(ctx, mockMetric.ID, mockMetric.MType)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, mockMetric.ID, m.ID)
	assert.Equal(t, mockMetric.MType, m.MType)
	assert.Equal(t, mockMetric.Delta, m.Delta)

	mock.ExpectQuery("^select (.+) from metrics").WillReturnError(errors.New("mock error"))
	_, err = dr.Get(ctx, "testCounterWrong", Counter)
	assert.Error(t, err, ErrInvalidMetric)
}

func TestDBRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	del := int64(12)
	val := 2.2
	rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
		AddRow(1, Gauge, &del, nil).
		AddRow(2, Counter, nil, &val)

	mock.ExpectQuery("^select (.+) from metrics order by id$").WillReturnRows(rows)

	repo := &dbRepository{db, ""}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	metrics, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	assert.Equal(t, "1", metrics[0].ID)
	assert.Equal(t, Gauge, metrics[0].MType)
	assert.Equal(t, &del, metrics[0].Delta)

	assert.Equal(t, "2", metrics[1].ID)
	assert.Equal(t, Counter, metrics[1].MType)
	assert.Equal(t, &val, metrics[1].Value)

	mock.ExpectQuery("^select (.+) from metrics order by id$").
		WillReturnError(errors.New("unexpected error"))

	_, err = repo.GetAll(ctx)

	assert.Error(t, err)
	assert.Equal(t, "unexpected error", err.Error())

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
