package storage

import (
	"encoding/json"
	"github.com/Nikolay961996/metsys/models"
	"os"
)

type FileStorage struct {
	*MemStorage

	savesFilePath string
	syncSave      bool
}

func NewFileStorage(savesFile string, syncSave bool, restore bool) *FileStorage {
	m := NewMemStorage()
	s := FileStorage{
		MemStorage:    m,
		savesFilePath: savesFile,
		syncSave:      syncSave,
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

func (m *FileStorage) SetGauge(metricName string, value float64) {
	m.MemStorage.SetGauge(metricName, value)
	if m.syncSave {
		m.TryFlushToFile()
	}
}

//func (m *FileStorage) GetGauge(metricName string) (float64, error) {
//	return m.MemStorage.GetGauge(metricName)
//}

func (m *FileStorage) AddCounter(metricName string, value int64) {
	m.MemStorage.AddCounter(metricName, value)
	if m.syncSave {
		m.TryFlushToFile()
	}
}

//func (m *FileStorage) GetCounter(metricName string) (int64, error) {
//	value, ok := m.CounterMetrics[metricName]
//	if !ok {
//		return 0, errors.New("not Found")
//	}
//	return value, nil
//}

//func (m *FileStorage) GetAll() []repositories.MetricDto {
//	var r []repositories.MetricDto
//	for k, v := range m.GaugeMetrics {
//		r = append(r, repositories.MetricDto{
//			Name:  k,
//			Type:  models.Gauge,
//			Value: strconv.FormatFloat(v, 'f', -1, 64),
//		})
//	}
//	for k, v := range m.CounterMetrics {
//		r = append(r, repositories.MetricDto{
//			Name:  k,
//			Type:  models.Counter,
//			Value: strconv.FormatInt(v, 10),
//		})
//	}
//	return r
//}

func (m *FileStorage) TryFlushToFile() {
	models.Log.Info("Metrics try save")
	d, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		models.Log.Error(err.Error())
		return
	}
	err = os.WriteFile(m.savesFilePath, d, 0666)
	if err != nil {
		models.Log.Error(err.Error())
		return
	}
	models.Log.Info("Save success")
}
