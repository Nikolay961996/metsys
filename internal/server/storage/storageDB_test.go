package storage

import (
	"errors"
	"net"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// TestShouldRetryDBError тестирует логику определения повторных попыток
func TestShouldRetryDBError(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "Network timeout error",
			err:      &net.DNSError{IsTimeout: true},
			expected: true,
		},
		{
			name:     "Network non-timeout error",
			err:      &net.DNSError{IsTimeout: false},
			expected: false,
		},
		{
			name: "PostgreSQL admin shutdown",
			err: &pgconn.PgError{
				Code: pgerrcode.AdminShutdown,
			},
			expected: true,
		},
		{
			name: "PostgreSQL cannot connect now",
			err: &pgconn.PgError{
				Code: pgerrcode.CannotConnectNow,
			},
			expected: true,
		},
		{
			name: "PostgreSQL too many connections",
			err: &pgconn.PgError{
				Code: pgerrcode.TooManyConnections,
			},
			expected: true,
		},
		{
			name: "PostgreSQL connection exception",
			err: &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			},
			expected: true,
		},
		{
			name: "PostgreSQL serialization failure",
			err: &pgconn.PgError{
				Code: pgerrcode.SerializationFailure,
			},
			expected: true,
		},
		{
			name: "PostgreSQL other error",
			err: &pgconn.PgError{
				Code: "00000", // Successful completion
			},
			expected: false,
		},
		{
			name:     "Mixed error types",
			err:      errors.New("custom error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetryDBError(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetryDBError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
