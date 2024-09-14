package wmpt

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/common/core/util/storage"
	"golang.org/x/sync/errgroup"
)

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState          = encryption.EmptyHashBytes
	ErrNotFound         = errors.New("not found")
	ErrWeightNotInRange = errors.New("weight not in range")
	ErrInvalidKey       = errors.New("invalid key")
)

const (
	keyLength = 32
)

type WeightedMerkleTrie struct {
	root        Node
	db          storage.StorageAdapter
	oldRoot     hashNode
	deleted     [][]byte
	tempDeleted [][]byte
	created     [][]byte
	sync.Mutex
}

// New creates a new weighted merkle trie
func New(root Node, db storage.StorageAdapter) *WeightedMerkleTrie {
	if root == nil {
		root = &nilNode{}
	}
	return &WeightedMerkleTrie{db: db, root: root}
}

func (t *WeightedMerkleTrie) CopyRoot(collapseLevel int) Node {
	t.Lock()
	defer t.Unlock()
	if t.root == nil {
		return emptyNode
	}
	return t.root.CopyRoot(0, collapseLevel)
}

func (t *WeightedMerkleTrie) SetRoot(root Node) {
	t.Lock()
	defer t.Unlock()
	t.root = root
}

func (t *WeightedMerkleTrie) GetRoot() Node {
	return t.root
}

func (t *WeightedMerkleTrie) Update(key, value []byte, weight uint64) error {
	if len(key) != keyLength {
		return ErrInvalidKey
	}
	k := keybytesToHex(key)
	t.Lock()
	defer t.Unlock()
	if t.root == nil {
		t.root = emptyNode
	}
	if len(value) != 0 {
		_, n, err := t.insert(t.root, nil, k, &valueNode{value: value, weight: weight, dirty: true})
		if err != nil {
			return err
		}
		t.root = n
	} else {
		_, n, err := t.delete(t.root, nil, k)
		if err != nil {
			return err
		}
		t.root = n
		if t.root == nil {
			t.root = emptyNode
		}
	}
	return nil
}

func (t *WeightedMerkleTrie) insert(node Node, prefix, key []byte, value Node) (uint64, Node, error) {
	if len(key) == 0 {
		if v, ok := node.(*valueNode); ok {
			v.dirty = true
			c := value.Weight() - v.weight
			v.weight = value.Weight()
			return c, v, nil
		}
		return value.Weight(), value, nil
	}

	switch n := node.(type) {
	case *routingNode:
		n.dirty = true
		change, newNode, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if err != nil {
			return 0, nil, err
		}
		n.weight += change
		n.Children[key[0]] = newNode
		return change, n, nil
	case *shortNode:
		n.dirty = true
		prefixLen := commonPrefix(n.key, key)
		if prefixLen == len(n.key) {
			change, newNode, err := t.insert(n.value, append(prefix, key[:prefixLen]...), key[prefixLen:], value)
			if err != nil {
				return 0, nil, err
			}
			n.value = newNode
			return change, n, nil
		}
		branch := &routingNode{dirty: true, weight: n.Weight() + value.Weight()}
		var err error
		_, branch.Children[n.key[prefixLen]], err = t.insert(nil, append(prefix, n.key[:prefixLen+1]...), n.key[prefixLen+1:], n.value)
		if err != nil {
			return 0, nil, err
		}
		_, branch.Children[key[prefixLen]], err = t.insert(nil, append(prefix, key[:prefixLen+1]...), key[prefixLen+1:], value)
		if err != nil {
			return 0, nil, err
		}
		if prefixLen == 0 {
			return value.Weight(), branch, nil
		}
		return value.Weight(), &shortNode{key: key[:prefixLen], value: branch, dirty: true}, nil
	case *hashNode:
		rn, err := t.resolveHashNode(n)
		if err != nil {
			return 0, nil, err
		}
		return t.insert(rn, prefix, key, value)
	case nil:
		return value.Weight(), &shortNode{key: key, value: value, dirty: true}, nil
	case *nilNode:
		return value.Weight(), &shortNode{key: key, value: value, dirty: true}, nil
	default:
		return 0, nil, errors.New("unknown node type")
	}
}

func (t *WeightedMerkleTrie) delete(node Node, prefix, key []byte) (uint64, Node, error) {
	switch n := node.(type) {
	case *shortNode:
		prefixLen := commonPrefix(n.key, key)
		if prefixLen < len(n.key) {
			return 0, n, ErrNotFound
		}
		if prefixLen == len(key) {
			//delete the node
			t.tempDeleted = append(t.tempDeleted, n.Hash(), n.value.Hash())
			return n.Weight(), nil, nil
		}
		//the key is longer than the short node key, call delete on the child
		n.dirty = true
		change, newNode, err := t.delete(n.value, append(prefix, key[:len(n.key)]...), key[len(n.key):])
		if err != nil {
			return 0, nil, err
		}
		switch child := newNode.(type) {
		case *shortNode:
			//merge the short node
			newKey := make([]byte, len(n.key)+len(child.key))
			copy(newKey, n.key)
			copy(newKey[len(n.key):], child.key)
			n.key = newKey
			n.value = child.value
			return change, n, nil
		default:
			n.value = newNode
			return change, n, nil
		}
	case *routingNode:
		change, newNode, err := t.delete(n.Children[key[0]], append(prefix, key[0]), key[1:])
		if err != nil {
			return 0, nil, err
		}
		n.dirty = true
		n.Children[key[0]] = newNode
		n.weight -= change
		//if child is not nil, we can return the branching node as it has at least 2 children
		if newNode != nil {
			return change, n, nil
		}
		//Reduction:
		// check if n has only one child, if so, merge it with the child
		pos := -1
		for i, child := range &n.Children {
			if child != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			cnode, err := t.resolve(n.Children[pos])
			if err != nil {
				return 0, nil, err
			}
			// merge if the child is short node
			if cnode, ok := cnode.(*shortNode); ok {
				newKey := make([]byte, len(cnode.key)+1)
				newKey[0] = byte(pos)
				copy(newKey[1:], cnode.key)
				t.tempDeleted = append(t.tempDeleted, n.Hash(), cnode.Hash())
				newShortNode := &shortNode{
					key:   newKey,
					value: cnode.value,
					dirty: true,
				}
				return change, newShortNode, nil
			}
		}
		return change, n, nil
	case *valueNode:
		t.tempDeleted = append(t.tempDeleted, n.Hash())
		return n.weight, nil, nil
	case nil:
		return 0, nil, ErrNotFound
	case *nilNode:
		return 0, nil, ErrNotFound
	case *hashNode:
		rn, err := t.resolveHashNode(n)
		if err != nil {
			return 0, nil, err
		}
		return t.delete(rn, prefix, key)
	default:
		return 0, nil, errors.New("unknown node type")
	}
}

// Resolves the node by loading it from the database if it is a hash node, otherwise it returns the node
func (t *WeightedMerkleTrie) resolve(node Node) (Node, error) {
	if n, ok := node.(*hashNode); ok {
		if t.db == nil {
			return node, nil
		}
		return t.resolveHashNode(n)
	}
	return node, nil
}

func (t *WeightedMerkleTrie) resolveHashNode(node *hashNode) (Node, error) {
	if t.db == nil {
		return nil, errors.New("database is not set")
	}
	data, err := t.db.Get(node.Hash())
	if err != nil {
		return nil, err
	}
	loadedNode, err := DeserializeNode(data)
	return loadedNode, err
}

// Put puts a key-value pair into the trie
func (t *WeightedMerkleTrie) Put(key, value []byte, weight uint64) error {
	k := keybytesToHex(key)
	_, newNode, err := t.insert(t.root, nil, k, &valueNode{value: value, weight: weight, dirty: true})
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
	if len(t.oldRoot.hash) > 0 {
		t.root = &hashNode{
			hash:   t.oldRoot.hash,
			weight: t.oldRoot.weight,
		}
	} else {
		t.root = emptyNode
	}
	t.tempDeleted = nil
	t.DeleteNodes() //nolint:errcheck
}

// DeleteNodes deletes the nodes from the underlying storage and sets nextDelete to the tempDeleted nodes collected in previous mutations
func (t *WeightedMerkleTrie) DeleteNodes() error {
	if len(t.deleted) > 0 {
		batcher := t.db.NewBatch()
		for _, key := range t.deleted {
			err := batcher.Delete(key)
			if err != nil {
				return err
			}
		}
		err := batcher.Commit(false)
		if err != nil {
			return err
		}
	}
	t.deleted = t.tempDeleted
	t.tempDeleted = nil
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

func (t *WeightedMerkleTrie) Weight() uint64 {
	if t.root == nil {
		return 0
	}
	return t.root.Weight()
}

// Commit collapses the trie to the specified level and returns the batcher and the deleted nodes, it is the caller's responsibility to commit the batch
func (t *WeightedMerkleTrie) Commit(collapseLevel int) (storage.Batcher, error) {
	batcher := t.db.NewBatch()
	root, ok := t.root.(*routingNode)
	deleteChan := make(chan []byte, 10)
	createdChan := make(chan []byte, 10)
	defer close(deleteChan)
	defer close(createdChan)
	t.collectDeleteAndCreated(deleteChan, createdChan)
	if ok {
		eg, _ := errgroup.WithContext(context.Background())
		eg.SetLimit(5)
		deleteChan <- root.Hash()
		for i := 0; i < len(&root.Children); i++ {
			if root.Children[i] == nil || !root.Children[i].Dirty() {
				continue
			}
			ind := i
			eg.Go(func() error {
				var collapsedNode Node
				collapsedNode, err := t.commit(root.Children[ind], batcher, collapseLevel, 1, deleteChan, createdChan)
				if err != nil {
					return err
				}
				if collapsedNode != nil {
					root.Children[ind] = collapsedNode
				}
				return nil
			})
		}
		err := eg.Wait()
		if err != nil {
			return nil, err
		}
		err = root.Save(batcher)
		if err != nil {
			return nil, err
		}
		createdChan <- root.Hash()
		t.root = root
		return batcher, nil
	}
	node, err := t.commit(t.root, batcher, collapseLevel, 0, deleteChan, createdChan)
	if err != nil {
		return nil, err
	}
	t.root = node
	return batcher, nil
}

func (t *WeightedMerkleTrie) RollbackTrie(node Node) {
	if node == nil {
		node = emptyNode
	} else if bytes.Equal(node.Hash(), t.root.Hash()) {
		return
	}
	t.root = node
	if len(t.created) > 0 {
		batcher := t.db.NewBatch()
		for _, hash := range t.created {
			_ = batcher.Delete(hash)
		}
		batcher.Commit(false) //nolint:errcheck
	}
	t.created = nil
	t.deleted = nil
}

func (t *WeightedMerkleTrie) Delete(key []byte) (uint64, error) {
	if t.root == nil {
		return 0, ErrNotFound
	}
	k := keybytesToHex(key)
	change, node, err := t.delete(t.root, nil, k)
	if err != nil {
		return 0, err
	}
	t.root = node
	if t.root == nil {
		t.root = emptyNode
	}
	return change, nil
}

func (t *WeightedMerkleTrie) commit(node Node, batcher storage.Batcher, collapseLevel, level int, deleteChan, createdChan chan []byte) (Node, error) {
	if node == nil {
		return nil, nil
	}

	if !node.Dirty() {
		return node, nil
	}
	deleteChan <- node.Hash()
	var err error
	switch n := node.(type) {
	case *routingNode:
		for i := 0; i < len(n.Children); i++ {
			if n.Children[i] == nil || !n.Children[i].Dirty() {
				continue
			}
			var collapsedNode Node
			collapsedNode, err = t.commit(n.Children[i], batcher, collapseLevel, level+1, deleteChan, createdChan)
			if err != nil {
				return nil, err
			}
			if collapsedNode != nil {
				n.Children[i] = collapsedNode
			}
		}
		err = n.Save(batcher)
		if err != nil {
			return nil, err
		}
		if level == collapseLevel {
			n.Children = [16]Node{}
			return &hashNode{
				hash:   n.Hash(),
				weight: n.Weight(),
			}, nil
		}
		createdChan <- n.Hash()
		return n, nil
	case *shortNode:
		collapsedNode, err := t.commit(n.value, batcher, collapseLevel, level+1, deleteChan, createdChan)
		if err != nil {
			return nil, err
		}
		if collapsedNode != nil {
			n.value = collapsedNode
		}
		err = n.Save(batcher)
		if err != nil {
			return nil, err
		}
		if level == collapseLevel {
			return &hashNode{
				hash:   n.Hash(),
				weight: n.Weight(),
			}, nil
		}
		createdChan <- n.Hash()
		return n, nil
	case *valueNode:
		err = n.Save(batcher)
		if err != nil {
			return nil, err
		}
		createdChan <- n.Hash()
		return n, nil
	}

	return node, nil
}

func commonPrefix(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func (t *WeightedMerkleTrie) collectDeleteAndCreated(deleteChan, createdChan chan []byte) {
	t.tempDeleted = nil
	t.created = nil
	go func() {
		for hash := range deleteChan {
			if len(hash) > 0 {
				t.tempDeleted = append(t.tempDeleted, hash)
			}
		}
	}()
	go func() {
		for hash := range createdChan {
			t.created = append(t.created, hash)
		}
	}()
}
