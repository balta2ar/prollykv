package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generate1(n int) []*Message {
	m := []*Message{}
	for i := range n {
		data := fmt.Sprintf("value %d", i)
		m = append(m, &Message{timestamp: i, data: data})
	}
	return m
}

func generate2(n int) []*Message {
	m := []*Message{}
	for i := range n {
		data := fmt.Sprintf("value2 %d", i)
		m = append(m, &Message{timestamp: i, data: data})
	}
	return m
}

func TestBuild(t *testing.T) {
	messages := generate1(10)
	kv := NewKVFile()
	kv.MustReset()
	tree := NewTree(messages)
	fmt.Println(tree)
	// mustNil(tree.Build(files))
}

func TestDiff(t *testing.T) {
	// {
	// 	all := generate(20)
	// 	t := NewTree(all)
	// 	t.Dot("all.dot")
	// 	t1 := NewTree(all)
	// 	// t2 := NewTree(slices.Delete(slices.Delete(all, 3, 4), 14, 15))
	// 	t2 := NewTree(slices.Delete(slices.Delete(all, 3, 18), 0, 1))
	// 	t1.Dot("t1.dot")
	// 	t2.Dot("t2.dot")
	// 	fmt.Println(t2)
	// 	fmt.Println(t1)
	// 	// fmt.Println(Diff(t1, t2))
	// 	fmt.Println(Diff(t2, t1))
	// }

	// {
	// 	t1 := NewTree(generate1(2))
	// 	t2 := NewTree(generate1(3))
	// 	t1.Dot("t1.dot")
	// 	t2.Dot("t2.dot")
	// 	assert.Len(t, Diff(t1, t2).Add, 1)
	// }
	// {
	// 	t1 := NewTree(generate1(1))
	// 	t2 := NewTree(generate1(51))
	// 	assert.Len(t, Diff(t1, t2).Add, 50)
	// }
	// {
	// 	t1 := NewTree(generate1(30)[10:])
	// 	t2 := NewTree(generate1(30))
	// 	assert.Len(t, Diff(t1, t2).Add, 10)
	// }
	// {
	// 	t1 := NewTree(generate1(30)[:20])
	// 	t2 := NewTree(generate1(30))
	// 	assert.Len(t, Diff(t1, t2).Add, 10)
	// }
	// {
	// 	t1 := NewTree(generate1(30)[10:20])
	// 	t2 := NewTree(generate1(30))
	// 	assert.Len(t, Diff(t1, t2).Add, 20)
	// }
	// {
	// 	g := generate1(100)
	// 	t1 := NewTree(append(g[:10], append(g[20:30], append(g[40:50], append(g[60:70], g[80:90]...)...)...)...))
	// 	t2 := NewTree(generate1(100))
	// 	assert.Len(t, Diff(t1, t2).Add, 50)
	// }
	// {
	// 	t1 := NewTree(generate1(10))
	// 	t2 := NewTree(generate2(10))
	// 	d := Diff(t1, t2)
	// 	assert.Len(t, d.Add, 0)
	// 	assert.Len(t, d.Update, 10)
	// 	assert.Len(t, d.Remove, 0)
	// }
	{
		t1 := NewTree(generate1(20))
		t2 := NewTree(generate2(30)[10:])
		d := Diff(t1, t2)
		assert.Len(t, d.Add, 10)
		assert.Len(t, d.Update, 10)
		assert.Len(t, d.Remove, 10)
	}

	// fmt.Println(tree)
	// mustNil(tree.Build(files))
}
