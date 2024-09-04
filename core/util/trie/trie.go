package trie

import (
	"errors"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/logging"
)

var (
	// emptyRoot is the known root hash of an empty trie.
	emptyRoot = encryption.EmptyHashBytes

	// emptyState is the known hash of an empty state trie entry.
	emptyState = encryption.EmptyHashBytes
	NIL        = &nilNode{}
)

//go:generate msgp -v -io=false -tests=false -unexported=true

// As we have fixed key length we will benefit several simplifications
// - No mixed nodes, we have leaf or routing Node only
type FixedLengthHexKeyMerkleTrie struct {
	root Node `msgpack:"n"`
}

func New() *FixedLengthHexKeyMerkleTrie {
	trie := FixedLengthHexKeyMerkleTrie{root: &nilNode{}}
	return &trie
}

func (t *FixedLengthHexKeyMerkleTrie) Get(key []byte) (val []byte, found bool) {
	n := find(t.root, key)
	if n == nil || n == NIL {
		return nil, false
	}
	if curNode, ok := n.(*valueNode); ok {
		return curNode.value, true
	} else {
		logging.Logger.Error("error while looking for Node")
		return nil, false
	}

}

func (t *FixedLengthHexKeyMerkleTrie) Proof() []byte {
	if t.root == nil {
		return emptyState
	}
	return t.root.Hash()
}

func find(n Node, key []byte) Node {
	if n == nil {
		return nil
	}

	prefix := CommonPrefix(n.Key(), key)
	postfix := key[len(prefix):]

	if Equal(prefix, key) {
		return n
	} else {
		switch n.(type) {
		case *routingNode:
			curNode := n.(*routingNode)
			ind := findIndex(postfix[0])
			nextNode := curNode.Children[ind]
			return find(nextNode, postfix)
		case *valueNode:
			curNode := n.(*valueNode)
			if Equal(prefix, postfix) {
				return curNode
			} else {
				return nil
			}
		case *nilNode:
			return nil
		}
	}
	return nil
}

// key is a hex encoded byte key
func (t *FixedLengthHexKeyMerkleTrie) InsertOrUpdate(key []byte, weight uint64, value []byte) {
	if t.root == nil { //Node not found, can be only if trie is empty
		t.root = &valueNode{key: key, weight: weight, value: value}
		t.root.CalcHash()
		return
	}

	_, n := insert(t.root, key, weight, value)

	if n != nil {
		t.root = n
	}

}

// delete
func (t *FixedLengthHexKeyMerkleTrie) Delete(key []byte) Node {
	if t.root == nil { //Node not found, can be only if trie is empty
		return nil
	}

	_, n := del(t.root, key)
	return n
}

func (t *FixedLengthHexKeyMerkleTrie) Copy() *FixedLengthHexKeyMerkleTrie {
	tcopy := t.root.Copy()

	return &FixedLengthHexKeyMerkleTrie{root: tcopy}
}

func insert(n Node, key []byte, weight uint64, value []byte) (change uint64, newNode Node) {
	if n == NIL {
		vNode := &valueNode{key: key, weight: weight, value: value}
		vNode.CalcHash()
		return weight, vNode
	}

	prefix := CommonPrefix(n.Key(), key)
	postfix := key[len(prefix):]

	switch n.(type) {
	case *nilNode:
		vNode := &valueNode{key: key, weight: weight, value: value}
		vNode.CalcHash()
		return weight, vNode
	case *routingNode:
		curNode := n.(*routingNode)
		ind := findIndex(postfix[0])
		nextNode := curNode.Children[ind]
		c, newNode2 := insert(nextNode, postfix, weight, value)
		if newNode2 != NIL {
			curNode.Children[ind] = newNode2
			curNode.weight += c
			curNode.CalcHash()
			return c, curNode
		}
	case *valueNode:
		current := n.(*valueNode)
		if Equal(prefix, key) { //exact match, update value
			current.value = value
			c := weight - current.weight
			current.weight = weight
			current.CalcHash()
			return c, current
		} else { //split Node
			nodes := [16]Node{}
			for i := range nodes {
				nodes[i] = NIL
			}
			rNode := &routingNode{key: prefix, Children: nodes}
			postfixS := n.Key()[len(prefix):]

			ind := findIndex(postfix[0])
			indS := findIndex(postfixS[0])
			vNode := &valueNode{key: postfix, weight: weight, value: value}
			vNode.CalcHash()
			rNode.Children[ind] = vNode
			current.key = postfixS
			rNode.Children[indS] = current
			rNode.weight = vNode.weight + current.weight
			rNode.CalcHash()
			return weight, rNode
		}
	default:
		panic("wrong Node type")
	}

	return 0, NIL
}

func del(n Node, key []byte) (propagate bool, node Node) {
	if n == nil {
		return false, nil
	}

	prefix := CommonPrefix(n.Key(), key)
	postfix := key[len(prefix):]

	switch n.(type) {
	case *nilNode:
		return false, nil
	case *routingNode:
		curNode := n.(*routingNode)
		ind := findIndex(postfix[0])
		nextNode := curNode.Children[ind]
		if nextNode == NIL {
			return false, NIL
		}

		prop, deleted := del(nextNode, postfix)
		if deleted == NIL {
			return false, NIL
		}

		curNode.weight -= deleted.Weight()

		if prop {
			curNode.Children[ind] = NIL
			empty := true
			for _, child := range curNode.Children {
				if child != NIL {
					empty = false
				}
			}
			if !empty {
				curNode.CalcHash()
			}
			return empty, deleted
		}
		curNode.CalcHash()

	case *valueNode:
		if Equal(prefix, key) { //exact match, delete value
			return true, n
		} else { //split Node
			return false, NIL
		}
	default:
		panic("wrong Node type")
	}

	return false, NIL
}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func CommonPrefix(a, b []byte) []byte {
	var c []byte
	for i, v := range a {
		if v == b[i] {
			c = append(c, v)
		} else {
			return c
		}
	}

	return c
}

// Ordered array made with left-right-root walk
func (t *FixedLengthHexKeyMerkleTrie) Values() (values [][]byte) {

	return aggregate(values, t.root)
}

// Ordered array made with left-right-root walk
func (t *FixedLengthHexKeyMerkleTrie) FloorNodeValue(number uint64) (value []byte, err error) {
	if number > t.root.Weight() {
		return nil, errors.New("value is not in the range")
	}
	f := floorNode(t.root, number)
	return f.(*valueNode).value, nil
}

func (t *FixedLengthHexKeyMerkleTrie) Serialize() ([]byte, error) {
	return t.MarshalMsg(nil)
}

func Deserialize(b []byte) (*FixedLengthHexKeyMerkleTrie, error) {
	t := &FixedLengthHexKeyMerkleTrie{root: &routingNode{}}
	_, err := t.UnmarshalMsg(b)
	return t, err
}

func aggregate(source [][]byte, n Node) [][]byte {
	if n == NIL {
		return source
	}
	switch n.(type) {
	case *nilNode:
		return source
	case *valueNode:
		return append(source, n.(*valueNode).value)
	case *routingNode:
		for _, child := range n.(*routingNode).Children {
			source = aggregate(source, child)
		}
		return source
	}

	return nil
}

// Ordered array made with left-right-root walk
func (t *FixedLengthHexKeyMerkleTrie) Weights() (values []uint64) {

	return aggregateWeights(values, t.root)
}

func aggregateWeights(source []uint64, n Node) []uint64 {
	if n == NIL {
		return source
	}
	switch n.(type) {
	case *nilNode:
		return source
	case *valueNode:
		return append(source, n.(*valueNode).weight)
	case *routingNode:
		source = append(source, n.(*routingNode).weight)
		for _, child := range n.(*routingNode).Children {
			source = aggregateWeights(source, child)
		}
		return source
	}

	return nil
}

// Ordered array made with left-right-root walk
func (t *FixedLengthHexKeyMerkleTrie) Hashes() (values [][]byte) {

	return aggregateHashes(values, t.root)
}

func floorNode(n Node, number uint64) Node {
	if n == NIL {
		return NIL
	}
	switch n.(type) {
	case *nilNode:
		return NIL
	case *valueNode:
		return n
	case *routingNode:
		r := n.(*routingNode).Children
		for _, child := range r {
			if child == NIL {
				continue
			}
			if number < child.Weight() {
				return floorNode(child, number)
			}
			number -= child.Weight()
		}
	}
	return NIL
}

func aggregateHashes(source [][]byte, n Node) [][]byte {
	if n == NIL {
		return source
	}
	switch n.(type) {
	case *nilNode:
		return source
	case *valueNode:
		return append(source, n.(*valueNode).hash)
	case *routingNode:
		source = append(source, n.(*routingNode).hash)
		for _, child := range n.(*routingNode).Children {
			source = aggregateHashes(source, child)
		}
		return source
	}

	return nil
}
