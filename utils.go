package main

import (
	"crypto/sha256"
	"os"
)

func mustNil(err error) {
	if err != nil {
		panic(err)
	}
}

func mustSlurp(path string) []byte {
	data, err := os.ReadFile(path)
	mustNil(err)
	return data
}

func rehash(xs ...[]byte) []byte {
	h := sha256.New()
	for _, x := range xs {
		h.Write(x)
	}
	return h.Sum(nil)
}
