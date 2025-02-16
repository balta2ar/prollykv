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

func (this *Tree) Get(key []byte) ([]byte, error) {
	panic("not implemented")
	// return this.kv.Get(key)
}

type Iter func(cb func(key []byte, value []byte))

func (this *Tree) Build(reader Iter) {

}
