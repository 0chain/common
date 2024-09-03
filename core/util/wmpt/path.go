package wmpt

import (
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/fxamacker/cbor/v2"
)

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

	// eg, cancel := errgroup.WithContext()

	// switch n := t.root.(type) {
	// case *nilNode:
	// 	val, err := n.Serialize()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	persistTrie.Pairs = append(persistTrie.Pairs, &PersistTriePair{
	// 		Value: val,
	// 	})
	// 	return cbor.Marshal(persistTrie)
	// case *valueNode:
	// 	val, err := n.Serialize()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	persistTrie.Pairs = append(persistTrie.Pairs, &PersistTriePair{
	// 		Value: val,
	// 	})
	// 	return cbor.Marshal(persistTrie)
	// case *routingNode:
	// 	n.dirty = true
	// }
	for _, key := range keys {
		_, err := t.markToCollect(t.root, key)
		if err != nil {
			if errors.Is(err, pebble.ErrNotFound) {
				err = ErrNotFound
			}
			return nil, err
		}
	}
	err := t.collectNodes(t.root, persistTrie)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(persistTrie)
}

func (t *WeightedMerkleTrie) markToCollect(node Node, key []byte) (Node, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *routingNode:
		prefix := commonPrefix(node.Key(), key)
		n.toCollect = true
		if len(prefix) != len(n.key) {
			return nil, nil
		}
		postfix := key[len(prefix):]
		expandedNode, err := t.markToCollect(n.Children[postfix[0]], postfix[1:])
		if err != nil {
			return nil, err
		}
		if expandedNode != nil {
			n.Children[postfix[0]] = expandedNode
		}
		return n, nil
	case *hashNode:
		data, err := t.db.Get(n.Hash())
		if err != nil {
			return nil, err
		}
		loadedNode, err := DeserializeNode(data)
		if err != nil {
			return nil, err
		}
		return t.markToCollect(loadedNode, key)
	}
	return node, nil
}

func (t *WeightedMerkleTrie) collectNodes(node Node, persistTrie *PersistTrie) error {
	if node == nil {
		node = emptyNode
	}

	if !node.ToCollect() {
		if r, ok := node.(*routingNode); ok {
			node = &hashNode{
				hash:   r.Hash(),
				weight: r.Weight(),
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
	}
	return nil
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
	return err
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
		for i := 0; i < len(n.Children); i++ {
			child, err := t.deserializeTrie(pairs, ind)
			if err != nil {
				return nil, err
			}
			n.Children[i] = child
		}
	}
	return node, nil
}
