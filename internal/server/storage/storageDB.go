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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net"
	"strconv"
)

type DBStorage struct {
	databaseDSN string
	db          *sql.DB
	tx          *sql.Tx

	sqlInsertOrUpdateGauge   *sql.Stmt
	sqlInsertOrUpdateCounter *sql.Stmt
	sqlGetGauge              *sql.Stmt
	sqlGetCounter            *sql.Stmt
	sqlGetAll                *sql.Stmt
}

func NewDBStorage(databaseDSN string) *DBStorage {
	s := DBStorage{}
	s.open(databaseDSN)
	s.migrate()
	s.prepareSQL()

	return &s
}

func (m *DBStorage) SetGauge(metricName string, value float64) {
	ctx := context.Background()
	err := models.RetryerCon(func() error {
		_, err := m.sqlInsertOrUpdateGauge.ExecContext(ctx, metricName, value)
		return err
	}, shouldRetryDBError)
	if err != nil {
		models.Log.Error(fmt.Sprintf("Failed to set for metric %s: %s", metricName, err.Error()))
	}
}

func (m *DBStorage) GetGauge(metricName string) (float64, error) {
	ctx := context.Background()
	var value float64
	err := models.RetryerCon(func() error {
		return m.sqlGetGauge.QueryRowContext(ctx, metricName).Scan(&value)
	}, shouldRetryDBError)
	return value, err
}

func (m *DBStorage) AddCounter(metricName string, value int64) {
	ctx := context.Background()
	err := models.RetryerCon(func() error {
		_, err := m.sqlInsertOrUpdateCounter.ExecContext(ctx, metricName, value)
		return err
	}, shouldRetryDBError)
	if err != nil {
		models.Log.Error(fmt.Sprintf("failed to set for metric %s: %s", metricName, err.Error()))
	}
}

func (m *DBStorage) GetCounter(metricName string) (int64, error) {
	ctx := context.Background()
	var delta int64
	err := models.RetryerCon(func() error {
		return m.sqlGetCounter.QueryRowContext(ctx, metricName).Scan(&delta)
	}, shouldRetryDBError)
	return delta, err
}

func (m *DBStorage) GetAll() []repositories.MetricDto {
	var r []repositories.MetricDto
	ctx := context.Background()
	var rows *sql.Rows

	err := models.RetryerCon(func() error {
		rs, err := m.sqlGetAll.QueryContext(ctx)
		if err != nil && rs.Err() != nil {
			rows = rs
		}
		return err
	}, shouldRetryDBError)

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
	migrateFunc := func() error {
		driver, err := postgres.WithInstance(m.db, &postgres.Config{})
		if err != nil {
			models.Log.Fatal(fmt.Sprintf("migration driver creation error: %s", err.Error()))
			return err
		}

		instance, err := migrate.NewWithDatabaseInstance("file://internal/server/migrations", m.databaseDSN, driver)
		if err != nil {
			models.Log.Fatal(fmt.Sprintf("migration instance creation error: %s", err.Error()))
			return err
		}

		if err := instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			models.Log.Fatal(fmt.Sprintf("migration instance up error: %s", err.Error()))
			return err
		}
		return nil
	}

	err := models.RetryerCon(migrateFunc, shouldRetryDBError)
	if err != nil {
		panic(err)
	}
}

func (m *DBStorage) open(databaseDSN string) {
	err := models.RetryerCon(
		func() error {
			db, err := sql.Open("pgx", databaseDSN)
			if err == nil {
				m.db = db
			}
			return err
		}, shouldRetryDBError)

	if err != nil {
		panic(err)
	}
	m.databaseDSN = databaseDSN
}

func (m *DBStorage) prepareSQL() {
	sqlInsertOrUpdateGauge, err := m.db.Prepare(
		`
		INSERT INTO metrics (id, type, value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (id, type) DO UPDATE 
		SET value = EXCLUDED.value;`)
	if err != nil {
		panic(err)
	}

	sqlInsertOrUpdateCounter, err := m.db.Prepare(
		`
		INSERT INTO metrics (id, type, delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (id, type) DO UPDATE 
		SET delta = EXCLUDED.delta + metrics.delta;`)
	if err != nil {
		panic(err)
	}

	sqlGetGauge, err := m.db.Prepare(`SELECT value FROM metrics WHERE id = $1 AND type = 'gauge'`)
	if err != nil {
		panic(err)
	}

	sqlGetCounter, err := m.db.Prepare(`SELECT delta FROM metrics WHERE id = $1 AND type = 'counter'`)
	if err != nil {
		panic(err)
	}

	sqlGetAll, err := m.db.Prepare(`SELECT id, type, value, delta FROM metrics`)
	if err != nil {
		panic(err)
	}

	m.sqlInsertOrUpdateGauge = sqlInsertOrUpdateGauge
	m.sqlInsertOrUpdateCounter = sqlInsertOrUpdateCounter
	m.sqlGetGauge = sqlGetGauge
	m.sqlGetCounter = sqlGetCounter
	m.sqlGetAll = sqlGetAll
}

func shouldRetryDBError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.AdminShutdown, // Сервер закрывает соединение
			pgerrcode.CannotConnectNow,     // Сервер не принимает подключения
			pgerrcode.TooManyConnections,   // Слишком много подключений
			pgerrcode.ConnectionException,  // Обрыв соединения
			pgerrcode.SerializationFailure: // Конфликт транзакций
			return true
		}
	}

	return false
}
