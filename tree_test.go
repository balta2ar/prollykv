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

func generate(n int) []*Message {
	m := []*Message{}
	// n := 1000
	for i := range n {
		data := fmt.Sprintf("value %d", i)
		m = append(m, &Message{timestamp: i, data: data})
	}
	return m
}

func TestBuild(t *testing.T) {
	messages := generate(10)
	kv := NewKVFile()
	kv.MustReset()
	tree := NewTree(messages)
	fmt.Println(tree)
	// mustNil(tree.Build(files))
}

func TestDiff(t *testing.T) {
	{
		all := generate(20)
		t := NewTree(all)
		t.Dot("all.dot")
		t1 := NewTree(all[:4])
		t2 := NewTree(all[1:3])
		t1.Dot("t1.dot")
		t2.Dot("t2.dot")
		fmt.Println(t2)
		fmt.Println(t1)
		// fmt.Println(Diff(t1, t2))
		fmt.Println(Diff(t2, t1))
	}
	// {
	// 	t1 := NewTree(generate(3))
	// 	t2 := NewTree(generate(2))
	// 	fmt.Println(Diff(t1, t2))
	// }
	// fmt.Println(tree)
	// mustNil(tree.Build(files))
}
