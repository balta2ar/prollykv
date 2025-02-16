package main

import (
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

func TestBuild(t *testing.T) {
	kv := NewKVFile()
	kv.MustReset()
	tree := NewTree(kv)
	files := map[string]string{
		"a": "The cat sleeps",
		"b": "Birds fly high",
		"c": "Wind blows softly",
	}
	mustNil(tree.Build(mapIter(files)))
}
