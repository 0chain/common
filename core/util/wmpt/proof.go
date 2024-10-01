package wmpt

import (
	"errors"

	"github.com/fxamacker/cbor/v2"
)

func (t *WeightedMerkleTrie) GetBlockProof(block uint64) (key, proof []byte, err error) {
	if t.root == nil {
		return nil, nil, ErrNotFound
	}
	if block > t.root.Weight() {
		return nil, nil, ErrWeightNotInRange
	}
	persistTrie := &PersistTrie{}
	key, err = t.getBlockProof(t.root, block, nil, persistTrie)
	if err != nil {
		if errors.Is(err, ErrKVNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	key = hexToKeybytes(key)
	proof, err = cbor.Marshal(persistTrie)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (t *WeightedMerkleTrie) getBlockProof(node Node, block uint64, prefix []byte, persistTrie *PersistTrie) (key []byte, err error) {
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
		for i, child := range &n.Children {
			if child == nil {
				continue
			}
			if block <= child.Weight() {
				return t.getBlockProof(child, block, append(prefix, byte(i)), persistTrie)
			}
			block -= child.Weight()
		}
	case *shortNode:
		if block > n.Weight() {
			return nil, ErrWeightNotInRange
		}
		return t.getBlockProof(n.value, block, append(prefix, n.key...), persistTrie)
	case *valueNode:
		return prefix, nil
	case *hashNode:
		rn, err := t.resolveHashNode(n)
		if err != nil {
			return nil, err
		}
		return t.getBlockProof(rn, block, prefix, persistTrie)
	}

	return nil, ErrNotFound
}

func (t *WeightedMerkleTrie) VerifyBlockProof(block uint64, proof []byte) (hash, value []byte, err error) {
	if len(proof) == 0 {
		return nil, nil, errors.New("proof is empty")
	}
	persistTrie := &PersistTrie{}
	dm, err := cbor.DecOptions{
		MaxArrayElements: 1342177280,
		MaxMapPairs:      1342177280,
	}.DecMode()
	if err != nil {
		return nil, nil, err
	}
	err = dm.Unmarshal(proof, persistTrie)
	if err != nil {
		return nil, nil, err
	}
	if len(persistTrie.Pairs) == 0 {
		return nil, nil, errors.New("proof is empty")
	}
	ind := 0
	t.root, value, err = verifyProof(persistTrie, block, &ind)
	if err != nil {
		return nil, nil, err
	}
	hash = t.root.Hash()
	return
}

func verifyProof(persistTrie *PersistTrie, block uint64, ind *int) (Node, []byte, error) {
	if *ind >= len(persistTrie.Pairs) {
		return nil, nil, errors.New("index out of bounds")
	}

	node, err := DeserializeNode(persistTrie.Pairs[*ind].Value)
	if err != nil {
		return nil, nil, err
	}
	*ind++

	switch n := node.(type) {
	case *routingNode:
		for i := 0; i < len(&n.Children); i++ {
			if n.Children[i] == nil {
				continue
			}
			child := n.Children[i]
			if block <= child.Weight() {
				newNode, val, err := verifyProof(persistTrie, block, ind)
				if err != nil {
					return nil, nil, err
				}
				n.Children[i] = newNode
				n.dirty = true
				n.CalcHash()
				return n, val, nil
			}
			block -= child.Weight()
		}
		return nil, nil, ErrWeightNotInRange
	case *shortNode:
		if block > n.Weight() {
			return nil, nil, ErrWeightNotInRange
		}
		newNode, val, err := verifyProof(persistTrie, block, ind)
		if err != nil {
			return nil, nil, err
		}
		n.value = newNode
		n.dirty = true
		n.CalcHash()
		return n, val, nil
	case *valueNode:
		if block > n.Weight() {
			return nil, nil, ErrWeightNotInRange
		}
		n.dirty = true
		n.CalcHash()
		return n, n.value, nil
	}
	return nil, nil, errors.New("invalid node")
}
