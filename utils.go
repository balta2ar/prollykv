package main

import (
	"crypto/sha256"
	"encoding/hex"
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

func Rehash(xs ...string) string {
	h := sha256.New()
	for _, x := range xs {
		h.Write([]byte(x))
	}
	return hex.EncodeToString(h.Sum(nil))
}
