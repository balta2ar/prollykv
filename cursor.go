package main

// Cursor can iterate key-value storage in ascending order.
type Cursor interface {
	Goto(key []byte)
	Next()
	Key() []byte
	Value() []byte
}
