package main

import (
	"os"
	"path/filepath"
)

type KVFile struct {
	dir string
}

func NewKVFile() *KVFile {
	dir := filepath.Join(os.TempDir(), "prollykv")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	return &KVFile{
		dir: dir,
	}
}

func (kv *KVFile) Get(key []byte) ([]byte, error) {
	path := filepath.Join(kv.dir, string(key))
	return os.ReadFile(path)
}

func (kv *KVFile) Set(key []byte, value []byte) error {
	path := filepath.Join(kv.dir, string(key))
	return os.WriteFile(path, value, 0644)
}
