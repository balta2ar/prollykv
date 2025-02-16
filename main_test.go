package main

import (
	"testing"
	"sort"
)

func mapIter(m map[string]string) Iter {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return func(cb func(key []byte, value []byte)) {
		for _, k := range keys {
			cb([]byte(k), []byte(m[k]))
		}
	}
}

func TestSmoke(t *testing.T) {
	kv := NewKVFile()
	tree := NewTree(kv)
	files := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	tree.Build(mapIter(files))
}
