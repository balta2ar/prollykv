package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func generate1(n int) []*Message {
	m := []*Message{}
	for i := range n {
		i := i + 1
		data := fmt.Sprintf("value %d", i)
		m = append(m, &Message{timestamp: strconv.Itoa(i), data: data})
	}
	return m
}

func generate2(n int) []*Message {
	m := []*Message{}
	for i := range n {
		i := i + 1
		data := fmt.Sprintf("value2 %d", i)
		m = append(m, &Message{timestamp: strconv.Itoa(i), data: data})
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

func TestDiffSimple(t *testing.T) {
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

	{
		t1 := NewTree(generate1(2))
		t2 := NewTree(generate1(3))
		t1.Dot("t1.dot")
		t2.Dot("t2.dot")
		require.Len(t, Diff(t1, t2).Add, 1)
	}
	{
		t1 := NewTree(generate1(1))
		t2 := NewTree(generate1(51))
		require.Len(t, Diff(t1, t2).Add, 50)
	}
	{
		t1 := NewTree(generate1(30)[10:])
		t2 := NewTree(generate1(30))
		require.Len(t, Diff(t1, t2).Add, 10)
	}
	{
		t1 := NewTree(generate1(30)[:20])
		t2 := NewTree(generate1(30))
		require.Len(t, Diff(t1, t2).Add, 10)
	}
	{
		t1 := NewTree(generate1(30)[10:20])
		t2 := NewTree(generate1(30))
		require.Len(t, Diff(t1, t2).Add, 20)
	}
	{
		g := generate1(100)
		t1 := NewTree(append(g[:10], append(g[20:30], append(g[40:50], append(g[60:70], g[80:90]...)...)...)...))
		t2 := NewTree(generate1(100))
		require.Len(t, Diff(t1, t2).Add, 50)
	}
	{
		t1 := NewTree(generate1(10))
		t2 := NewTree(generate2(10))
		d := Diff(t1, t2)
		require.Len(t, d.Add, 0)
		require.Len(t, d.Update, 10)
		require.Len(t, d.Remove, 0)
	}
	{
		t1 := NewTree(generate1(20))
		t2 := NewTree(generate2(30)[10:])
		d := Diff(t1, t2)
		require.Len(t, d.Add, 10)
		require.Len(t, d.Update, 10)
		require.Len(t, d.Remove, 10)
	}

	// fmt.Println(tree)
	// mustNil(tree.Build(files))
}

func pickN(xs []*Message, n int) []*Message {
	rand.Shuffle(len(xs), func(i, j int) {
		xs[i], xs[j] = xs[j], xs[i]
	})
	return xs[:n]
}

// func TestDiffPermute(t *testing.T) {
// 	remove := generate1(10)
// 	add := generate2(30)[20:]
// 	update := generate1(20)[10:]
// 	for nAdd := range 10 {
// 		for nUpdate := range 10 {
// 			for nRemove := range 10 {

// 			}
// 		}
// 	}
// }

func ListBasedDiff(a, b []*Message) (out DeltaTrio) {
	mapA := make(map[string]*Message)
	mapB := make(map[string]*Message)
	for _, msg := range a {
		mapA[msg.timestamp] = msg
	}
	for _, msg := range b {
		mapB[msg.timestamp] = msg
	}
	for ts, msgB := range mapB {
		if msgA, exists := mapA[ts]; exists {
			if msgA.data != msgB.data {
				out.Update = append(out.Update, Delta{})
			}
		} else {
			out.Add = append(out.Add, Delta{})
		}
	}
	for ts := range mapA {
		if _, exists := mapB[ts]; !exists {
			out.Remove = append(out.Remove, Delta{})
		}
	}
	return out
}

func TestSerializeLevel0(t *testing.T) {
	t1 := NewTree(generate1(10))
	kv := NewKVFile()
	kv.MustReset()
	require.Nil(t, t1.SerializeLevel0(kv))
	t2, err := DeserializeLevel0(kv)
	require.Nil(t, err)
	d := Diff(t1, t2)
	require.Len(t, d.Add, 0)
	require.Len(t, d.Update, 0)
	require.Len(t, d.Remove, 0)
}

func TestSerializeWithKids(t *testing.T) {
	t1 := NewTree(generate1(10))
	kv := NewKVFile()
	kv.MustReset()
	gen := 42
	require.Nil(t, t1.SerializeWithKids(gen, kv))
	t2, err := DeserializeWithKids(gen, kv)
	require.Nil(t, err)
	d := Diff(t1, t2)
	require.Len(t, d.Add, 0)
	require.Len(t, d.Update, 0)
	require.Len(t, d.Remove, 0)
}

func TestSerializeJSON(t *testing.T) {
	t1 := NewTree(generate1(10))
	kv := NewKVFile()
	kv.MustReset()
	gen := 42
	file, err := os.Create(filepath.Join(kv.dir, fmt.Sprintf("gen-%d.json", gen)))
	require.Nil(t, err)
	defer file.Close()
	require.Nil(t, t1.SerializeJSON(gen, file))
}

func MustDirSize(path string) int {
	var totalSize int
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += int(info.Size())
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return totalSize
}

func TestSizeSerializeJSON(t *testing.T) {
	if os.Getenv("SIZE") == "" {
		t.Skipf("set SIZE=1 to run this test")
	}
	kv := NewKVFile()
	kv.MustReset()
	for gen := range 1000 {
		t1 := NewTree(generate1(gen))
		file, err := os.Create(filepath.Join(kv.dir, fmt.Sprintf("gen-%d.json", gen)))
		require.Nil(t, err)
		defer file.Close()
		require.Nil(t, t1.SerializeJSON(gen, file))
		fmt.Printf("%d,json,%d,%d\n", gen, MustDirSize(kv.dir), t1.Height())
	}
}

func TestSizeSerializeWithKids(t *testing.T) {
	if os.Getenv("SIZE") == "" {
		t.Skipf("set SIZE=1 to run this test")
	}
	kv := NewKVFile()
	kv.MustReset()
	for gen := range 1000 {
		t1 := NewTree(generate1(gen))
		require.Nil(t, t1.SerializeWithKids(gen, kv))
		fmt.Printf("%d,prolly,%d,%d\n", gen, MustDirSize(kv.dir), t1.Height())
	}
}
