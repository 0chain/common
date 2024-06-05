package verkletrie

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ethereum/go-verkle"
	"github.com/holiman/uint256"
)

var DBVerkleNodeKeyPrefix = []byte("verkle_node_")

var ChunkSize = uint256.NewInt(32) // 32

var ErrNodeNotFound = errors.New("node not found")

type (
	Hash      [32]byte
	Address32 [32]byte
)

type VerkleTrie struct {
	db      DB
	rootKey []byte // the key in db where the whole serialized verkle trie to be persisted. Could be the allocation id
	root    verkle.VerkleNode
}

func rootKey(key string) []byte {
	return append(DBVerkleNodeKeyPrefix, key...)
}

func New(key string, db DB) *VerkleTrie {
	var root verkle.VerkleNode
	rootKey := rootKey(key)

	payload, err := db.Get([]byte(rootKey))
	if err != nil {
		if err == ErrNodeNotFound {
			return &VerkleTrie{root: verkle.New(), rootKey: rootKey, db: db}
		}

		panic(err)
	}

	root, err = verkle.ParseNode(payload, 0)
	if err != nil {
		panic(err)
	}

	return &VerkleTrie{
		db:      db,
		rootKey: rootKey,
		root:    root,
	}
}

func (m *VerkleTrie) dbKey(key []byte) []byte {
	return append([]byte(m.rootKey), key...)
}

func (m *VerkleTrie) nodeResolver(key []byte) ([]byte, error) {
	return m.db.Get(append(m.rootKey, key...))
}

func (m *VerkleTrie) Get(key []byte) ([]byte, error) {
	return m.root.Get(key, m.nodeResolver)
}

func (m *VerkleTrie) GetFileRootHash(filepathHash []byte) ([]byte, error) {
	key := GetTreeKeyForFileRootHash(filepathHash)
	return m.Get(key)
}

func (m *VerkleTrie) InsertValue(filepathHash []byte, data []byte) error {
	// insert the value size
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	if err := m.Insert(storageSizeKey, uint256.NewInt(uint64(len(data))).Bytes()); err != nil {
		return errors.Wrap(err, "insert storage size")
	}

	chunks := getStorageDataChunks(data)
	for i, chunk := range chunks {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, uint64(i))
		if err := m.Insert(chunkKey, chunk); err != nil {
			return errors.Wrap(err, "insert storage chunk")
		}
	}
	return nil
}

func getStorageDataChunks(data []byte) [][]byte {
	size := len(data)
	chunkSize := int(ChunkSize.Uint64())
	chunks := make([][]byte, 0, size/chunkSize+1)

	chunkNum := size / chunkSize
	for i := 0; i < chunkNum; i++ {
		chunks = append(chunks, data[i*chunkSize:(i+1)*chunkSize])
	}
	if size%chunkSize > 0 {
		chunks = append(chunks, data[chunkNum*chunkSize:])
	}
	return chunks
}

func (m *VerkleTrie) GetValue(filepathHash []byte) ([]byte, error) {
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	sizeBytes, err := m.Get(storageSizeKey)
	if err != nil {
		return nil, err
	}

	size := new(uint256.Int).SetBytes(sizeBytes)
	mod := new(uint256.Int)
	chunkNum := new(uint256.Int)
	chunkNum, mod = size.DivMod(size, ChunkSize, mod)

	valueBytes := make([]byte, 0, size.Uint64())
	for i := uint64(0); i < chunkNum.Uint64(); i++ {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, i)
		chunk, err := m.Get(chunkKey)
		if err != nil {
			return nil, err
		}

		valueBytes = append(valueBytes, chunk...)
	}
	if mod.Uint64() > 0 {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, chunkNum.Uint64())
		chunk, err := m.Get(chunkKey)
		if err != nil {
			return nil, err
		}
		valueBytes = append(valueBytes, chunk[:mod.Uint64()]...)
	}
	return valueBytes, nil
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

func (m *VerkleTrie) Commit() Hash {
	return m.root.Commit().Bytes()
}

func (m *VerkleTrie) CommitAndFlush() (Hash, error) {
	root, ok := m.root.(*verkle.InternalNode)
	if !ok {
		return Hash{}, errors.New("unexpected root node type")
	}

	nodes, err := root.BatchSerialize()
	if err != nil {
		return Hash{}, fmt.Errorf("serializing tree nodes: %s", err)
	}

	batch := m.db.NewBatch()
	rootKeyLen := len(m.rootKey)
	path := make([]byte, 0, rootKeyLen+32)
	path = append(path, []byte(m.rootKey)...)
	for _, node := range nodes {
		path := append(path[:rootKeyLen], node.Path...)
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

	if err := batch.Write(); err != nil {
		return Hash{}, fmt.Errorf("batch write error: %s", err)
	}

	return m.Hash(), nil
}

// func (m *VerkleTrie) Flush() error {
// 	v, err := m.root.Serialize()
// 	if err != nil {
// 		return Hash{}, fmt.Errorf("serializing root node: %s", err)
// 	}
// 	if err := batch.Put(DBKey(m.key), v); err != nil {
// 		return Hash{}, fmt.Errorf("put root node to disk: %s", err)
// 	}
// }

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

func HexToBytes(s string) []byte {
	v, _ := hex.DecodeString(s)
	return v
}
