package main

import (
	"slices"
	"strconv"
	"strings"
)

const HashSize = 32

type Node struct {
	level      int8
	timestamp  int
	data       string
	up         *Node
	down       *Node
	left       *Node
	right      *Node
	nodeHash   string // node hash, sha256
	merkleHash string // rolling merkle hash
	boundary   *bool
	isTail     bool
	// key       []byte
	// value     []byte
	// hash       []byte // merkle hash
}

// types of Nodes
// boundary / promoted -- leades to node promotion, nodeHash <= BoundaryThreshold
//   contains rolling merkleHash of the group of non-boundary nodes
// tail / anchor / fake -- Node(non-data) inserted at each level.
//   always a boundary node by default. Convas put tail nodes on the left side of the tree.
//   in this design, it's on the right side.
//

// func NewNode(key []byte, value []byte) *Node {
// 	this := &Node{
// 		level: 0,
// 		key:   key,
// 		value: value,
// 		hash:  rehash(key, value),
// 	}
// 	return this
// }

func NewNode(timestamp int, data string, isTail bool) *Node {
	payload := strconv.Itoa(timestamp) + data
	hash := Rehash(payload)
	this := &Node{
		timestamp:  timestamp,
		data:       data,
		isTail:     isTail,
		merkleHash: hash,
		nodeHash:   hash,
		boundary:   nil,
	}
	return this
}

func (n *Node) IsBoundary() bool {
	if n.boundary != nil {
		return *n.boundary
	}
	boundary := n.isTail || IsBoundaryHash(n.nodeHash)
	n.boundary = &boundary
	return boundary
}

func (n *Node) CreateHigherLevel() *Node {
	node := NewNode(n.timestamp, "", n.isTail)
	node.level = n.level + 1
	node.down = n
	n.up = node
	node.nodeHash = Rehash(n.nodeHash)
	return node
}

func (n *Node) FillMerkleHash() {
	var bucket []*Node
	node := n.down
	bucket = append(bucket, node)

	for node.left != nil {
		if node.left.IsBoundary() {
			break
		}
		node = node.left
		bucket = append(bucket, node)
	}

	for _, node := range bucket {
		if node.merkleHash == "" {
			node.FillMerkleHash()
		}
	}

	slices.Reverse(bucket)
	n.merkleHash = BucketHash(bucket)
}

const BoundaryThreshold = 7

func IsBoundaryHash(hash string) bool {
	digit := hash[:1]
	assert(len(digit) == 1, "hash must be a single digit")
	hashInt, _ := strconv.ParseInt(digit, 16, 64)
	return hashInt < BoundaryThreshold
}

func BucketHash(nodes []*Node) string {
	var sb strings.Builder
	for _, node := range nodes {
		sb.WriteString(node.merkleHash)
	}
	return Rehash(sb.String())
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// func (this *Node) isBoundary() bool { return this.hash[0] < BoundaryThreshold }
// func (this *Node) isAnchor() bool   { return this.key == nil }
// func (this *Node) equal(that *Node) bool {
// 	return this.level == that.level &&
// 		bytes.Equal(this.key, that.key) &&
// 		bytes.Equal(this.hash, that.hash)
// }

// func (this *Node) parse(entry_key []byte, entry_value []byte) *Node {
// 	this.level, this.key = DecodeKey(entry_key)
// 	this.hash, this.value = DecodeValue(entry_value)
// 	return this
// }
