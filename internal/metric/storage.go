package metric

type Repository interface {
	Set(metric Metric) (*Metric, error)
	Get(mName string, mType string) (*Metric, error)
	GetAll() ([]*Metric, error)
}
