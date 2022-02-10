package util

import (
	"context"
	"fmt"
	"time"
)

type Retry struct {
	sleep     time.Duration
	timeout   time.Duration
	operation func() error
}

func (r *Retry) Do() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	for {
		err = r.operation()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("Timeout while trying operation: %w", err)
		case <-time.After(r.sleep):
		}
	}
}

func NewRetry(sleep, timeout time.Duration, operation func() error) *Retry {
	return &Retry{
		sleep:     sleep,
		timeout:   timeout,
		operation: operation,
	}
}
