package wmpt

import (
	"encoding/binary"
	"errors"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/util/storage"
	"github.com/fxamacker/cbor/v2"
)

type Node interface {
	Key() []byte
	Hash() []byte
	CalcHash() []byte
	Copy() Node
	Weight() uint64
	Serialize() ([]byte, error)
	ToCollect() bool
	Dirty() bool
}

type (
	routingNode struct {
		key       []byte
		hash      []byte
		Children  [16]Node
		weight    uint64
		dirty     bool
		toCollect bool
	}
	valueNode struct {
		key    []byte
		hash   []byte
		value  []byte
		weight uint64
		dirty  bool
	}
	nilNode  struct{}
	hashNode struct {
		hash   []byte
		weight uint64
	}
)

var (
	emptyNode      = &nilNode{}
	emptyNodeBytes []byte
)

func init() {
	pNode := PersistNodeBase{
		NilNode: &PersistNilNode{},
	}

	emptyNodeBytes, _ = cbor.Marshal(&pNode)
}

const (
	hashWithWeightLength = 40
)

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
	if !r.dirty {
		return r.hash
	}
	m := binary.BigEndian.AppendUint64(nil, r.weight)
	for _, child := range r.Children {
		if child == nil {
			child = emptyNode
		}
		m = append(m, child.CalcHash()...)
	}
	m = append(m, r.key...)
	h := encryption.RawHash(m)
	r.hash = h
	return h
}

func (r *routingNode) Weight() uint64 {
	return r.weight
}

func (r *routingNode) ToCollect() bool {
	return r.toCollect
}

func (r *routingNode) Dirty() bool {
	return r.dirty
}

func (r *routingNode) Save(batcher storage.Batcher) error {
	if !r.dirty {
		return nil
	}
	data, err := r.Serialize()
	if err != nil {
		return err
	}
	return batcher.Put(r.hash, data)
}

func (r *routingNode) Serialize() ([]byte, error) {
	persistBranchNode := PersistNodeBranch{}
	persistBranchNode.Key = r.key
	persistBranchNode.Weight = r.weight
	r.toCollect = false
	if r.dirty {
		r.CalcHash()
		persistBranchNode.Children = make([][]byte, 16)
		for i, child := range r.Children {
			if child != nil {
				childBytes := child.Hash()
				childBytes = binary.BigEndian.AppendUint64(childBytes, child.Weight())
				persistBranchNode.Children[i] = childBytes
			}
		}
	}
	persistBranchNode.Hash = r.hash

	pNode := PersistNodeBase{
		Branch: &persistBranchNode,
	}

	return cbor.Marshal(&pNode)
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
	if !v.dirty {
		return v.hash
	}
	m := binary.BigEndian.AppendUint64(nil, v.weight)
	m = append(m, v.value...)
	h := encryption.RawHash(m)
	v.hash = h
	v.dirty = false
	return h
}

func (v *valueNode) Weight() uint64 {
	return v.weight
}

func (v *valueNode) ToCollect() bool {
	return true
}

func (v *valueNode) Dirty() bool {
	return v.dirty
}

func (v *valueNode) Save(batcher storage.Batcher) error {
	if !v.dirty {
		return nil
	}
	data, err := v.Serialize()
	if err != nil {
		return err
	}
	return batcher.Put(v.hash, data)
}

func (v *valueNode) Serialize() ([]byte, error) {
	if v.dirty {
		v.CalcHash()
	}
	pNode := PersistNodeBase{
		Value: &PersistNodeValue{
			Value:  v.value,
			Key:    v.key,
			Weight: v.weight,
			Hash:   v.hash,
		},
	}

	return cbor.Marshal(&pNode)
}

func (n *nilNode) Key() []byte {
	return []byte("")
}

func (n *nilNode) Hash() []byte {
	return emptyState
}

func (n *nilNode) Copy() Node {
	return n
}

func (n *nilNode) CalcHash() []byte {
	return emptyState
}

func (n *nilNode) Weight() uint64 {
	return 0
}

func (n *nilNode) ToCollect() bool {
	return true
}

func (n *nilNode) Dirty() bool {
	return false
}

func (n *nilNode) Serialize() ([]byte, error) {
	return emptyNodeBytes, nil
}

func NewHashNode(hash []byte, weight uint64) Node {
	return &hashNode{hash: hash, weight: weight}
}

func (h *hashNode) Key() []byte {
	return []byte("")
}

func (h *hashNode) Hash() []byte {
	return h.hash
}

func (h *hashNode) Copy() Node {
	return &hashNode{hash: h.hash, weight: h.weight}
}

func (h *hashNode) CalcHash() []byte {
	return h.hash
}

func (h *hashNode) Weight() uint64 {
	return h.weight
}

func (h *hashNode) ToCollect() bool {
	return true
}

func (h *hashNode) Dirty() bool {
	return false
}

func (h *hashNode) Serialize() ([]byte, error) {
	pNode := PersistNodeBase{
		HashNode: &PersistHashNode{
			Hash:   h.hash,
			Weight: h.weight,
		},
	}

	return cbor.Marshal(&pNode)
}

func DeserializeNode(data []byte) (Node, error) {
	pNode := PersistNodeBase{}
	err := cbor.Unmarshal(data, &pNode)
	if err != nil {
		return nil, err
	}
	if pNode.Branch != nil {
		branchNode := routingNode{}
		branchNode.key = pNode.Branch.Key
		branchNode.weight = pNode.Branch.Weight
		branchNode.hash = pNode.Branch.Hash
		for i, child := range pNode.Branch.Children {
			if len(child) == hashWithWeightLength {
				childHash := child[:32]
				childWeight := binary.BigEndian.Uint64(child[32:])
				branchNode.Children[i] = &hashNode{hash: childHash, weight: childWeight}
			}
		}
		return &branchNode, nil
	}
	if pNode.Value != nil {
		valueNode := valueNode{}
		valueNode.key = pNode.Value.Key
		valueNode.value = pNode.Value.Value
		valueNode.weight = pNode.Value.Weight
		valueNode.hash = pNode.Value.Hash
		return &valueNode, nil
	}
	if pNode.NilNode != nil {
		return &nilNode{}, nil
	}
	if pNode.HashNode != nil {
		hashNode := hashNode{}
		hashNode.hash = pNode.HashNode.Hash
		hashNode.weight = pNode.HashNode.Weight
		return &hashNode, nil
	}
	return nil, errors.New("invalid node")
}
