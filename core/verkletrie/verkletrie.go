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

func (m *VerkleTrie) getWithHashedKey(key []byte) ([]byte, error) {
	return m.root.Get(key, m.nodeResolver)
}

func (m *VerkleTrie) getFileRootHash(filepathHash []byte) ([]byte, error) {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.getWithHashedKey(key)
}

func (m *VerkleTrie) deleteFileRootHash(filepathHash []byte) (bool, error) {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.deleteWithHashedKey(key)
}

func (m *VerkleTrie) insertFileRootHash(filepathHash []byte, rootHash []byte) error {
	key := GetTreeKeyForFileHash(filepathHash)
	return m.insert(key, rootHash)
}

func (m *VerkleTrie) insertValue(filepathHash []byte, data []byte) error {
	// insert the value size
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	vb := uint256.NewInt(uint64(len(data))).Bytes32()
	if err := m.insert(storageSizeKey, vb[:]); err != nil {
		return errors.Wrap(err, "insert storage size")
	}

	chunks := getStorageDataChunks(data)
	for i, chunk := range chunks {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, uint64(i))
		if err := m.insert(chunkKey, chunk); err != nil {
			return errors.Wrap(err, "insert storage chunk")
		}
	}
	return nil
}

func (m *VerkleTrie) deleteValue(filepathHash []byte) error {
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	sizeBytes, err := m.getWithHashedKey(storageSizeKey)
	if err != nil {
		return errors.Wrap(err, "delete value error on getting storage size")
	}

	size := new(uint256.Int).SetBytes(sizeBytes)
	// remove all chunks nodes
	chunkNum := getChunkNum(*size)
	for i := 0; i < int(chunkNum); i++ {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, uint64(i))
		_, err = m.deleteWithHashedKey(chunkKey)
		if err != nil {
			return errors.Wrap(err, "delete value error on deleting storage chunk")
		}
	}

	// delete the storage size node
	if _, err := m.deleteWithHashedKey(storageSizeKey); err != nil {
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

func (m *VerkleTrie) getValue(filepathHash []byte) ([]byte, error) {
	storageSizeKey := GetTreeKeyForStorageSize(filepathHash)
	sizeBytes, err := m.getWithHashedKey(storageSizeKey)
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
		chunk, err := m.getWithHashedKey(chunkKey)
		if err != nil {
			return nil, err
		}

		valueBytes = append(valueBytes, chunk...)
	}
	if mod.Uint64() > 0 {
		chunkKey := GetTreeKeyForStorageSlot(filepathHash, chunkNum.Uint64())
		chunk, err := m.getWithHashedKey(chunkKey)
		if err != nil {
			return nil, err
		}
		valueBytes = append(valueBytes, chunk[:mod.Uint64()]...)
	}
	return valueBytes, nil
}

func (m *VerkleTrie) insert(key []byte, value []byte) error {
	return m.root.Insert(key, value, m.nodeResolver)
}

func (m *VerkleTrie) deleteWithHashedKey(key []byte) (bool, error) {
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

func (m *VerkleTrie) InsertFileMeta(filepathHash []byte, rootHash, metaData []byte) error {
	if err := m.insertFileRootHash(filepathHash, rootHash); err != nil {
		return err
	}

	// insert metaData
	return m.insertValue(filepathHash, metaData)
}

func (m *VerkleTrie) DeleteFileMeta(filepathHash []byte) error {
	_, err := m.deleteWithHashedKey(GetTreeKeyForFileHash(filepathHash))
	if err != nil {
		return err
	}

	return m.deleteValue(filepathHash)
}

func (m *VerkleTrie) GetFileMetaRootHash(filepathHash []byte) ([]byte, error) {
	return m.getFileRootHash(filepathHash)
}

func (m *VerkleTrie) GetFileMeta(filepathHash []byte) ([]byte, error) {
	return m.getValue(filepathHash)
}

type Keylist [][]byte

func makeProof(trie *VerkleTrie, keys Keylist) (*verkle.VerkleProof, verkle.StateDiff, error) {
	proof, _, _, _, err := verkle.MakeVerkleMultiProof(trie.root, nil, keys, trie.nodeResolver)
	if err != nil {
		return nil, nil, err
	}

	return verkle.SerializeProof(proof)
}

func MakeProofFileMeta(trie *VerkleTrie, files [][]byte) (*verkle.VerkleProof, verkle.StateDiff, error) {
	keys := make([][]byte, 0, len(files))
	for _, file := range files {
		keys = append(keys, GetTreeKeyForFileHash(file))
	}
	proof, _, _, _, err := verkle.MakeVerkleMultiProof(trie.root, nil, keys, trie.nodeResolver)
	if err != nil {
		return nil, nil, err
	}

	return verkle.SerializeProof(proof)
}

func VerifyProofPresenceFileMeta(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte, files Keylist) error {
	keys := make([][]byte, 0, len(files))
	for _, file := range files {
		keys = append(keys, GetTreeKeyForFileHash(file))
	}
	return verifyProofPresence(vp, stateDiff, stateRoot, keys)
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

// verifyProofPresence verifies that the verkle proof is valid and keys are presence in the state tree
func verifyProofPresence(vp *verkle.VerkleProof, stateDiff verkle.StateDiff, stateRoot []byte, keys Keylist) error {
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
