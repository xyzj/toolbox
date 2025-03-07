package storage

type StorageType byte

const (
	Memory StorageType = iota
	File               // not supported
	Redis              // not supported
)
