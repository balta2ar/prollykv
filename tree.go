package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
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
	// kv KV
	// cursor
	// encoder
	levels []*Level
}

func (t *Tree) Root() *Node {
	return t.levels[len(t.levels)-1].tail.down.left
}

func NewTree(messages []*Message) *Tree {
	tree := &Tree{}
	base := BaseLevel(messages)
	tree.levels = append(tree.levels, base)
	for !base.OnlyTail() {
		base = NextLevel(base)
		tree.levels = append(tree.levels, base)
	}
	return tree
}

func (t *Tree) Dot(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "digraph {")
	fmt.Fprintln(f, "  node [shape=box];")
	fmt.Fprintln(f, "  rankdir=LR;")

	// Map to store node IDs to avoid duplicates
	visited := make(map[*Node]bool)

	// Generate unique node IDs
	nodeID := func(n *Node) string {
		if n == nil {
			return "nil"
		}
		return fmt.Sprintf("n%p", n)
	}

	// Write a node to the DOT file
	writeNode := func(n *Node) {
		if n == nil || visited[n] {
			return
		}
		visited[n] = true

		// Node definition
		label := fmt.Sprintf("ts:%d\\nlvl:%d", n.timestamp, n.level)
		if n.isTail {
			label += "\\nTAIL"
		}
		if n.IsBoundary() {
			label += "\\nBOUNDARY"
		}
		fmt.Fprintf(f, "  %s [label=\"%s\"];\n", nodeID(n), label)

		// Edge definitions
		if n.left != nil {
			fmt.Fprintf(f, "  %s -> %s [color=blue, label=\"left\"];\n", nodeID(n), nodeID(n.left))
		}
		if n.right != nil {
			fmt.Fprintf(f, "  %s -> %s [color=green, label=\"right\"];\n", nodeID(n), nodeID(n.right))
		}
		if n.up != nil {
			fmt.Fprintf(f, "  %s -> %s [color=red, label=\"up\"];\n", nodeID(n), nodeID(n.up))
		}
		if n.down != nil {
			fmt.Fprintf(f, "  %s -> %s [color=purple, label=\"down\"];\n", nodeID(n), nodeID(n.down))
		}
	}

	// Traverse the tree and write all nodes
	for _, level := range t.levels {
		for node := level.tail; node != nil; node = node.left {
			writeNode(node)
		}
	}

	// Create a subgraph for each level to ensure proper visual hierarchy
	for i, level := range t.levels {
		fmt.Fprintf(f, "  subgraph cluster_level_%d {\n", i)
		fmt.Fprintf(f, "    label=\"Level %d\";\n", level.level)
		fmt.Fprintf(f, "    rank=same;\n")

		for node := level.tail; node != nil; node = node.left {
			fmt.Fprintf(f, "    %s;\n", nodeID(node))
		}
		fmt.Fprintln(f, "  }")
	}

	// Add nil node if needed (for edges pointing to nil)
	if len(visited) > 0 {
		fmt.Fprintln(f, "  nil [shape=point];")
	}

	fmt.Fprintln(f, "}")
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
	xs := []string{"tail"}
	for t := l.tail; t != nil; t = t.left {
		s := fmt.Sprintf("%d", t.timestamp)
		if t.IsBoundary() {
			s += "*"
		}
		xs = append(xs, s)
	}
	xs = append(xs, "nil")
	slices.Reverse(xs)
	ts := strings.Join(xs, " < ")
	return fmt.Sprintf("Level(level=%d, size=%d, %s)", l.level, l.size, ts)
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

func (n *Node) String() string {
	return fmt.Sprintf("Node(timestamp=%d, level=%d)", n.timestamp, n.level)
}

func (n *Node) Descend(o *Node) *Node {
	var p = n
	for ; p.level > o.level; p = p.down {
	}
	return p
}

func (n *Node) Bottom() *Node {
	for ; n.down != nil; n = n.down {
	}
	return n
}

func (n *Node) UntilBoundary(cb func(*Node)) {
	for p := n; p != nil; p = p.left {
		cb(p)
		if p.IsBoundary() {
			break
		}
	}
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

const AverageBucketSize = 10
const BoundaryThreshold = uint32((1 << 32) / AverageBucketSize)

func IsBoundaryHash(hash string) bool {
	hashBytes, _ := hex.DecodeString(hash)
	value := binary.BigEndian.Uint32(hashBytes[:4])
	return value < BoundaryThreshold
}

const BoundaryThresholdBits = 5

func IsBoundaryHash2(hash string) bool {
	digit := hash[:1]
	assert(len(digit) == 1, "hash must be a single digit")
	hashInt, _ := strconv.ParseInt(digit, 16, 64)
	return hashInt < BoundaryThresholdBits
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

func GetNonBoundaryNodes(nodes []*Node) []*Node {
	var out []*Node
	for _, n := range nodes {
		fmt.Println("n.down ", n.down)
		if n.down == nil {
			continue
		}
		n.down.UntilBoundary(func(p *Node) {
			out = append(out, p)
		})
	}
	slices.Reverse(out)
	return out
}

func GetNonBoundaryNodesForLevel0(nodes []*Node) []*Node {
	var out []*Node
	for _, n := range nodes {
		n.Bottom().UntilBoundary(func(p *Node) {
			out = append(out, p)
		})
	}
	slices.Reverse(nodes)
	return out
}

type Delta struct {
	key    int    // timestamp
	typ    string // "add", "remove", "update"
	source string
	target string
}

type Chain struct {
	nodes []*Node
	p     *Node
}

func NewChain(nodes ...*Node) *Chain {
	c := &Chain{nodes: nodes}
	c.Left()
	return c
}

func (c *Chain) Append(n *Node)  { c.nodes = append(c.nodes, n) }
func (c *Chain) Reverse() *Chain { slices.Reverse(c.nodes); return c }
func (c *Chain) More() bool      { return c.Current() != nil }
func (c *Chain) Current() *Node  { return c.p }
func (c *Chain) Left() *Node {
	if c.p != nil {
		c.p = c.p.left
	}
	if c.p == nil {
		k := len(c.nodes)
		if k > 0 {
			c.p = c.nodes[k-1]
			c.nodes = c.nodes[:k-1]
		}
	}
	return c.p
}
func (c *Chain) Nodes() []*Node {
	var out []*Node
	for n := c.Current(); n != nil; n = c.Left() {
		out = append(out, n)
	}
	return out
}

func Diff(source, target *Tree) []Delta {
	var out []Delta
	s, t := source.Root().Descend(target.Root()), target.Root().Descend(source.Root())
	assert(s.level == t.level, "levels must match")
	nodes1 := NewChain(s)
	nodes2 := NewChain(t)
	fmt.Println("nodes1", nodes1)
	fmt.Println("nodes2", nodes2)

	var diffAtLevel func(nodes1, nodes2 *Chain, level int8)
	diffAtLevel = func(nodes1, nodes2 *Chain, level int8) {
		if level < 0 {
			return
		}
		fmt.Println("diffAtLevel", level)
		fmt.Println("starting with nodes1", nodes1)
		fmt.Println("starting with nodes2", nodes2)
		moreNodes1 := []*Node{}
		moreNodes2 := []*Node{}

		addP2 := func(p2 *Node) {
			if p2.level == 0 {
				fmt.Println("OUT p2", p2)
				out = append(out, Delta{key: p2.timestamp, typ: "add", source: "", target: p2.data})
			} else {
				moreNodes2 = append(moreNodes2, p2)
			}
		}

		for nodes1.More() && nodes2.More() {
			p1, p2 := nodes1.Current(), nodes2.Current()
			if p1.timestamp == p2.timestamp { // might be update
				fmt.Println("! p1 == p2 [could be update]", p1, p2)
				if p1.merkleHash != p2.merkleHash {
					fmt.Println("queing for expansion both", p1, p2)
					moreNodes1 = append(moreNodes1, p1)
					moreNodes2 = append(moreNodes2, p2)
				}
				nodes1.Left()
				nodes2.Left()
			} else if p1.timestamp < p2.timestamp { // add
				fmt.Println("! p1 < p2 (add p2)", p2)
				// source: p1=1        3
				// target:    1  p2=2  3
				addP2(p2)
				nodes2.Left()
			} else { // p1.timestamp > p2.timestamp, remove
				fmt.Println("! p1 > p2 (remove p1)", p1)
				// source:    1  p1=2  3
				// target: p2=1        3
				nodes1.Left() // keys are gone, we skip them (reverse run will catch them)
			}
		}

		// for ; p2 != nil; p2, nodes2 = left(p2, nodes2) {
		// 	fmt.Println("leftover p2", p2)
		// 	moreNodes2 = append(moreNodes2, p2)
		// }
		for _, p2 := range nodes2.Nodes() {
			addP2(p2)
		}
		nodes1 = NewChain(GetNonBoundaryNodes(moreNodes1)...)
		nodes2 = NewChain(GetNonBoundaryNodes(moreNodes2)...)
		fmt.Println("expanded nodes1", nodes1)
		fmt.Println("expanded nodes2", nodes2)
		if !nodes1.More() && !nodes2.More() {
			fmt.Println("no more nodes")
			return
		} else if !nodes1.More() { // add everything from the target
			fmt.Println("OUT everything from the target")
			for _, p2 := range GetNonBoundaryNodesForLevel0(nodes2.Nodes()) {
				out = append(out, Delta{key: p2.timestamp, typ: "add", source: "", target: p2.data})
			}
		}
		diffAtLevel(nodes1, nodes2, level-1)
	}
	diffAtLevel(nodes1, nodes2, s.level)
	return out
}

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
