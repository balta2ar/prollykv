package main

import (
	"fmt"
	"os"
)

func mustTrue(b bool, msg string, args ...interface{}) {
	if !b {
		panic(fmt.Sprintf(msg, args...))
	}
}

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
