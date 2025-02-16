package main

import "bytes"

const BoundaryThreshold = 7

type Node struct {
	level int
	key   []byte
	hash  []byte
	value []byte
}

func (this *Node) isBoundary() bool { return this.hash[0] < BoundaryThreshold }
func (this *Node) isAnchor() bool   { return this.key == nil }
func (this *Node) equal(that *Node) bool {
	return this.level == that.level &&
		bytes.Equal(this.key, that.key) &&
		bytes.Equal(this.hash, that.hash)
}

func (this *Node) parse(key []byte, value []byte) *Node {
	panic("not implemented")
	return this
}
