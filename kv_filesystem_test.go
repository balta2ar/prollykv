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
	require.Equal(t, []byte("21"), cursor.Value())

	cursor.Next()
	require.Equal(t, []byte("b1"), cursor.Key())
	require.Equal(t, []byte("22"), cursor.Value())

	cursor.Next()
	require.Equal(t, []byte("b2"), cursor.Key())
	require.Equal(t, []byte("23"), cursor.Value())

	cursor.Next()
	require.Equal(t, []byte("c"), cursor.Key())
	require.Equal(t, []byte("3"), cursor.Value())

	cursor.Next()
	require.Nil(t, cursor.Key())
	require.Nil(t, cursor.Value())
}
