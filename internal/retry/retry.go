// Package retry provides retry helpers with backoff for transient failures.
package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type RetryConfig struct {
	Intervals   []time.Duration
	MaxAttempts int
}

func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 4,
		Intervals:   []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func ExecuteWithRetry[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var zero T
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 1
	}

	var executionError error
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}
		executionError = err

		if !IsRetriableError(err) {
			return zero, err
		}

		if attempt < config.MaxAttempts-1 {
			var interval time.Duration
			if len(config.Intervals) == 0 {
				interval = time.Second
			} else {
				idx := attempt
				if idx >= len(config.Intervals) {
					idx = len(config.Intervals) - 1
				}
				interval = config.Intervals[idx]
			}

			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(interval):
				continue
			}

		}

	}
	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, executionError)
}

func IsRetriableError(err error) bool {

	if err == nil {
		return false
	}

	if IsNetworkError(err) {
		return true
	}

	if IsPgConnectionError(err) || IsPgRetriableError(err) {
		return true
	}
	return false
}

func IsNetworkError(err error) bool {

	var netError net.Error
	if errors.As(err, &netError) {
		return true
	}

	var opError *net.OpError
	if errors.As(err, &opError) {
		return true
	}

	var syscallErr *net.OpError
	if errors.As(err, &syscallErr) {
		if errno, ok := syscallErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT:
				return true
			}
		}
	}

	return false
}

func IsPgConnectionError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) {
		return strings.HasPrefix(string(pqError.Code), "08")
	}

	return false
}

func IsPgRetriableError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure, pgerrcode.SQLClientUnableToEstablishSQLConnection, pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection, pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation:
			return true
		default:
			return strings.HasPrefix(string(pqError.Code), "08")
		}
	}
	return false
}