package main

import (
	"fmt"
	"sort"
	"testing"
)

func mapIter(m map[string]string) Iter {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return func(cb func(key []byte, value []byte) error) error {
		for _, k := range keys {
			err := cb([]byte(k), []byte(m[k]))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func generate() map[string]string {
	m := make(map[string]string)
	n := 1000
	w := len(fmt.Sprint(n))
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("%0*d", w, i)] = fmt.Sprintf("value %d", i)
	}
	return m
}

func TestBuild(t *testing.T) {
	kv := NewKVFile()
	kv.MustReset()
	tree := NewTree(kv)
	files := generate()
	mustNil(tree.Build(mapIter(files)))
}
