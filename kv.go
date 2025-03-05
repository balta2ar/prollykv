package main

type KV interface {
	Get(key []byte) ([]byte, bool, error)
	Set(key []byte, value []byte) error
	Cursor() KVCursor
}

// KVCursor can iterate key-value storage in ascending order.
type KVCursor interface {
	Goto(key []byte)
	Next()
	Key() []byte
	Value() []byte
}
