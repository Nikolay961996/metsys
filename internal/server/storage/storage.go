package storage

import (
	"encoding/json"
	"errors"
	"github.com/Nikolay961996/metsys/models"
	"os"
	"strconv"
)

type MetricDto struct {
	Name  string
	Type  string
	Value string
}

type MemStorage struct {
	savesFile string
	syncSave  bool

	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewMemStorage(savesFile string, syncSave bool, restore bool) *MemStorage {
	s := MemStorage{
		savesFile:      savesFile,
		syncSave:       syncSave,
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	if restore {
		d, err := os.ReadFile(savesFile)
		if err != nil {
			models.Log.Error(err.Error())
			return &s
		}
		err = json.Unmarshal(d, &s)
		if err != nil {
			models.Log.Error(err.Error())
			return &s
		}
	}

	return &s
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.GaugeMetrics[metricName] = value
	if m.syncSave {
		m.TryFlushToFile()
	}
}

func (m *MemStorage) GetGauge(metricName string) (float64, error) {
	value, ok := m.GaugeMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *MemStorage) AddCounter(metricName string, value int64) {
	m.CounterMetrics[metricName] += value
	if m.syncSave {
		m.TryFlushToFile()
	}
}

func (m *MemStorage) GetCounter(metricName string) (int64, error) {
	value, ok := m.CounterMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *MemStorage) GetAll() []MetricDto {
	var r []MetricDto
	for k, v := range m.GaugeMetrics {
		r = append(r, MetricDto{
			Name:  k,
			Type:  models.Gauge,
			Value: strconv.FormatFloat(v, 'f', -1, 64),
		})
	}
	for k, v := range m.CounterMetrics {
		r = append(r, MetricDto{
			Name:  k,
			Type:  models.Counter,
			Value: strconv.FormatInt(v, 10),
		})
	}
	return r
}

func (m *MemStorage) TryFlushToFile() {
	models.Log.Info("Metrics try save")
	d, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		models.Log.Error(err.Error())
		return
	}
	err = os.WriteFile(m.savesFile, d, 0666)
	if err != nil {
		models.Log.Error(err.Error())
		return
	}
	models.Log.Info("Save success")
}
