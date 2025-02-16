package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCursor(t *testing.T) {
	kv := NewKVFile()
	kv.MustReset()
	mustNil(kv.Set([]byte("a"), []byte("1")))
	mustNil(kv.Set([]byte("b"), []byte("21")))
	mustNil(kv.Set([]byte("b1"), []byte("22")))
	mustNil(kv.Set([]byte("b2"), []byte("23")))
	mustNil(kv.Set([]byte("c"), []byte("3")))

	cursor := kv.Cursor()
	cursor.Goto([]byte("b"))
	require.Equal(t, []byte("b"), cursor.Key())
}
