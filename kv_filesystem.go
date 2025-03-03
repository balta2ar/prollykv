package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FileSystem struct {
	dir string
}

var _ KV = &FileSystem{}

func NewKVFile() *FileSystem {
	this := &FileSystem{
		dir: filepath.Join(os.TempDir(), "prollykv"),
	}
	this.MustBaseDir()
	return this
}

func (kv *FileSystem) Get(key []byte) ([]byte, error) {
	path := filepath.Join(kv.dir, string(key))
	return os.ReadFile(path)
}

func (kv *FileSystem) Set(key []byte, value []byte) error {
	path := filepath.Join(kv.dir, string(key))
	fmt.Printf("Set %q %q\n", path, value)
	return os.WriteFile(path, value, 0644)
}

func (kv *FileSystem) MustCleanup() {
	err := os.RemoveAll(kv.dir)
	mustNil(err)
}

func (kv *FileSystem) MustBaseDir() {
	err := os.MkdirAll(kv.dir, 0755)
	mustNil(err)
}

func (kv *FileSystem) MustReset() {
	kv.MustCleanup()
	kv.MustBaseDir()
}

func (kv *FileSystem) Cursor() KVCursor {
	return &FileSystemCursor{
		dir: kv.dir,
	}
}

type FileSystemCursor struct {
	dir   string
	keys  []string
	index int
}

var _ KVCursor = &FileSystemCursor{}

func mustList(dir string) []string {
	entries, err := os.ReadDir(dir)
	mustNil(err)
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func (f *FileSystemCursor) Goto(key []byte) {
	f.keys = mustList(f.dir)
	sort.Strings(f.keys)
	s := string(key)
	i := sort.SearchStrings(f.keys, s)
	if i < len(f.keys) && f.keys[i] == s {
		f.index = i
	} else {
		f.index = len(f.keys)
	}
}

func (f *FileSystemCursor) Next() {
	if f.index < len(f.keys) {
		f.index++
	}
}

func (f *FileSystemCursor) Key() []byte {
	if f.index < len(f.keys) {
		return []byte(f.keys[f.index])
	}
	return nil
}

func (f *FileSystemCursor) Value() []byte {
	if f.index < len(f.keys) {
		path := filepath.Join(f.dir, f.keys[f.index])
		return mustSlurp(path)
	}
	return nil
}
