package wmpt

import (
	"encoding/binary"
	"errors"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/util/storage"
	"github.com/fxamacker/cbor/v2"
)

type Node interface {
	Hash() []byte
	CalcHash() []byte
	Copy() Node
	CopyRoot(currLevel, collapseLevel int) Node
	Weight() uint64
	Serialize() ([]byte, error)
	ToCollect() bool
	Dirty() bool
}

type (
	routingNode struct {
		hash      []byte
		Children  [16]Node
		weight    uint64
		dirty     bool
		toCollect bool
	}
	valueNode struct {
		hash   []byte
		value  []byte
		weight uint64
		dirty  bool
	}
	shortNode struct {
		key       []byte
		hash      []byte
		value     Node
		dirty     bool
		toCollect bool
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
	hashWithWeightLength     = 40
	branchNodeLength         = 16
	branchNodeHashDataLength = 16*32 + 8 //16 children + 8 bytes for weight
)

func (r *routingNode) Hash() []byte {
	return r.hash
}

func (r *routingNode) Copy() Node {
	cpy := &routingNode{hash: r.hash, weight: r.weight}

	for i, n := range r.Children {
		if n != nil {
			cpy.Children[i] = &hashNode{
				hash:   n.Hash(),
				weight: n.Weight(),
			}
		}
	}

	return cpy
}

func (r *routingNode) CopyRoot(currLevel, collapseLevel int) Node {
	if currLevel == collapseLevel {
		return &hashNode{
			hash:   r.hash,
			weight: r.weight,
		}
	}
	cpy := &routingNode{hash: r.hash, weight: r.weight}

	for i, n := range r.Children {
		if n != nil {
			cpy.Children[i] = n.CopyRoot(currLevel+1, collapseLevel)
		}
	}
	return cpy
}

func (r *routingNode) CalcHash() []byte {
	if !r.dirty {
		return r.hash
	}
	m := make([]byte, 0, branchNodeHashDataLength)
	m = binary.BigEndian.AppendUint64(m, r.weight)
	for _, child := range r.Children {
		if child == nil {
			child = emptyNode
		}
		m = append(m, child.CalcHash()...)
	}
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
	r.toCollect = false
	if r.dirty {
		r.CalcHash()
	}
	persistBranchNode.Children = make([][]byte, 16)
	for i, child := range r.Children {
		if child != nil {
			var childBytes []byte
			if c, ok := child.(*shortNode); ok {
				childBytes = make([]byte, 0, hashWithWeightLength+32+len(c.key))
				childBytes = append(childBytes, child.Hash()...)
				childBytes = binary.BigEndian.AppendUint64(childBytes, child.Weight())
				childBytes = append(childBytes, c.value.Hash()...)
				childBytes = append(childBytes, c.key...)
			} else {
				childBytes = make([]byte, 0, hashWithWeightLength)
				childBytes = append(childBytes, child.Hash()...)
				childBytes = binary.BigEndian.AppendUint64(childBytes, child.Weight())
			}
			persistBranchNode.Children[i] = childBytes
		}
	}

	persistBranchNode.Hash = r.hash

	pNode := PersistNodeBase{
		Branch: &persistBranchNode,
	}

	return cbor.Marshal(&pNode)
}

func (v *valueNode) Hash() []byte {
	return v.hash
}

func (v *valueNode) Copy() Node {
	valueCopy := make([]byte, len(v.value))
	copy(valueCopy, v.value)

	return &valueNode{value: valueCopy, weight: v.weight, hash: v.hash}
}

func (v *valueNode) CopyRoot(currLevel, collapseLevel int) Node {
	return v.Copy()
}

func (v *valueNode) CalcHash() []byte {
	if !v.dirty {
		return v.hash
	}
	m := make([]byte, 0, hashWithWeightLength)
	m = binary.BigEndian.AppendUint64(m, v.weight)
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
			Weight: v.weight,
			Hash:   v.hash,
		},
	}

	return cbor.Marshal(&pNode)
}

func (n *nilNode) Hash() []byte {
	return emptyState
}

func (n *nilNode) Copy() Node {
	return n
}

func (n *nilNode) CopyRoot(currLevel, collapseLevel int) Node {
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

func (h *hashNode) Hash() []byte {
	return h.hash
}

func (h *hashNode) Copy() Node {
	return &hashNode{hash: h.hash, weight: h.weight}
}

func (h *hashNode) CopyRoot(currLevel, collapseLevel int) Node {
	return h.Copy()
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

func (s *shortNode) Hash() []byte {
	return s.hash
}

func (s *shortNode) CalcHash() []byte {
	if !s.dirty {
		return s.hash
	}
	m := make([]byte, 0, len(s.key)+32)
	m = append(m, s.key...)
	if s.value != nil {
		m = append(m, s.value.CalcHash()...)
	}
	s.hash = encryption.RawHash(m)
	s.dirty = false
	return s.hash
}

func (s *shortNode) Weight() uint64 {
	return s.value.Weight()
}

func (s *shortNode) ToCollect() bool {
	return s.toCollect
}

func (s *shortNode) Dirty() bool {
	return s.dirty
}

func (s *shortNode) Copy() Node {
	keyCopy := make([]byte, len(s.key))
	copy(keyCopy, s.key)

	valueCopy := &hashNode{
		hash:   s.value.Hash(),
		weight: s.value.Weight(),
	}

	return &shortNode{key: keyCopy, value: valueCopy, hash: s.hash}
}

func (s *shortNode) CopyRoot(currLevel, collapseLevel int) Node {
	keyCopy := make([]byte, len(s.key))
	copy(keyCopy, s.key)
	valueCopy := s.value.CopyRoot(currLevel+1, collapseLevel)
	return &shortNode{key: keyCopy, value: valueCopy, hash: s.hash}
}

func (s *shortNode) Serialize() ([]byte, error) {
	s.toCollect = false
	if s.dirty {
		s.CalcHash()
	}
	valueHashWithWeight := make([]byte, hashWithWeightLength)
	copy(valueHashWithWeight, s.value.Hash())
	binary.BigEndian.PutUint64(valueHashWithWeight[32:], s.value.Weight())

	shortPersistNode := PersistNodeShort{
		Key:   s.key,
		Hash:  s.hash,
		Value: valueHashWithWeight,
	}
	pNode := PersistNodeBase{
		Short: &shortPersistNode,
	}

	return cbor.Marshal(&pNode)
}

func (s *shortNode) Save(batcher storage.Batcher) error {
	if !s.dirty {
		return nil
	}
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return batcher.Put(s.hash, data)
}

func DeserializeNode(data []byte) (Node, error) {
	pNode := PersistNodeBase{}
	err := cbor.Unmarshal(data, &pNode)
	if err != nil {
		return nil, err
	}
	if pNode.Branch != nil {
		branchNode := routingNode{}
		branchNode.hash = pNode.Branch.Hash
		for i, child := range pNode.Branch.Children {
			if len(child) >= hashWithWeightLength {
				childHash := child[:32]
				childWeight := binary.BigEndian.Uint64(child[32:])
				branchNode.weight += childWeight
				childNodeValue := &hashNode{hash: childHash, weight: childWeight}
				if len(child) == hashWithWeightLength {
					branchNode.Children[i] = childNodeValue
				} else {
					childNodeValue.hash = child[hashWithWeightLength : hashWithWeightLength+32]
					childKey := child[hashWithWeightLength+32:]
					branchNode.Children[i] = &shortNode{
						key:   childKey,
						hash:  childHash,
						value: childNodeValue,
					}
				}
			}
		}
		return &branchNode, nil
	}
	if pNode.Value != nil {
		valueNode := valueNode{}
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
	if pNode.Short != nil {
		shortNode := shortNode{}
		shortNode.key = pNode.Short.Key
		if len(pNode.Short.Value) != hashWithWeightLength {
			return nil, errors.New("invalid hash with weight")
		}
		shortNode.hash = pNode.Short.Hash
		hashNode := hashNode{
			hash:   pNode.Short.Value[:32],
			weight: binary.BigEndian.Uint64(pNode.Short.Value[32:]),
		}
		shortNode.value = &hashNode
		return &shortNode, nil
	}

	return nil, errors.New("invalid node")
}

func keybytesToHex(str []byte) []byte {
	l := len(str) * 2
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	return nibbles
}

// hexToKeybytes turns hex nibbles into key bytes.
// This can only be used for keys of even length.
func hexToKeybytes(hex []byte) []byte {
	key := make([]byte, len(hex)/2)
	decodeNibbles(hex, key)
	return key
}

func decodeNibbles(nibbles []byte, bytes []byte) {
	for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
		bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
	}
}
