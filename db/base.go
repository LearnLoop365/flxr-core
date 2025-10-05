package db

type Client[T any] interface {
	Init() error
	Close() error
	DBHandle() T // generic handle
}
