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
	MaxAttempts int
	Intervals   []time.Duration
}

func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 4,
		Intervals:   []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

type RetryableFunc func() error

func ExecuteWithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var executionError error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}
		executionError = err

		if !IsRetriableError(err) {
			return err
		}

		if attempt < config.MaxAttempts-1 {
			interval := config.Intervals[0]
			if attempt < len(config.Intervals) {
				interval = config.Intervals[attempt]
			} else {
				interval = config.Intervals[len(config.Intervals)-1]
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
				continue
			}

		}

	}
	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, executionError)
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
