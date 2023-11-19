package storages

import (
	"context"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
)

type IRepository interface {
	Set(ctx context.Context, metric models.Metric) (*models.Metric, error)
	Get(ctx context.Context, mName string, mType string) (*models.Metric, error)
	GetAll(ctx context.Context) ([]*models.Metric, error)
	Updates(ctx context.Context, metric []models.Metric) ([]models.Metric, error)
	Ping() error
}
