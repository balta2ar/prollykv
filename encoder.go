package main

import (
	"fmt"
	"strings"
)

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
	return hash[:HashSize] + value // TODO: remove later, use the whole hash
}

func StrDecodeValue(data string) (string, string) {
	return data[:HashSize], data[HashSize:]
}

func StrEncodeKeyWithKids(hash string) string {
	return hash[:HashSize]
}

func StrDecodeKeyWithKids(data string) string {
	return data[:HashSize]
}

func StrEncodeValueWithKids(level int8, kids []string, key int, data string) string {
	// level, nKids, kids, data
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%02d", level))
	sb.WriteString(fmt.Sprintf("%010d", key))
	sb.WriteString(fmt.Sprintf("%04d", len(kids)))
	for _, kid := range kids {
		sb.WriteString(kid[:HashSize])
	}
	sb.WriteString(data)
	return sb.String()
}

func StrDecodeValueWithKids(data string) (level int8, kids []string, key int, data_ string) {
	_, err := fmt.Sscanf(data[:2], "%d", &level)
	mustNil(err)
	_, err = fmt.Sscanf(data[2:12], "%d", &key)
	mustNil(err)
	nKids := 0
	_, err = fmt.Sscanf(data[12:16], "%d", &nKids)
	mustNil(err)
	kids = make([]string, nKids)
	offset := 16
	for i := range nKids {
		kids[i] = data[offset : offset+HashSize]
		offset += HashSize
	}
	data_ = data[offset:]
	return
}
