package wmpt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/sync/errgroup"
)

var ErrKVNotFound = errors.New("pebble: not found")

func (t *WeightedMerkleTrie) GetPath(keys [][]byte) ([]byte, error) {
	persistTrie := &PersistTrie{}

	if t.root != nil {
		if node, ok := t.root.(*hashNode); ok {
			data, err := t.db.Get(node.Hash())
			if err != nil {
				return nil, err
			}
			loadedNode, err := DeserializeNode(data)
			if err != nil {

			}
			t.root = loadedNode
		}
	}
	now := time.Now()

	if len(keys) > 10 {
		eg, _ := errgroup.WithContext(context.TODO())
		eg.SetLimit(5)
		if node, ok := t.root.(*routingNode); ok {
			node.toCollect = true
			var branchMu = [16]sync.Mutex{}
			for i := 0; i < len(keys); i++ {
				ind := i
				eg.Go(func() error {
					k := keybytesToHex(keys[ind])
					branchMu[k[0]].Lock()
					defer branchMu[k[0]].Unlock()
					child, err := t.markToCollect(node.Children[k[0]], k, 1)
					if err != nil {
						if errors.Is(err, ErrKVNotFound) {
							err = ErrNotFound
						}
						return err
					}
					node.Children[k[0]] = child
					return nil
				})
			}
			err := eg.Wait()
			if err != nil {
				return nil, err
			}
		}
	} else {
		for _, key := range keys {
			k := keybytesToHex(key)
			_, err := t.markToCollect(t.root, k, 0)
			if err != nil {
				if errors.Is(err, ErrKVNotFound) {
					err = ErrNotFound
				}
				return nil, err
			}
		}
	}
	elapsedMark := time.Since(now)
	err := t.collectNodes(t.root, persistTrie)
	if err != nil {
		return nil, err
	}
	elapsedCollect := time.Since(now) - elapsedMark
	data, err := cbor.Marshal(persistTrie)
	fmt.Println("getPath", "elapsedMark: ", elapsedMark.Milliseconds(), "elapsedColect", elapsedCollect.Milliseconds(), "total", time.Since(now).Milliseconds())
	return data, err
}

func (t *WeightedMerkleTrie) collectNodes(node Node, persistTrie *PersistTrie) error {
	if node == nil {
		return nil
	}

	if !node.ToCollect() {
		if r, ok := node.(*routingNode); ok {
			node = &hashNode{
				hash:   r.hash,
				weight: r.weight,
			}
		}
	}

	data, err := node.Serialize()
	if err != nil {
		return err
	}
	persistTrie.Pairs = append(persistTrie.Pairs, &PersistTriePair{
		Value: data,
	})

	switch n := node.(type) {
	case *routingNode:
		for _, child := range n.Children {
			t.collectNodes(child, persistTrie)
		}
	case *shortNode:
		t.collectNodes(n.value, persistTrie)
	}
	return nil
}

func (t *WeightedMerkleTrie) markToCollect(node Node, key []byte, pos int) (Node, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *routingNode:
		child, err := t.markToCollect(n.Children[key[pos]], key, pos+1)
		if err != nil {
			return nil, err
		}
		n.Children[key[pos]] = child
		n.toCollect = true
		return n, nil
	case *shortNode:
		n.toCollect = true
		if len(key)-pos < len(n.key) || !bytes.Equal(n.key, key[pos:pos+len(n.key)]) {
			return n, nil
		}
		child, err := t.markToCollect(n.value, key, pos+len(n.key))
		if err != nil {
			return nil, err
		}
		n.value = child
		return n, nil
	case *hashNode:
		rn, err := t.resolveHashNode(n)
		if err != nil {
			return nil, err
		}
		return t.markToCollect(rn, key, pos)
	}
	return node, nil
}

func (t *WeightedMerkleTrie) Deserialize(data []byte) error {
	persistTrie := &PersistTrie{}
	dm, err := cbor.DecOptions{
		MaxArrayElements: 1342177280,
		MaxMapPairs:      1342177280,
	}.DecMode()
	if err != nil {
		return err
	}
	err = dm.Unmarshal(data, persistTrie)
	if err != nil {
		return err
	}
	if len(persistTrie.Pairs) == 0 {
		return nil
	}
	ind := 0
	t.root, err = t.deserializeTrie(persistTrie.Pairs, &ind)
	if err != nil {
		return err
	}
	hash := t.root.Hash()
	switch n := t.root.(type) {
	case *routingNode:
		n.dirty = true
		n.CalcHash()
	case *shortNode:
		n.dirty = true
		n.CalcHash()
	}
	if !bytes.Equal(hash, t.root.Hash()) {
		return errors.New("root hash mismatch")
	}
	return nil
}

func (t *WeightedMerkleTrie) deserializeTrie(pairs []*PersistTriePair, ind *int) (Node, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	if *ind >= len(pairs) {
		return nil, errors.New("index out of bounds")
	}

	node, err := DeserializeNode(pairs[*ind].Value)
	if err != nil {
		return nil, err
	}
	*ind++

	switch n := node.(type) {
	case *routingNode:
		for i := 0; i < branchNodeLength; i++ {
			if n.Children[i] != nil {
				child, err := t.deserializeTrie(pairs, ind)
				if err != nil {
					return nil, err
				}
				if !bytes.Equal(n.Children[i].Hash(), child.Hash()) {
					return nil, errors.New("child hash mismatch")
				}
				n.Children[i] = child
			}
		}
	case *shortNode:
		child, err := t.deserializeTrie(pairs, ind)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(n.value.Hash(), child.Hash()) {
			return nil, errors.New("child hash mismatch")
		}
		n.value = child
	}
	return node, nil
}
