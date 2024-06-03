package verkletrie

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-verkle"
)

const DBVerkleNodeKeyPrefix = "verkle_node_"

type Hash [32]byte

type VerkleTrie struct {
	db   DB
	key  string // the key in db where the whole serialized verkle trie to be persisted. Could be the allocation id
	root verkle.VerkleNode
}

func VKey(key string) []byte {
	return []byte(DBVerkleNodeKeyPrefix + key)
}

func New(key string, db DB) *VerkleTrie {
	var root verkle.VerkleNode
	payload, err := db.Get(VKey(key))
	if err != nil {
		return &VerkleTrie{root: verkle.New(), db: db}
	}

	fmt.Println("payload:", payload)

	root, err = verkle.ParseNode(payload, 0)
	if err != nil {
		panic(err)
	}

	return &VerkleTrie{
		db:   db,
		root: root,
	}
}

func (m *VerkleTrie) nodeResolver(key []byte) ([]byte, error) {
	return m.db.Get(key)
}

func (m *VerkleTrie) Get(key []byte) ([]byte, error) {
	return m.root.Get(key, m.nodeResolver)
}

func (m *VerkleTrie) Insert(key []byte, value []byte) error {
	return m.root.Insert(key, value, m.nodeResolver)
}

func (m *VerkleTrie) Delete(key []byte) (bool, error) {
	return m.root.Delete(key, m.nodeResolver)
}

func (m *VerkleTrie) Hash() Hash {
	return m.root.Commit().Bytes()
}

func (m *VerkleTrie) Commit(saveTrieToDB bool) (Hash, error) {
	root, ok := m.root.(*verkle.InternalNode)
	if !ok {
		return Hash{}, errors.New("unexpected root node type")
	}

	nodes, err := root.BatchSerialize()
	if err != nil {
		return Hash{}, fmt.Errorf("serializing tree nodes: %s", err)
	}

	batch := m.db.NewBatch()
	path := make([]byte, 0, len(DBVerkleNodeKeyPrefix)+32)
	path = append(path, []byte(DBVerkleNodeKeyPrefix)...)
	for _, node := range nodes {
		path := append(path[:len(DBVerkleNodeKeyPrefix)], node.Path...)
		fmt.Println("path:", string(path))

		if err := batch.Put(path, node.SerializedBytes); err != nil {
			return Hash{}, fmt.Errorf("put node to disk: %s", err)
		}

		if batch.Size() >= IdealBatchSize {
			if err := batch.Write(); err != nil {
				return Hash{}, fmt.Errorf("batch write error: %s", err)
			}
			batch.Reset()
		}
	}
	fmt.Println("nodes:", len(nodes))

	if saveTrieToDB {
		v, err := m.root.Serialize()
		if err != nil {
			return Hash{}, fmt.Errorf("serializing root node: %s", err)
		}
		if err := batch.Put(VKey(m.key), v); err != nil {
			return Hash{}, fmt.Errorf("put root node to disk: %s", err)
		}
	}

	if err := batch.Write(); err != nil {
		return Hash{}, fmt.Errorf("batch write error: %s", err)
	}

	return m.Hash(), nil
}

type Keylist [][]byte

func MakeProof(trie *VerkleTrie, keys Keylist) (*verkle.VerkleProof, verkle.StateDiff, error) {
	proof, _, _, _, err := verkle.MakeVerkleMultiProof(trie.root, nil, keys, trie.nodeResolver)
	if err != nil {
		return nil, nil, err
	}

	return verkle.SerializeProof(proof)
}

func VerifyProof(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte, keys Keylist) error {
	dproof, err := verkle.DeserializeProof(vp, stateDiff)
	if err != nil {
		return fmt.Errorf("verkle proof deserialization error: %w", err)
	}

	root := new(verkle.Point)
	if err := root.SetBytes(stateRoot); err != nil {
		return fmt.Errorf("verkle root deserialization error: %w", err)
	}

	tree, err := verkle.PreStateTreeFromProof(dproof, root)
	if err != nil {
		return fmt.Errorf("error rebuilding the pre-tree from proof: %w", err)
	}

	return verkle.VerifyVerkleProofWithPreState(dproof, tree)
}
