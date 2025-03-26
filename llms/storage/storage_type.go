package storage

type StorageType byte

const (
	Memory StorageType = iota
	File               // use bolt storage
	Redis              // not supported
)
