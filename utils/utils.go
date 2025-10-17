// Package models functions for retries and other utils
package utils

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Nikolay961996/metsys/models"
)

// Retryer special function for expenecial back-off with error types
func Retryer(f func() error, retryableError ...error) error {
	// 1.2...3.....4x
	tryStep := 1
	for tryStep <= models.MaxErrRetryCount {
		models.Log.Error(fmt.Sprintf("Unknown flags: %v\n", flag.Args()))
		err := f()
		if err == nil {
			return nil
		}

		ok := false
		for _, e := range retryableError {
			if errors.Is(err, e) {
				ok = true
				break
			}
		}

		if !ok || tryStep == 4 {
			return err
		}

		time.Sleep(time.Duration(tryStep*2-1) * time.Second)

		tryStep++
	}

	return nil
}

// RetryerCon special function for expenecial back-off with check function.
func RetryerCon(f func() error, isRetryable func(error) bool) error {
	// 1.2...3.....4x
	tryStep := 1
	for tryStep <= models.MaxErrRetryCount {
		err := f()
		if err == nil {
			return nil
		}
		models.Log.Warn(fmt.Sprintf("retry step: %d", tryStep))
		if !isRetryable(err) || tryStep == 4 {
			return err
		}
		time.Sleep(time.Duration(tryStep*2-1) * time.Second)
		tryStep++
	}

	return nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
