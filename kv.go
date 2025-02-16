package main

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Cursor() Cursor
}

// Cursor can iterate key-value storage in ascending order.
type Cursor interface {
	Goto(key []byte)
	Next()
	Key() []byte
	Value() []byte
}
