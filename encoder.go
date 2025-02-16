package main

import "fmt"

type Entry struct {
	Key   []byte
	Value []byte
}

func EncodeKey(level int8, key []byte) []byte {
	sLevel := fmt.Sprintf("%02d", level)
	return append([]byte(sLevel), key...)
}

func EncodeValue(hash []byte, value []byte) []byte {
	return append(hash, value...)
}

func DecodeKey(data []byte) (int8, []byte) {
	level := int8(0)
	_, err := fmt.Sscanf(string(data[:2]), "%d", &level)
	mustNil(err)
	return level, data[2:]
}

func DecodeValue(data []byte) ([]byte, []byte) {
	return data[:HashSize], data[HashSize:]
}

func Encode(node *Node) []byte {
	value := append(node.hash, node.value...)
	return append(EncodeKey(node.level, node.key), value...)
}
