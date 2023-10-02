package services

import (
	"testing"

	"github.com/dbulyk/metrics-alerting-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCheckHashAndAddDelta(t *testing.T) {
	db, mock, _ := sqlmock.New()
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

	err := checkHashAndAddDelta(db, metric, "")
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

	err = checkHashAndAddDelta(db, metric, "")
	assert.Nil(t, err)

	assert.Equal(t, int64(15), *metric.Delta)
}
