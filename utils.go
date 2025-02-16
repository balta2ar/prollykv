package main

import "os"

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
