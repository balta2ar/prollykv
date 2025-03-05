package main

import "fmt"

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

func StrEncodeKey(level int8, key string) string {
	sLevel := fmt.Sprintf("%02d", level)
	return sLevel + key
}

func StrDecodeKey(data string) (int8, string) {
	level := int8(0)
	_, err := fmt.Sscanf(data[:2], "%d", &level)
	mustNil(err)
	return level, data[2:]
}

func StrEncodeValue(hash string, value string) string {
	return hash[:HashSize] + value // TODO: remove later
}

func StrDecodeValue(data string) (string, string) {
	return data[:HashSize], data[HashSize:]
}
