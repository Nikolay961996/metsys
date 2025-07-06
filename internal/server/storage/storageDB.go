package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strconv"
)

type DBStorage struct {
	databaseDSN string
	db          *sql.DB
	tx          *sql.Tx

	sqlInsertOrUpdate *sql.Stmt
	sqlGetValue       *sql.Stmt
	sqlGetDelta       *sql.Stmt
	sqlGetAll         *sql.Stmt
}

func NewDBStorage(databaseDSN string) *DBStorage {
	s := DBStorage{}
	s.open(databaseDSN)
	s.migrate()

	err := s.prepareSql()
	if err != nil {
		panic(err)
	}

	return &s
}

func (m *DBStorage) SetGauge(metricName string, value float64) {
	ctx := context.Background()
	_, err := m.sqlInsertOrUpdate.ExecContext(ctx, metricName, models.Gauge, value)
	if err != nil {
		models.Log.Error(fmt.Sprintf("Failed to set for metric %s: %s", metricName, err.Error()))
	}
}

func (m *DBStorage) GetGauge(metricName string) (float64, error) {
	ctx := context.Background()
	var value float64
	err := m.sqlGetValue.QueryRowContext(ctx, metricName, models.Gauge).
		Scan(&value)
	return value, err
}

func (m *DBStorage) AddCounter(metricName string, value int64) {
	ctx := context.Background()
	_, err := m.sqlInsertOrUpdate.ExecContext(ctx, metricName, models.Counter, value)
	if err != nil {
		models.Log.Error(fmt.Sprintf("Failed to set for metric %s: %s", metricName, err.Error()))
	}
}

func (m *DBStorage) GetCounter(metricName string) (int64, error) {
	ctx := context.Background()
	var delta int64
	err := m.sqlGetValue.QueryRowContext(ctx, metricName, models.Gauge).
		Scan(&delta)
	return delta, err
}

func (m *DBStorage) GetAll() []repositories.MetricDto {
	var r []repositories.MetricDto
	ctx := context.Background()
	rows, err := m.sqlGetAll.QueryContext(ctx)
	if err != nil {
		models.Log.Error(err.Error())
		return r
	}
	defer rows.Close()

	for rows.Next() {
		var m repositories.MetricDto
		var valueNull sql.NullFloat64
		var deltaNull sql.NullInt64

		err = rows.Scan(&m.Name, &m.Type, &valueNull, &deltaNull)
		if err != nil {
			models.Log.Error(err.Error())
			return r
		}
		if valueNull.Valid {
			m.Value = strconv.FormatFloat(valueNull.Float64, 'f', -1, 64)
		} else if deltaNull.Valid {
			m.Value = strconv.FormatInt(deltaNull.Int64, 10)
		}

		r = append(r, m)
	}

	err = rows.Err()
	if err != nil {
		models.Log.Error(err.Error())
	}

	return r
}

func (m *DBStorage) Close() {
	defer m.db.Close()
}

func (m *DBStorage) PingContext(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func (m *DBStorage) StartTransaction(ctx context.Context) error {
	if m.tx != nil {
		return errors.New("transaction already started")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	m.tx = tx

	return nil
}

func (m *DBStorage) CommitTransaction() error {
	if m.tx != nil {
		return m.tx.Commit()
	}
	return nil
}

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

func (m *DBStorage) prepareSql() error {
	sqlInsertOrUpdate, err := m.db.Prepare(
		`
		INSERT INTO metrics (id, type, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, type) DO UPDATE 
		SET value = EXCLUDED.value;`)
	if err != nil {
		return err
	}

	sqlGetValue, err := m.db.Prepare(
		`SELECT value FROM metrics WHERE id = $1 AND type = $2`)
	if err != nil {
		return err
	}

	sqlGetDelta, err := m.db.Prepare(
		`SELECT delta FROM metrics WHERE id = $1 AND type = $2`)
	if err != nil {
		return err
	}

	sqlGetAll, err := m.db.Prepare(
		`SELECT id, type, value, delta FROM metrics`)
	if err != nil {
		return err
	}

	m.sqlInsertOrUpdate = sqlInsertOrUpdate
	m.sqlGetValue = sqlGetValue
	m.sqlGetDelta = sqlGetDelta
	m.sqlGetAll = sqlGetAll

	return nil
}
