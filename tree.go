package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Message struct {
	timestamp int
	data      string
}

func NewMessage(timestamp int, data string) *Message {
	return &Message{timestamp: timestamp, data: data}
}

type Tree struct {
	kv KV
	// cursor
	// encoder
	levels []*Level
}

func NewTree(kv KV, messages []*Message) *Tree {
	tree := &Tree{
		kv: kv,
	}

	base := BaseLevel(messages)
	tree.levels = append(tree.levels, base)
	for !base.OnlyTail() {
		base = NextLevel(base)
		tree.levels = append(tree.levels, base)
	}
	return tree
}

func (t *Tree) String() string {
	var sb strings.Builder
	for _, level := range t.levels {
		sb.WriteString(level.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

type Level struct {
	level int8
	size  int
	tail  *Node
}

func NewLevel(level int8) *Level {
	return &Level{level: level}
}

func (l *Level) OnlyTail() bool {
	return l.tail.left == nil
}

func (l *Level) AsList() []*Node {
	var nodes []*Node
	for n := l.tail; n != nil; n = n.left {
		nodes = append(nodes, n)
	}
	slices.Reverse(nodes)
	return nodes
}

func (l *Level) String() string {
	return fmt.Sprintf("Level(level=%d, size=%d)", l.level, l.size)
}

func BaseLevel(messages []*Message) *Level {
	level := NewLevel(0)
	var nodes []*Node
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].timestamp < messages[j].timestamp
	})
	const notTail = false
	for _, m := range messages {
		nodes = append(nodes, NewNode(m.timestamp, m.data, notTail))
	}
	const isTail = true
	fakeTail := NewNode(-1, "tail", isTail)
	nodes = append(nodes, fakeTail)
	level.tail = LinkNodes(nodes)[len(nodes)-1]
	level.size = len(nodes)
	return level
}

func NextLevel(prev *Level) *Level {
	nodes := prev.AsList()
	var eligible []*Node
	for _, n := range nodes {
		if n.IsBoundary() {
			eligible = append(eligible, n.CreateHigherLevel())
		}
	}
	LinkNodes(eligible)
	for _, n := range nodes {
		n.FillMerkleHash()
	}
	next := NewLevel(prev.level + 1)
	next.tail = eligible[len(eligible)-1]
	next.size = len(eligible)
	return next
}

func LinkNodes(nodes []*Node) []*Node {
	for i := range len(nodes) - 1 {
		nodes[i].right = nodes[i+1]
		nodes[i+1].left = nodes[i]
	}
	return nodes
}

// ----------------------------------------------------------------

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

func NewNode(timestamp int, data string, isTail bool) *Node {
	payload := strconv.Itoa(timestamp) + data
	hash := Rehash(payload)
	node := &Node{
		timestamp:  timestamp,
		data:       data,
		isTail:     isTail,
		merkleHash: hash,
		nodeHash:   hash,
		boundary:   nil,
	}
	return node
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
	node := n.down
	if node == nil {
		return
	}

	var bucket []*Node
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

const BoundaryThreshold = 5

func IsBoundaryHash(hash string) bool {
	digit := hash[:1]
	assert(len(digit) == 1, "hash must be a single digit")
	hashInt, _ := strconv.ParseInt(digit, 16, 64)
	return hashInt < BoundaryThreshold
}

func Rehash(xs ...string) string {
	h := sha256.New()
	for _, x := range xs {
		h.Write([]byte(x))
	}
	return hex.EncodeToString(h.Sum(nil))
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

// ----------------------------------------------------------------

// func (this *Tree) GetNode(level int8, key []byte) (*Node, error) {
// 	entry_key := EncodeKey(level, key)
// 	value, err := this.kv.Get(entry_key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var node Node
// 	return node.parse(key, value), nil
// }

// func (this *Tree) SetNode(node *Node) error {
// 	entry_key := EncodeKey(node.level, node.key)
// 	entry_value := EncodeValue(node.hash, node.value)
// 	return this.kv.Set(entry_key, entry_value)
// }

// func (this *Tree) Get(key []byte) ([]byte, error) {
// 	panic("not implemented")
// 	// return this.kv.Get(key)
// }

// type Iter func(cb func(key []byte, value []byte) error) error

// func sortedKeys(m map[string]string) []string {
// 	keys := make([]string, 0, len(m))
// 	for key := range m {
// 		keys = append(keys, key)
// 	}
// 	sort.Strings(keys)
// 	return keys
// }

// func (this *Tree) Build(m map[string]string) error {
// 	keys := sortedKeys(m)
// 	for _, key := range keys {
// 		value := m[key]
// 		node := NewNode([]byte(key), []byte(value))
// 		err := this.SetNode(node)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
