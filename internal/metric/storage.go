package metric

type Repository interface {
	SetMetric(metric Metric, restore bool) (*Metric, error)
	GetMetric(mName string, mType string) (*Metric, error)
	ListMetrics() ([]*Metric, error)
}
