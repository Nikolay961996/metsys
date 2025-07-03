package repositories

type MetricDto struct {
	Name  string
	Type  string
	Value string
}

type Storage interface {
	SetGauge(metricName string, value float64)
	GetGauge(metricName string) (float64, error)
	AddCounter(metricName string, value int64)
	GetCounter(metricName string) (int64, error)
	GetAll() []MetricDto
	TryFlushToFile()
}
