package main

import (
	"fmt"
	"testing"
)

// func mapIter(m map[string]string) Iter {
// 	keys := make([]string, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	return func(cb func(key []byte, value []byte) error) error {
// 		for _, k := range keys {
// 			err := cb([]byte(k), []byte(m[k]))
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}
// }

func generate() []*Message {
	m := []*Message{}
	n := 1000
	for i := 0; i < n; i++ {
		data := fmt.Sprintf("value %d", i)
		m = append(m, &Message{timestamp: i, data: data})
	}
	return m
}

func TestBuild(t *testing.T) {

	messages := generate()
	kv := NewKVFile()
	kv.MustReset()
	tree := NewTree(kv, messages)
	fmt.Println(tree)
	// mustNil(tree.Build(files))
}
