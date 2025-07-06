package storage

import (
	"context"
	"encoding/json"
	"github.com/Nikolay961996/metsys/models"
	"os"
	"time"
)

type FileStorage struct {
	*MemStorage

	isSyncSave    bool
	saveTimer     *time.Ticker
	savesFilePath string
}

func NewFileStorage(savesFile string, savePeriod time.Duration, restore bool) *FileStorage {
	s := FileStorage{
		MemStorage:    NewMemStorage(),
		savesFilePath: savesFile,
		isSyncSave:    savePeriod == 0,
	}

	if !s.isSyncSave {
		s.saveTimer = time.NewTicker(savePeriod)
		go s.backgroundSaver()
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
	if m.isSyncSave {
		m.tryFlushToFile()
	}
}

func (m *FileStorage) AddCounter(metricName string, value int64) {
	m.MemStorage.AddCounter(metricName, value)
	if m.isSyncSave {
		m.tryFlushToFile()
	}
}

func (m *FileStorage) Close() {
	defer m.saveTimer.Stop()
}

func (m *FileStorage) PingContext(_ context.Context) error {
	return nil
}

func (m *FileStorage) backgroundSaver() {
	for range m.saveTimer.C {
		m.tryFlushToFile()
	}
}

func (m *FileStorage) StartTransaction(_ context.Context) error {
	return nil
}
func (m *FileStorage) CommitTransaction() error {
	return nil
}

func (m *FileStorage) tryFlushToFile() {
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
