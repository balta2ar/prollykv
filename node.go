package main

import "bytes"

const BoundaryThreshold = 7
const HashSize = 32

type Node struct {
	level int
	key   []byte
	value []byte
	hash  []byte
}

func NewNode(key []byte, value []byte) *Node {
	this := &Node{
		level: 0,
		key:   key,
		value: value,
		hash:  rehash(key, value),
	}
	return this
}

func (this *Node) isBoundary() bool { return this.hash[0] < BoundaryThreshold }
func (this *Node) isAnchor() bool   { return this.key == nil }
func (this *Node) equal(that *Node) bool {
	return this.level == that.level &&
		bytes.Equal(this.key, that.key) &&
		bytes.Equal(this.hash, that.hash)
}

func (this *Node) parse(entry_key []byte, entry_value []byte) *Node {
	this.level = int(entry_key[0])
	this.key = entry_key[1:]
	this.hash = entry_value[:HashSize]
	this.value = entry_value[HashSize:]
	return this
}
