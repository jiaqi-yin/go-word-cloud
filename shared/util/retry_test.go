package util

import (
	"fmt"
	"testing"
	"time"
)

func TestRetryWithSuccess(t *testing.T) {
	operation := func() error {
		return nil
	}
	retry := NewRetry(1*time.Second, 2*time.Second, operation)
	got := retry.Do()
	if got != nil {
		t.Fatalf(`want %v, but got %v`, nil, got)
	}
}

func TestRetryWithOperationError(t *testing.T) {
	operation := func() error {
		return fmt.Errorf("Foo bar error")
	}
	retry := NewRetry(1*time.Second, 2*time.Second, operation)
	got := retry.Do()
	if got == nil {
		t.Fatalf(`want %v, but got %v`, "Timeout while trying operation", nil)
	}
}
