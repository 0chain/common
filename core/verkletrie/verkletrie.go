package verkletrie

import (
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/0chain/common/core/verkletrie/database"
	"github.com/ethereum/go-verkle"
	"github.com/holiman/uint256"
)

var DBVerkleNodeKeyPrefix = []byte("verkle_node_")

var ChunkSize = uint256.NewInt(32) // 32

type (
	Hash      [32]byte
	Address32 [32]byte
)

type VerkleTrie struct {
	db      database.DB
	rootKey []byte // the key in db where the whole serialized verkle trie to be persisted. Could be the allocation id
	root    verkle.VerkleNode
}

func rootKey(key string) []byte {
	return append(DBVerkleNodeKeyPrefix, key...)
}

func New(key string, db database.DB) *VerkleTrie {
	var root verkle.VerkleNode
	rootKey := rootKey(key)

	payload, err := db.Get([]byte(rootKey))
	if err != nil {
		if err == database.ErrMissingNode {
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

func (m *VerkleTrie) GetWithHashedKey(key []byte) ([]byte, error) {
	return m.root.Get(key, m.nodeResolver)
}

func (m *VerkleTrie) GetFileRootHash(filepathHash []byte) ([]byte, error) {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.GetWithHashedKey(key)
}

func (m *VerkleTrie) DeleteFileRootHash(filepathHash []byte) (bool, error) {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.DeleteWithHashedKey(key)
}

func (m *VerkleTrie) InsertFileRootHash(filepathHash []byte, rootHash []byte) error {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.Insert(key, rootHash)
}

func (m *VerkleTrie) InsertValue(filepathHash []byte, data []byte) error {
	// insert the value size
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	vb := uint256.NewInt(uint64(len(data))).Bytes32()
	if err := m.Insert(storageSizeKey, vb[:]); err != nil {
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

func (m *VerkleTrie) DeleteValue(filepathHash []byte) error {
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	sizeBytes, err := m.GetWithHashedKey(storageSizeKey)
	if err != nil {
		return errors.Wrap(err, "delete value error on getting storage size")
	}

	size := new(uint256.Int).SetBytes(sizeBytes)
	// remove all chunks nodes
	chunkNum := getChunkNum(*size)
	for i := 0; i < int(chunkNum); i++ {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, uint64(i))
		_, err = m.DeleteWithHashedKey(chunkKey)
		if err != nil {
			return errors.Wrap(err, "delete value error on deleting storage chunk")
		}
	}

	// delete the storage size node
	if _, err := m.DeleteWithHashedKey(storageSizeKey); err != nil {
		return errors.Wrap(err, "delete storage size")
	}

	return nil
}

func getChunkNum(size uint256.Int) uint64 {
	mod := new(uint256.Int)
	size.DivMod(&size, ChunkSize, mod)
	if mod.CmpUint64(0) > 0 {
		return size.Uint64() + 1
	}
	return size.Uint64()
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
	sizeBytes, err := m.GetWithHashedKey(storageSizeKey)
	if err != nil {
		return nil, err
	}

	size := new(uint256.Int).SetBytes(sizeBytes)
	if size.Uint64() == 0 {
		return nil, nil
	}

	mod := new(uint256.Int)
	chunkNum := new(uint256.Int)
	chunkNum, mod = size.DivMod(size, ChunkSize, mod)

	valueBytes := make([]byte, 0, size.Uint64())
	for i := uint64(0); i < chunkNum.Uint64(); i++ {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, i)
		chunk, err := m.GetWithHashedKey(chunkKey)
		if err != nil {
			return nil, err
		}

		valueBytes = append(valueBytes, chunk...)
	}
	if mod.Uint64() > 0 {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, chunkNum.Uint64())
		chunk, err := m.GetWithHashedKey(chunkKey)
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

func (m *VerkleTrie) DeleteWithHashedKey(key []byte) (bool, error) {
	return m.root.Delete(key, m.nodeResolver)
}

func (m *VerkleTrie) Hash() Hash {
	return m.root.Commit().Bytes()
}

func (m *VerkleTrie) Commit() Hash {
	return m.root.Commit().Bytes()
}

var flushCount int32

func (m *VerkleTrie) flushFunc(p []byte, node verkle.VerkleNode) {
	nodeBytes, err := node.Serialize()
	if err != nil {
		panic(fmt.Errorf("serializing node: %s", err))
	}

	rootKeyLen := len(m.rootKey)
	path := make([]byte, 0, rootKeyLen+32)
	path = append(path, []byte(m.rootKey)...)
	path = append(path[:rootKeyLen], p...)
	if err := m.db.Set(path, nodeBytes[:]); err != nil {
		panic(fmt.Errorf("put node to disk: %s", err))
	}

	atomic.AddInt32(&flushCount, 1)
}

func (m *VerkleTrie) Flush() {
	m.root.Commit()
	m.root.(*verkle.InternalNode).Flush(m.flushFunc)
}

type Keylist [][]byte

func MakeProof(trie *VerkleTrie, keys Keylist) (*verkle.VerkleProof, verkle.StateDiff, error) {
	proof, _, _, _, err := verkle.MakeVerkleMultiProof(trie.root, nil, keys, trie.nodeResolver)
	if err != nil {
		return nil, nil, err
	}

	return verkle.SerializeProof(proof)
}

func verifyProof(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte) (*verkle.Proof, error) {
	dproof, err := verkle.DeserializeProof(vp, stateDiff)
	if err != nil {
		return nil, fmt.Errorf("verkle proof deserialization error: %w", err)
	}

	root := new(verkle.Point)
	if err := root.SetBytes(stateRoot); err != nil {
		return nil, fmt.Errorf("verkle root deserialization error: %w", err)
	}

	tree, err := verkle.PreStateTreeFromProof(dproof, root)
	if err != nil {
		return nil, fmt.Errorf("error rebuilding the pre-tree from proof: %w", err)
	}

	if err := verkle.VerifyVerkleProofWithPreState(dproof, tree); err != nil {
		return nil, fmt.Errorf("verkle proof verification error: %w", err)
	}

	return dproof, nil
}

// VerifyProofPresence verifies that the verkle proof is valid and keys are presence in the state tree
func VerifyProofPresence(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte, keys Keylist) error {
	if _, err := verifyProof(vp, stateDiff, stateRoot); err != nil {
		return err
	}

	// v, _ := json.MarshalIndent(stateDiff, "", "  ")
	// fmt.Println(string(v))

	sdMap := make(map[string][32]byte, len(stateDiff))
	for _, sd := range stateDiff {
		for _, su := range sd.SuffixDiffs {
			path := append(sd.Stem[:], su.Suffix)
			if su.CurrentValue != nil {
				sdMap[string(path)] = *su.CurrentValue
			}
		}
	}

	for _, key := range keys {
		if _, ok := sdMap[string(key)]; !ok {
			return fmt.Errorf("verkle proof could not find key: %x", key)
		}
	}

	return nil
}

func VerifyProofAbsence(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte, keys Keylist) error {
	dproof, err := verifyProof(vp, stateDiff, stateRoot)
	if err != nil {
		return err
	}

	for _, v := range dproof.PreValues {
		if len(v) != 0 {
			return errors.New("verkle proof contains value")
		}
	}
	return nil
}

func HexToBytes(s string) []byte {
	v, _ := hex.DecodeString(s)
	return v
}
