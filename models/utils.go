package models

import (
	"errors"
	"time"
)

func Retryer(f func() error, retryableError ...error) error {
	// 1.2...3.....4x
	tryStep := 1
	for tryStep <= MaxErrRetryCount {
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

func RetryerCon(f func() error, isRetryable func(error) bool) error {
	// 1.2...3.....4x
	tryStep := 1
	for tryStep <= MaxErrRetryCount {
		err := f()
		if err == nil {
			return nil
		}
		if !isRetryable(err) || tryStep == 4 {
			return err
		}
		time.Sleep(time.Duration(tryStep*2-1) * time.Second)
		tryStep++
	}

	return nil
}
