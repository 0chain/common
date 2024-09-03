package wmpt

import (
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/fxamacker/cbor/v2"
)

func (t *WeightedMerkleTrie) GetBlockProof(block uint64) (key, proof []byte, err error) {
	if t.root == nil {
		return nil, nil, ErrNotFound
	}
	if block >= t.root.Weight() {
		return nil, nil, ErrWeightNotInRange
	}
	persistTrie := &PersistTrie{}
	key, err = t.getBlockProof(t.root, block, persistTrie)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	proof, err = cbor.Marshal(persistTrie)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (t *WeightedMerkleTrie) getBlockProof(node Node, block uint64, persistTrie *PersistTrie) (key []byte, err error) {
	if node == nil {
		return nil, ErrNotFound
	}

	data, err := node.Serialize()
	if err != nil {
		return nil, err
	}
	persistTrie.Pairs = append(persistTrie.Pairs, &PersistTriePair{
		Value: data,
	})

	switch n := node.(type) {
	case *routingNode:
		for _, child := range n.Children {
			if child == nil {
				continue
			}
			if block < child.Weight() {
				return t.getBlockProof(child, block, persistTrie)
			}
			block -= child.Weight()
		}
	case *valueNode:
		if block > n.weight {
			return nil, ErrWeightNotInRange
		}
		return n.key, nil
	case *hashNode:
		data, err := t.db.Get(n.Hash())
		if err != nil {
			return nil, err
		}
		loadedNode, err := DeserializeNode(data)
		if err != nil {
			return nil, err
		}
		return t.getBlockProof(loadedNode, block, persistTrie)
	}

	return nil, ErrNotFound
}
