package queue

import "errors"

var (
	// ErrClosed is returned when trying to get an item from a closed queue.
	ErrClosed = errors.New("queue is closed")

	// ErrEmpty is returned when trying to remove an item from an empty queue.
	ErrEmpty = errors.New("queue is empty")

	// ErrFull is returned when trying to add an item to a full queue.
	ErrFull = errors.New("queue is full")

	// ErrTimeout indicates that an operation did not complete within the allowed time period.
	// It is returned when a queue operation or related IO exceeds its deadline or configured timeout.
	// Callers can compare against ErrTimeout to detect timeout conditions and implement retries or cancellation.
	ErrTimeout  = errors.New("operation timed out")
	ErrAbnormal = errors.New("operation ended abnormally")
)
