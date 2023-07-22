package storages

import "github.com/dbulyk/metrics-alerting-service/internal/models"

type Repository interface {
	Set(metric models.Metric) (*models.Metric, error)
	Get(mName string, mType string) (*models.Metric, error)
	GetAll() ([]*models.Metric, error)
	Ping() error
}
