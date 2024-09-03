package wmpt

import (
	"bytes"
	"errors"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/util/storage"
)

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState          = encryption.EmptyHashBytes
	ErrNotFound         = errors.New("not found")
	ErrWeightNotInRange = errors.New("weight not in range")
)

type WeightedMerkleTrie struct {
	root    Node
	db      storage.StorageAdapter
	oldRoot hashNode
	deleted [][]byte
}

// New creates a new weighted merkle trie
func New(root Node, db storage.StorageAdapter) *WeightedMerkleTrie {
	if root == nil {
		root = &nilNode{}
	}
	return &WeightedMerkleTrie{db: db, root: root}
}

// Put puts a key-value pair into the trie
func (t *WeightedMerkleTrie) Put(key, value []byte, weight uint64) error {
	_, newNode, err := t.put(t.root, key, value, weight)
	if err != nil {
		return err
	}
	t.root = newNode
	return nil
}

// SaveRoot saves the current root to the old root
func (t *WeightedMerkleTrie) SaveRoot() {
	if t.root != nil {
		t.oldRoot.hash = t.root.Hash()
		t.oldRoot.weight = t.root.Weight()
	}
}

// Rollback rolls back the trie to the previous root
func (t *WeightedMerkleTrie) Rollback() {
	t.root = &t.oldRoot
	t.deleted = nil
}

// DeleteNodes deletes the nodes from the underlying storage and sets nextDelete to the next nodes to be deleted
func (t *WeightedMerkleTrie) DeleteNodes(nextDelete [][]byte) error {
	batcher := t.db.NewBatch()
	for _, key := range t.deleted {
		err := batcher.Delete(key)
		if err != nil {
			return err
		}
	}
	if len(t.deleted) > 0 {
		err := batcher.Commit(false)
		if err != nil {
			return err
		}
	}
	t.deleted = nextDelete
	return nil
}

// Root returns the root hash of the trie, if the root is dirty, it will recalculate the hash
func (t *WeightedMerkleTrie) Root() []byte {
	if t.root == nil {
		return emptyState
	}
	if t.root.Dirty() {
		return t.root.CalcHash()
	}
	return t.root.Hash()
}

// Commit collapses the trie to the specified level and returns the batcher and the deleted nodes, it is the caller's responsibility to commit the batch
func (t *WeightedMerkleTrie) Commit(collapseLevel int) (storage.Batcher, [][]byte, error) {
	batcher := t.db.NewBatch()
	nextDelete := make([][]byte, 0, 10)
	node, nextDelete, err := t.commit(t.root, batcher, collapseLevel, 0, nextDelete)
	if err != nil {
		return nil, nil, err
	}
	t.root = node
	return batcher, nextDelete, nil
}

func (t *WeightedMerkleTrie) Delete(key []byte) (deletedKeys [][]byte, err error) {
	if t.root == nil {
		return nil, ErrNotFound
	}

	_, node, deletedKeys, err := t.delete(t.root, key, nil)
	if err != nil {
		return nil, err
	}
	t.root = node
	if t.root == nil {
		t.root = emptyNode
	}
	return deletedKeys, nil
}

func (t *WeightedMerkleTrie) delete(node Node, key []byte, deletedKeys [][]byte) (uint64, Node, [][]byte, error) {
	if node == nil {
		return 0, nil, deletedKeys, ErrNotFound
	}

	prefix := commonPrefix(node.Key(), key)
	postfix := key[len(prefix):]

	var (
		err         error
		deletedNode Node
		change      uint64
	)
	switch n := node.(type) {
	case *routingNode:
		if len(prefix) != len(n.key) {
			return 0, nil, deletedKeys, ErrNotFound
		}
		n.dirty = true
		change, deletedNode, deletedKeys, err = t.delete(n.Children[postfix[0]], postfix[1:], deletedKeys)
		if err != nil {
			return 0, nil, deletedKeys, err
		}
		n.Children[postfix[0]] = deletedNode
		n.weight -= change
		if n.weight == 0 {
			deletedKeys = append(deletedKeys, n.Hash())
			return change, nil, deletedKeys, nil
		}
		return change, n, deletedKeys, nil
	case *valueNode:
		if bytes.Equal(prefix, key) {
			deletedKeys = append(deletedKeys, n.Hash())
			return n.weight, nil, deletedKeys, nil
		}
		return 0, nil, deletedKeys, ErrNotFound
	case *hashNode:
		data, err := t.db.Get(n.Hash())
		if err != nil {
			return 0, nil, deletedKeys, err
		}
		loadedNode, err := DeserializeNode(data)
		if err != nil {
			return 0, nil, deletedKeys, err
		}
		return t.delete(loadedNode, key, deletedKeys)
	}
	return 0, nil, deletedKeys, ErrNotFound
}

func (t *WeightedMerkleTrie) commit(node Node, batcher storage.Batcher, collapseLevel, level int, nextDelete [][]byte) (Node, [][]byte, error) {
	if node == nil {
		return nil, nextDelete, nil
	}

	if !node.Dirty() {
		return node, nextDelete, nil
	}

	delHash := make([]byte, 32)
	copy(delHash, node.Hash())
	nextDelete = append(nextDelete, delHash)
	var err error
	switch n := node.(type) {
	case *routingNode:
		for i := 0; i < len(n.Children); i++ {
			if n.Children[i] == nil || !n.Children[i].Dirty() {
				continue
			}
			var collapsedNode Node
			collapsedNode, nextDelete, err = t.commit(n.Children[i], batcher, collapseLevel, level+1, nextDelete)
			if err != nil {
				return nil, nil, err
			}
			if collapsedNode != nil {
				n.Children[i] = collapsedNode
			}
		}
		err = n.Save(batcher)
		if err != nil {
			return nil, nil, err
		}
		if level == collapseLevel {
			n.Children = [16]Node{}
			return &hashNode{
				hash:   n.Hash(),
				weight: n.Weight(),
			}, nextDelete, nil
		}
		return n, nextDelete, nil
	case *valueNode:
		err = n.Save(batcher)
		if err != nil {
			return nil, nil, err
		}
		return n, nextDelete, nil
	}

	return node, nextDelete, nil
}

func (t *WeightedMerkleTrie) put(node Node, key, value []byte, weight uint64) (uint64, Node, error) {
	if node == nil {
		vNode := &valueNode{key: key, weight: weight, value: value, dirty: true}
		return weight, vNode, nil
	}

	prefix := commonPrefix(node.Key(), key)
	postfix := key[len(prefix):]

	switch n := node.(type) {
	case *nilNode:
		return weight, &valueNode{key: key, weight: weight, value: value, dirty: true}, nil
	case *routingNode:
		n.dirty = true
		//split the routing node into two routing nodes
		if len(prefix) != len(n.key) {
			newNode := &routingNode{
				key:    prefix,
				weight: n.weight,
				dirty:  true,
			}
			n.key = n.key[len(prefix):]
			newNode.Children[n.key[0]] = n
			c, newValueNode, err := t.put(newNode.Children[postfix[0]], postfix[1:], value, weight)
			if err != nil {
				return 0, nil, err
			}
			newNode.weight += c
			newNode.Children[postfix[0]] = newValueNode
			return weight, newNode, nil
		} else {
			c, newValueNode, err := t.put(n.Children[postfix[0]], postfix[1:], value, weight)
			if err != nil {
				return 0, nil, err
			}
			n.weight += c
			n.Children[postfix[0]] = newValueNode
			return weight, n, nil
		}
	case *valueNode:
		n.dirty = true
		if bytes.Equal(prefix, key) {
			n.value = value
			c := weight - n.weight
			n.weight = weight
			return c, n, nil
		} else {
			//split the value node
			newRoutingNode := &routingNode{
				key:    prefix,
				dirty:  true,
				weight: n.weight,
			}
			n.key = n.key[len(prefix):]
			keyInd := n.key[0]
			n.key = n.key[1:]
			newRoutingNode.Children[keyInd] = n
			c, newValueNode, err := t.put(newRoutingNode.Children[postfix[0]], postfix[1:], value, weight)
			if err != nil {
				return 0, nil, err
			}
			newRoutingNode.weight += c
			newRoutingNode.Children[postfix[0]] = newValueNode
			return weight, newRoutingNode, nil
		}
	case *hashNode:
		data, err := t.db.Get(n.Hash())
		if err != nil {
			return 0, nil, err
		}
		loadedNode, err := DeserializeNode(data)
		if err != nil {
			return 0, nil, err
		}
		return t.put(loadedNode, key, value, weight)
	}
	return 0, nil, nil
}

func commonPrefix(a, b []byte) []byte {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	if minLen == 0 {
		return nil
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:minLen]
}
