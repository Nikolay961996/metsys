package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	databaseDSN string
	db          *sql.DB
}

func NewDBStorage(databaseDSN string) *DBStorage {
	s := DBStorage{}
	s.open(databaseDSN)
	s.migrate()

	return &s
}

/*
func (m *DBStorage) SetGauge(metricName string, value float64) {
	m.GaugeMetrics[metricName] = value
	if m.syncSave {
		m.TryFlushToFile()
	}

}

func (m *DBStorage) GetGauge(metricName string) (float64, error) {
	value, ok := m.GaugeMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *DBStorage) AddCounter(metricName string, value int64) {
	m.CounterMetrics[metricName] += value
	if m.syncSave {
		m.TryFlushToFile()
	}
}

func (m *DBStorage) GetCounter(metricName string) (int64, error) {
	value, ok := m.CounterMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *DBStorage) GetAll() []MetricDto {
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

func (m *DBStorage) TryFlushToFile() {}
*/

func (m *DBStorage) migrate() {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		models.Log.Fatal(fmt.Sprintf("migration driver creation error: %s", err.Error()))
		panic(err)
	}

	instance, err := migrate.NewWithDatabaseInstance("file://internal/server/migrations", m.databaseDSN, driver)
	if err != nil {
		models.Log.Fatal(fmt.Sprintf("migration instance creation error: %s", err.Error()))
		panic(err)
	}

	if err := instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		models.Log.Fatal(fmt.Sprintf("migration instance up error: %s", err.Error()))
		panic(err)
	}
}

func (m *DBStorage) open(databaseDSN string) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		panic(err)
	}
	m.databaseDSN = databaseDSN
	m.db = db
}
