package cache

import (
	"container/ring"
	"sync"

	"github.com/xyzj/toolbox/json"
)

// New creates a new History instance with the specified context size.
// The context size determines how many messages can be stored in the circular buffer.
// When the buffer is full, new messages will overwrite the oldest messages.
//
// Parameters:
//   - context: Maximum number of messages to store in the history buffer
//
// Returns a new History instance ready for use.
func NewRing[T any](max int) *Ring[T] {
	return &Ring[T]{
		locker: sync.RWMutex{},
		data:   ring.New(max),
		max:    max,
	}
}

// Ring implements a circular buffer for storing chat completion messages.
// It provides efficient storage and retrieval of conversation history with
// automatic memory management through ring buffer overflow handling.
//
// The Ring struct ensures:
//   - Fixed memory footprint regardless of conversation length
//   - Preservation of most recent messages when capacity is exceeded
//   - Thread-safe operations for concurrent access patterns
//   - JSON serialization support for persistence
type Ring[T any] struct {
	locker sync.RWMutex
	data   *ring.Ring // Circular buffer storing the messages
	max    int        // Maximum context size (currently unused, kept for future use)
}

// Store adds a single message to the history buffer.
// If the buffer is full, the oldest message will be overwritten.
// Always returns true for consistency with interface expectations.
//
// Parameters:
//   - msg: The chat completion message to store
//
// Returns true to indicate successful storage.
func (u *Ring[T]) Store(data T) bool {
	u.locker.Lock()
	u.data.Value = data
	u.data = u.data.Next()
	u.locker.Unlock()
	return true
}

// StoreMany adds multiple messages to the history buffer in sequence.
// Each message is stored using the same overflow behavior as Store().
// This is more efficient than calling Store() multiple times.
//
// Parameters:
//   - msgs: Variable number of chat completion messages to store
func (u *Ring[T]) StoreMany(msgs ...T) {
	u.locker.Lock()
	for _, msg := range msgs {
		u.data.Value = msg
		u.data = u.data.Next()
	}
	u.locker.Unlock()
}

// Clear removes all messages from the history buffer by setting all
// ring elements to nil. The buffer structure remains intact and ready for new messages.
func (u *Ring[T]) Clear() {
	u.locker.Lock()
	u.data.Do(func(a any) {
		u.data.Value = nil
	})
	u.locker.Unlock()
}

// Len returns the capacity of the history buffer (not the number of stored messages).
// This represents the maximum number of messages that can be stored.
func (u *Ring[T]) Len() int {
	u.locker.RLock()
	l := u.data.Len()
	u.locker.RUnlock()
	return l
}

// Slice returns all non-nil messages from the history buffer as a slice.
// Messages are returned in the order they were stored, with nil entries filtered out.
// This is the primary method for retrieving the conversation history.
//
// Returns:
//   - []T: Slice of stored messages in chronological order
func (u *Ring[T]) Slice() []T {
	u.locker.RLock()
	x := make([]T, 0, u.data.Len())
	u.data.Do(func(a any) {
		if a == nil {
			return
		}
		x = append(x, a.(T))
	})
	u.locker.RUnlock()
	return x
}

// MarshalJSON implements the json.Marshaler interface for the History type.
// It serializes the history as a JSON array of chat completion messages.
//
// Returns:
//   - []byte: JSON representation of the message history
//   - error: Any error encountered during marshaling
func (u *Ring[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Slice())
}

// ToJSON converts the history to a JSON string representation.
// Returns an empty string if marshaling fails.
//
// Returns:
//   - string: JSON string representation of the message history, or empty string on error
func (u *Ring[T]) ToJSON() string {
	b, err := json.Marshal(u.Slice())
	if err != nil {
		return ""
	}
	return json.String(b)
}

// FromJSON populates the history from a JSON string representation.
// The existing history is cleared before loading the new messages.
//
// Parameters:
//   - s: JSON string containing an array of chat completion messages
//
// Returns:
//   - error: Any error encountered during unmarshaling or invalid JSON format
func (u *Ring[T]) FromJSON(s string) error {
	a := make([]T, 0)
	err := json.Unmarshal(json.Bytes(s), &a)
	if err != nil {
		return err
	}
	u.StoreMany(a...)
	return nil
}
