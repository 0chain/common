package trie

import (
	"strconv"
	"strings"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

//go:generate msgp -v -io=false -tests=false -unexported=true
type Node interface {
	Key() []byte
	Hash() []byte
	CalcHash() []byte
	Copy() Node
	Weight() uint64
	MarshalMsg(b []byte) ([]byte, error)
	UnmarshalMsg(b []byte) ([]byte, error)
	Msgsize() (s int)
}

type routingNode struct {
	key      []byte   `msgpack:"k"`
	hash     []byte   `msgpack:"h"`
	Children [16]Node `msgpack:"c"`
	weight   uint64   `msgpack:"w"`
}

func (r *routingNode) Key() []byte {
	return r.key
}

func (r *routingNode) Hash() []byte {
	return r.hash
}

func (r *routingNode) Copy() Node {
	keyCopy := make([]byte, len(r.key))
	copy(keyCopy, r.key)
	cpy := &routingNode{hash: r.hash, weight: r.weight, key: keyCopy}

	for i, n := range r.Children {
		if n != nil {
			cpy.Children[i] = n.Copy()
		}
	}

	return cpy
}

func (r *routingNode) CalcHash() []byte {
	m := []byte(strconv.FormatUint(r.weight, 10))
	for _, child := range r.Children {
		if child == nil {
			continue
		}
		m = append(m, child.Hash()...)
	}
	h := encryption.RawHash(m)
	r.hash = h
	return h
}

func (r *routingNode) Weight() uint64 {
	return r.weight
}

type valueNode struct {
	key    []byte `msgpack:"k"`
	hash   []byte `msgpack:"h"`
	value  []byte `msgpack:"v"`
	weight uint64 `msgpack:"w"`
}

func (v *valueNode) Key() []byte {
	return v.key
}

func (v *valueNode) Hash() []byte {
	return v.hash
}

func (v *valueNode) Copy() Node {
	keyCopy := make([]byte, len(v.key))
	copy(keyCopy, v.key)

	valueCopy := make([]byte, len(v.value))
	copy(valueCopy, v.value)

	return &valueNode{key: keyCopy, value: valueCopy, weight: v.weight, hash: v.hash}
}

func (v *valueNode) CalcHash() []byte {
	m := []byte(strconv.FormatUint(v.weight, 10))
	m = append(m, v.value...)
	h := encryption.RawHash(m)
	//fmt.Printf("hashing %v, hash %v", string(m), hex.EncodeToString(h))
	v.hash = h
	return h
}

func (v *valueNode) Weight() uint64 {
	return v.weight
}

type nilNode struct {
}

func (n *nilNode) Key() []byte {
	return []byte("")
}

func (n *nilNode) Hash() []byte {
	return nil
}

func (n *nilNode) Copy() Node {
	return n
}

func (n *nilNode) CalcHash() []byte {
	return nil
}

func (n *nilNode) Weight() uint64 {
	return 0
}

func findIndex(letter byte) int {
	for i, v := range indices {
		if v == strings.ToLower(string(letter)) {
			return i
		}
	}

	logging.Logger.Error("no index found for", zap.String("", string([]byte{letter})))
	return -1
}
