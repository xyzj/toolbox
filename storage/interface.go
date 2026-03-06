package storage

type Storage interface {
	Store(history ...string) error
	Load() ([]string, error)

	// Clear removes all stored conversation histories from the storage backend.
	// This operation is irreversible and should be used with caution.
	Clear() error
}
