package main

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
}
