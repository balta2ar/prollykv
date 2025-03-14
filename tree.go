package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Message struct {
	timestamp string
	data      string
}

func NewMessage(timestamp string, data string) *Message {
	return &Message{timestamp: timestamp, data: data}
}

type Tree struct {
	// kv KV
	// cursor
	// encoder
	levels []*Level
}

func (t *Tree) Height() int { return len(t.levels) }
func (t *Tree) Root() *Node { return t.levels[len(t.levels)-1].tail }

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
	// Run: dot -Kneato -Tpng -o tree.png tree.dot
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "digraph G {")
	fmt.Fprintln(f, "  layout=neato;")
	fmt.Fprintln(f, "  node [shape=box, fontname=\"Arial\"];")
	fmt.Fprintln(f, "  edge [fontsize=10, fontname=\"Arial\"];")

	// Helper function to generate unique node IDs
	nodeID := func(n *Node) string {
		return fmt.Sprintf("node_%p", n)
	}

	// Map to store all nodes
	allNodes := make(map[*Node]bool)

	// First pass: collect all nodes and calculate positions
	// Calculate node positions - right to left, bottom to top
	const (
		nodeWidth  = 2.0
		nodeHeight = 1.5
		xSpacing   = 2.0
		ySpacing   = 2.0
	)

	// Build position map
	nodePositions := make(map[*Node]struct{ x, y float64 })

	for levelIdx, level := range t.levels {
		y := float64(levelIdx) * ySpacing

		// Collect nodes at this level
		var nodesAtLevel []*Node
		for n := level.tail; n != nil; n = n.left {
			nodesAtLevel = append(nodesAtLevel, n)
			allNodes[n] = true
		}

		// Position nodes from right to left
		for i, node := range nodesAtLevel {
			x := float64(len(nodesAtLevel)-1-i) * xSpacing

			// Align nodes vertically with their up/down connections
			if node.down != nil && nodePositions[node.down].x != 0 {
				// Position above its down node
				x = nodePositions[node.down].x
			}

			nodePositions[node] = struct{ x, y float64 }{x, y}
		}
	}

	// Second pass: ensure vertical alignment of nodes with up/down connections
	// Apply fixes for vertical alignment
	for node := range allNodes {
		if node.up != nil {
			upPos := nodePositions[node.up]
			nodePos := nodePositions[node]
			// Align the x positions
			if upPos.x != nodePos.x {
				nodePositions[node.up] = struct{ x, y float64 }{nodePos.x, upPos.y}
			}
		}
	}

	// Third pass: output nodes with calculated positions
	for node := range allNodes {
		pos := nodePositions[node]

		// Determine node type
		nodeType := "Regular"
		if node.isTail {
			nodeType = "Tail"
		} else if node.IsBoundary() {
			nodeType = "Boundary"
		}

		// Truncate merkleHash for cleaner display
		shortHash := node.merkleHash
		if len(shortHash) > 4 {
			shortHash = shortHash[:4]
		}

		// Create node label with all relevant information
		label := fmt.Sprintf("ts: %s\\nhash: %s\\ntype: %s\\nlevel: %d",
			node.timestamp, shortHash, nodeType, node.level)

		// Style nodes based on type
		fillcolor := "white"
		if node.isTail {
			fillcolor = "lightblue"
		} else if node.IsBoundary() {
			fillcolor = "lightgreen"
		}

		// Output node with absolute position
		fmt.Fprintf(f, "  %s [label=\"%s\", style=\"filled\", fillcolor=\"%s\", pos=\"%f,%f!\"];\n",
			nodeID(node), label, fillcolor, pos.x, pos.y)
	}

	// Output edges
	// Down/up edges (vertical connections)
	for node := range allNodes {
		if node.down != nil {
			fmt.Fprintf(f, "  %s -> %s [color=\"purple\", label=\"down\", weight=10];\n",
				nodeID(node), nodeID(node.down))
		}
		if node.up != nil {
			fmt.Fprintf(f, "  %s -> %s [color=\"red\", label=\"up\", weight=10];\n",
				nodeID(node), nodeID(node.up))
		}
	}

	// Left/right edges (horizontal connections)
	for node := range allNodes {
		if node.left != nil {
			fmt.Fprintf(f, "  %s -> %s [color=\"blue\", label=\"left\", weight=1];\n",
				nodeID(node), nodeID(node.left))
		}
		if node.right != nil {
			fmt.Fprintf(f, "  %s -> %s [color=\"green\", label=\"right\", weight=1];\n",
				nodeID(node), nodeID(node.right))
		}
	}

	// Add invisible subgraphs for each level to help with visualization
	for i, level := range t.levels {
		fmt.Fprintf(f, "  subgraph cluster_level_%d {\n", i)
		fmt.Fprintf(f, "    label=\"Level %d\";\n", level.level)
		fmt.Fprintln(f, "    style=invis;")

		// Add all nodes in this level to the subgraph
		for n := level.tail; n != nil; n = n.left {
			fmt.Fprintf(f, "    %s;\n", nodeID(n))
		}

		fmt.Fprintln(f, "  }")
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
		s := fmt.Sprintf("%s", t.timestamp)
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
	nodes = append(nodes, NewNode(TailKey(), "tail", isTail))
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
	linked := LinkNodes(eligible)
	for _, n := range linked {
		n.FillMerkleHash()
	}
	next := NewLevel(prev.level + 1)
	next.tail = linked[len(linked)-1]
	next.size = len(linked)
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

// const HashSize = 32
const HashSize = 64

type Node struct {
	level      int8
	timestamp  string // key
	data       string // value
	up         *Node
	down       *Node
	left       *Node
	right      *Node
	merkleHash string // rolling merkle hash
	boundary   *bool
	isTail     bool
}

func (n *Node) Iter() Iter { return &NodeIter{P: n} }

// types of Nodes
// boundary / promoted -- leades to node promotion, nodeHash <= BoundaryThreshold
//   contains rolling merkleHash of the group of non-boundary nodes
// tail / anchor / fake -- Node(non-data) inserted at each level.
//   always a boundary node by default. Convas put tail nodes on the left side of the tree.
//   in this design, it's on the right side.
//

func TailKey() string           { return "<TAIL>" }
func IsTailKey(key string) bool { return key == TailKey() }

func NewNode(timestamp string, data string, isTail bool) *Node {
	payload := timestamp + data
	hash := Rehash(payload)
	node := &Node{
		timestamp:  timestamp,
		data:       data,
		isTail:     isTail,
		merkleHash: hash,
		boundary:   nil,
	}
	return node
}

func (n *Node) Key() string {
	return StrEncodeKey(n.level, fmt.Sprintf("%s", n.timestamp))
}

func (n *Node) Value() string {
	return StrEncodeValue(n.merkleHash, n.data)
}

func (n *Node) KeyWithKids() string {
	// hash
	return StrEncodeKeyWithKids(n.merkleHash)
}

func (n *Node) ValueWithKids() string {
	kids := []string{}
	n.Kids(func(n *Node) {
		kids = append(kids, n.merkleHash)
	})
	return StrEncodeValueWithKids(n.level, kids, n.timestamp, n.data)
}

func (n *Node) String() string {
	return fmt.Sprintf("Node(timestamp=%q, level=%d)", n.timestamp, n.level)
}

// -1 when left is less, 0 when equal, 1 when right is less
func (n *Node) CompareKey(o *Node) int {
	switch fmt.Sprintf("%t_%t", n.isTail, o.isTail) {
	case "true_true":
		return 0
	case "true_false":
		return 1
	case "false_true":
		return -1
	case "false_false":
		if n.timestamp < o.timestamp {
			return -1
		}
		if n.timestamp > o.timestamp {
			return 1
		}
		return 0
	default:
		panic("unreachable")
	}
}

func (n *Node) Descend(targetLevel int8) *Node {
	var p = n
	for ; p.level > targetLevel; p = p.down {
	}
	return p
}

func (n *Node) Bottom() *Node {
	for ; n.down != nil; n = n.down {
	}
	return n
}

func (n *Node) Kids(cb func(*Node)) {
	if n.down == nil {
		return
	}
	n.down.UntilBoundary(cb)
}

func (n *Node) ListKids() (out []string) {
	n.Kids(func(p *Node) {
		out = append(out, p.merkleHash)
	})
	return out
}

func (n *Node) UntilBoundary(cb func(*Node)) {
	cb(n)
	for p := n.left; p != nil; p = p.left {
		if p.IsBoundary() {
			break
		}
		cb(p)
	}
}

func (n *Node) IsBoundary() bool {
	if n.boundary != nil {
		return *n.boundary
	}
	boundary := n.isTail || IsBoundaryHash(n.merkleHash)
	n.boundary = &boundary
	return boundary
}

func (n *Node) CreateHigherLevel() *Node {
	node := NewNode(n.timestamp, "", n.isTail)
	node.level = n.level + 1
	node.down = n
	n.up = node
	node.merkleHash = "" // to be filled later by FillMerkleHash
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
	must(len(digit) == 1, "hash must be a single digit")
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

func must(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// ----------------------------------------------------------------

// func GetNonBoundaryNodes(nodes []*Node) []*Node {
// 	var out []*Node
// 	for _, n := range nodes {
// 		fmt.Println("n.down ", n.down)
// 		if n.down == nil {
// 			continue
// 		}
// 		n.down.UntilBoundary(func(p *Node) {
// 			out = append(out, p)
// 		})
// 	}
// 	slices.Reverse(out)
// 	return out
// }

// func GetNonBoundaryNodesForLevel0(nodes []*Node) []*Node {
// 	var out []*Node
// 	for _, n := range nodes {
// 		n.Bottom().UntilBoundary(func(p *Node) {
// 			out = append(out, p)
// 		})
// 	}
// 	slices.Reverse(nodes)
// 	return out
// }

type Delta struct {
	key    string // timestamp
	typ    string // "add", "remove", "update"
	source string
	target string
}

type DeltaTrio struct {
	Add    []Delta
	Remove []Delta
	Update []Delta
}

type Iter interface {
	Current() *Node
	Left() *Node
}

type NodeIter struct {
	P *Node
}

var _ Iter = &NodeIter{}

func (n *NodeIter) Current() *Node { return n.P }
func (n *NodeIter) Left() *Node {
	if n.P != nil {
		n.P = n.P.left
	}
	return n.P
}

type Boundary struct {
	Iter

	p    *Node
	done bool
}

var _ Iter = &Boundary{}

func (b *Boundary) Current() *Node {
	if b.done {
		return nil
	}
	if b.p == nil {
		return b.Iter.Current()
	}
	return b.p
}
func (b *Boundary) Left() *Node {
	if b.done {
		return nil
	}
	b.p = b.Iter.Left()
	if b.p == nil || b.p.IsBoundary() {
		b.done = true
		return nil
	}
	return b.p
}

type Chain struct {
	nodes []Iter
	p     *Node
}

var _ Iter = &Chain{}

type NilIter struct{}

var _ Iter = &NilIter{}

func (n *NilIter) Current() *Node { return nil }
func (n *NilIter) Left() *Node    { return nil }

func NewChain(nodes ...Iter) Iter {
	// must(len(nodes) > 0, "at least one node is required")
	if len(nodes) == 0 {
		return &NilIter{}
	}
	c := &Chain{nodes: nodes}
	c.p = c.nodes[0].Current()
	return c
}

func (c *Chain) Current() *Node { return c.p }
func (c *Chain) Left() *Node {
	if c.p != nil {
		c.p = c.nodes[0].Left()
	}
	if c.p == nil {
		if len(c.nodes) > 1 {
			c.nodes = c.nodes[1:]
			c.p = c.nodes[0].Current()
		}
	}
	return c.p
}

func Diff(source, target *Tree) (out DeltaTrio) {
	minHeight := min(source.Root().level, target.Root().level)
	s, t := source.Root().Descend(minHeight), target.Root().Descend(minHeight)
	must(s.level == t.level, "levels must match")

	var add, update []Delta
	emitUpdate := func(p1, p2 *Node) {
		if update != nil {
			fmt.Printf("= %v => %v\n", p1.timestamp, p2.timestamp)
			update = append(update, Delta{key: p2.timestamp, typ: "update", source: p1.data, target: p2.data})
		}
	}
	emitAdd := func(p2 *Node) {
		if add != nil {
			fmt.Printf("+ p2 %v\n", p2.merkleHash[:4])
			add = append(add, Delta{key: p2.timestamp, typ: "add", source: "", target: p2.data})
		}
	}
	emitAddAll := func(p2 Iter) {
		for p := p2.Current(); p != nil; p = p2.Left() {
			emitAdd(p)
		}
	}

	var diffAtLevel func(nodes1, nodes2 Iter, level int8)
	diffAtLevel = func(nodes1, nodes2 Iter, level int8) {
		if level < 0 {
			return
		}

		moreNodes1 := []Iter{}
		moreNodes2 := []Iter{}

		for l, r := nodes1.Current(), nodes2.Current(); l != nil && r != nil; {
			fmt.Printf("L%d l=%v r=%v key %q %q\n", level, l.merkleHash[:4], r.merkleHash[:4], l.timestamp, r.timestamp)
			switch l.CompareKey(r) {
			case -1: // l < r
				// the r subtree is missing, push it down or add if we're on level0
				if r.level == 0 {
					emitAdd(r)
				} else {
					moreNodes2 = append(moreNodes2, &Boundary{Iter: r.down.Iter()})
				}
				r = nodes2.Left()
			case 0: // l == r
				if l.merkleHash != r.merkleHash {
					if r.level == 0 { // no point in going down on level0
						emitUpdate(l, r)
					} else {
						// inspect the subtree, but with a limit -- up to the next boundary
						moreNodes1 = append(moreNodes1, &Boundary{Iter: l.down.Iter()})
						moreNodes2 = append(moreNodes2, &Boundary{Iter: r.down.Iter()})
					}
				}
				l = nodes1.Left()
				r = nodes2.Left()
			case 1: // l > r
				// the whole l subtree is missing -- we skip it because the reverse run will catch it
				if l.level > 0 {
					moreNodes1 = append(moreNodes1, &Boundary{Iter: l.down.Iter()})
				}
				l = nodes1.Left()
			}
		}

		fmt.Printf("nodes1 %v\n", nodes1.Current())
		fmt.Printf("nodes2 %v\n", nodes2.Current())

		for l := nodes1.Current(); l != nil; l = nodes1.Left() {
			if l.level > 0 {
				moreNodes1 = append(moreNodes1, &Boundary{Iter: l.down.Iter()})
			}
		}

		// if anything remains in the right, it should be added or pushed down.
		for r := nodes2.Current(); r != nil; r = nodes2.Left() {
			if r.level == 0 {
				emitAdd(r)
			} else {
				moreNodes2 = append(moreNodes2, &Boundary{Iter: r.down.Iter()})
			}
		}

		must(nodes1.Current() == nil, "nodes1 must be exhausted")
		must(nodes2.Current() == nil, "nodes2 must be exhausted")

		if len(moreNodes1) == 0 && len(moreNodes2) == 0 { // no more nodes worth inspecting
			return
		} else if len(moreNodes1) == 0 { // left is empty, add everything from the right
			nodes2 = NewChain(moreNodes2...)
			for r := nodes2.Current(); r != nil; r = nodes2.Left() {
				emitAddAll(&Boundary{Iter: r.Bottom().Iter()})
			}
			return
		}

		nodes1 = NewChain(moreNodes1...)
		nodes2 = NewChain(moreNodes2...)
		diffAtLevel(nodes1, nodes2, level-1)
	}
	add = []Delta{}
	update = []Delta{}
	diffAtLevel(s.Iter(), t.Iter(), s.level)
	out.Add, out.Update = add, update

	add = []Delta{}
	update = nil
	diffAtLevel(t.Iter(), s.Iter(), s.level)
	out.Remove = add
	return out
}

func (t *Tree) SerializeLevel0(onto KV) error {
	level := t.levels[0]
	for n := level.tail; n != nil; n = n.left {
		key := n.Key()
		value := n.Value()
		err := onto.Set([]byte(key), []byte(value))
		if err != nil {
			return err
		}
	}
	return onto.Set([]byte("root"), []byte(t.Root().Key()))
}

func DeserializeLevel0(kv KV) (*Tree, error) {
	cur := kv.Cursor()
	start := StrEncodeKey(0, "")
	cur.Goto([]byte(start))
	level0 := []*Message{}
	for ; strings.HasPrefix(string(cur.Key()), start); cur.Next() {
		encodedKey := cur.Key()
		encodedValue := cur.Value()
		_, key := StrDecodeKey(string(encodedKey))
		_, value := StrDecodeValue(string(encodedValue))
		// intKey := MustAtoi(key)
		if IsTailKey(key) {
			continue
		}
		m := &Message{timestamp: key, data: value}
		level0 = append(level0, m)
	}
	return NewTree(level0), nil
}

func MustAtoi(s string) int {
	v, err := strconv.Atoi(s)
	mustNil(err)
	return v
}

type CountingKV struct {
	KV
	stats map[string]int
}

func NewCountingKV(kv KV) *CountingKV { return &CountingKV{KV: kv, stats: map[string]int{}} }
func (kv *CountingKV) Set(key []byte, value []byte) error {
	kv.stats["set"]++
	return kv.KV.Set(key, value)
}
func (kv *CountingKV) Get(key []byte) ([]byte, bool, error) {
	kv.stats["get"]++
	return kv.KV.Get(key)
}
func (kv *CountingKV) String() string { return fmt.Sprintf("CountingKV{stats=%v}", kv.stats) }

func (t *Tree) SerializeWithKids(gen int, onto KV) error {
	for _, level := range t.levels {
		for n := level.tail; n != nil; n = n.left {
			key := StrEncodeKeyWithKids(n.merkleHash)
			value := StrEncodeValueWithKids(n.level, n.ListKids(), n.timestamp, n.data)
			err := onto.Set([]byte(key), []byte(value))
			if err != nil {
				return err
			}
		}
	}
	rootKeyName := fmt.Sprintf("root:%d", gen)
	return onto.Set([]byte(rootKeyName), []byte(t.Root().merkleHash))
}

func DeserializeWithKids(gen int, kv KV) (*Tree, error) {
	rootKeyName := fmt.Sprintf("root:%d", gen)
	kvKey, found, err := kv.Get([]byte(rootKeyName))
	mustNil(err)
	mustTrue(found, "root key not found: %q", rootKeyName)
	hashes := []string{string(kvKey)}
	nextHashes := []string{}
	messages := []*Message{}
	for len(hashes) > 0 {
		for _, key := range hashes {
			value, found, err := kv.Get([]byte(key))
			mustNil(err)
			mustTrue(found, "key not found: %q", key)
			kidLevel, kids, kidKey, data := StrDecodeValueWithKids(string(value))
			if kidLevel == 0 {
				if !IsTailKey(kidKey) {
					messages = append(messages, &Message{timestamp: kidKey, data: data})
				}
			} else {
				nextHashes = append(nextHashes, kids...)
			}
		}
		hashes, nextHashes = nextHashes, []string{}
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].timestamp < messages[j].timestamp
	})
	return NewTree(messages), nil
}

func (t *Tree) SerializeJSON(gen int, w1 io.Writer) error {
	w := &strings.Builder{}
	fmt.Fprintf(w, "{\"gen\":%d,\"root\":\"%s\",\"nodes\":[", gen, t.Root().merkleHash)

	first := true
	for _, level := range t.levels {
		for n := level.tail; n != nil; n = n.left {
			if !first {
				fmt.Fprint(w, ",")
			}
			first = false

			kids := n.ListKids()
			kidsJSON := "["
			for i, kid := range kids {
				if i > 0 {
					kidsJSON += ","
				}
				kidsJSON += "\"" + kid + "\""
			}
			kidsJSON += "]"

			fmt.Fprintf(w, "{\"hash\":\"%s\",\"level\":%d,\"timestamp\":%s,\"data\":\"%s\",\"kids\":%s}",
				n.merkleHash, n.level, n.timestamp, n.data, kidsJSON)
		}
	}

	fmt.Fprint(w, "]}")
	_, err := w1.Write([]byte(w.String()))
	return err
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
