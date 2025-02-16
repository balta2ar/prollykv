package main

type Tree struct {
	kv KV
	// cursor
	// encoder
}

func NewTree(kv KV) *Tree {
	return &Tree{
		kv: kv,
	}
}

func (this *Tree) GetNode(level int8, key []byte) (*Node, error) {
	entry_key := append([]byte{byte(level)}, key...)
	value, err := this.kv.Get(entry_key)
	if err != nil {
		return nil, err
	}
	var node Node
	return node.parse(key, value), nil
}

func (this *Tree) SetNode(node *Node) error {
	entry_key := append([]byte{byte(node.level)}, node.key...)
	entry_value := append(node.hash, node.value...)
	return this.kv.Set(entry_key, entry_value)
}

func (this *Tree) Get(key []byte) ([]byte, error) {
	panic("not implemented")
	// return this.kv.Get(key)
}

type Iter func(cb func(key []byte, value []byte))

func (this *Tree) Build(reader Iter) {

}
