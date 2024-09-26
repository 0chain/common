package wmpt

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/0chain/common/core/util/storage/kv"
	"github.com/stretchr/testify/assert"
)

func TestSerializeHashNode(t *testing.T) {
	hash := sha256.Sum256([]byte("hello"))
	node := hashNode{hash: hash[:], weight: 100000000000}
	data, err := node.Serialize()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(data))
}

func TestInsertOrderSensitive(t *testing.T) {
	// pebDir := "/pebble/storage"
	// os.RemoveAll(pebDir)
	// os.MkdirAll(pebDir, 0777)
	// defer os.RemoveAll(pebDir)
	// db, err := storage.NewPebbleAdapter(pebDir)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	trie := New(nil, nil)
	trie.Update(keys[0], []byte("hello"), 10)
	trie.Update(keys[1], []byte("hi"), 9)
	trie.Update(keys[2], []byte("hello"), 7)
	trie.Update(keys[3], []byte("hello"), 7)
	trie.Update(keys[4], []byte("hi"), 6)

	newTrie := New(nil, nil)
	newTrie.Update(keys[4], []byte("hi"), 6)
	newTrie.Update(keys[3], []byte("hello"), 7)
	newTrie.Update(keys[2], []byte("hello"), 7)
	newTrie.Update(keys[1], []byte("hi"), 9)
	newTrie.Update(keys[0], []byte("hello"), 10)
	assert.Equal(t, trie.root.Weight(), newTrie.root.Weight())
	assert.Equal(t, trie.root.CalcHash(), newTrie.root.CalcHash())
}

func TestTrieUpdate(t *testing.T) {
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	trie := New(nil, nil)
	trie.Update(keys[0], []byte("hello"), 10)
	trie.Update(keys[0], []byte("hi"), 9)
	trie.root.CalcHash()
	assert.Equal(t, trie.root.Weight(), uint64(9))
}

func TestEmptyTrie(t *testing.T) {
	trie := New(nil, nil)
	assert.Equal(t, trie.root.CalcHash(), emptyState)
	assert.Equal(t, trie.root.Weight(), uint64(0))
}

func TestTrieDelete(t *testing.T) {
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	trie := New(nil, nil)
	trie.Update(keys[0], []byte("a"), 10)
	trie.Update(keys[1], []byte("b"), 9)
	trie.Update(keys[2], []byte("c"), 7)
	trie.Update(keys[3], []byte("d"), 7)
	trie.root.CalcHash()
	assert.Equal(t, trie.root.Weight(), uint64(33))
	h1 := trie.root.CalcHash()
	trie.Update(keys[3], nil, 0)
	assert.Equal(t, trie.root.Weight(), uint64(26))
	trie.Update(keys[3], []byte("d"), 7)
	h2 := trie.root.CalcHash()
	assert.Equal(t, trie.root.Weight(), uint64(33))
	assert.Equal(t, h1, h2)
}

func TestTrieCommit(t *testing.T) {
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	trie := New(nil, nil)
	trie.Update(keys[0], []byte("a"), 10)
	trie.Update(keys[1], []byte("b"), 9)
	trie.Update(keys[2], []byte("c"), 7)
	trie.Update(keys[3], []byte("d"), 7)
	trie.root.CalcHash()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	pebDir := filepath.Join(wd, "pebble_storage")
	assert.NoError(t, os.RemoveAll(pebDir))
	assert.NoError(t, os.MkdirAll(pebDir, 0777))
	db, err := kv.NewPebbleAdapter(pebDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(pebDir)
	}()
	dbTrie := New(nil, db)
	dbTrie.Update(keys[0], []byte("a"), 10)
	dbTrie.Update(keys[1], []byte("b"), 9)
	dbTrie.Update(keys[2], []byte("c"), 7)
	dbTrie.Update(keys[3], []byte("d"), 7)
	batcher, err := dbTrie.Commit(0)
	assert.NoError(t, err)
	err = batcher.Commit(true)
	assert.NoError(t, err)
	assert.Equal(t, trie.root.Weight(), dbTrie.root.Weight())
	assert.Equal(t, trie.root.Hash(), dbTrie.root.Hash())
	dbTrie.DeleteNodes()
	dbTrie.Update(keys[4], []byte("e"), 6)
	trie.Update(keys[4], []byte("e"), 6)
	batcher, err = dbTrie.Commit(0)
	assert.NoError(t, err)
	err = batcher.Commit(true)
	assert.NoError(t, err)
	dbTrie.DeleteNodes()
	assert.Equal(t, trie.root.Weight(), dbTrie.root.Weight())
	assert.Equal(t, trie.root.CalcHash(), dbTrie.root.Hash())
}

func TestRollbackTrie(t *testing.T) {
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	wd, err := os.Getwd()
	assert.NoError(t, err)
	pebDir := filepath.Join(wd, "pebble_storage")
	assert.NoError(t, os.RemoveAll(pebDir))
	assert.NoError(t, os.MkdirAll(pebDir, 0777))
	db, err := kv.NewPebbleAdapter(pebDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(pebDir)
	}()
	dbTrie := New(nil, db)
	dbTrie.Update(keys[0], []byte("a"), 10)
	dbTrie.Update(keys[1], []byte("b"), 9)
	dbTrie.Update(keys[2], []byte("c"), 7)
	batcher, err := dbTrie.Commit(0)
	assert.NoError(t, err)
	err = batcher.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, dbTrie.DeleteNodes())
	rootNode := &hashNode{
		weight: dbTrie.root.Weight(),
		hash:   dbTrie.root.Hash(),
	}
	_, _, err = dbTrie.GetBlockProof(21)
	assert.NoError(t, err)
	dbTrie.Update(keys[3], []byte("d"), 7)
	batcher, err = dbTrie.Commit(0)
	assert.NoError(t, err)
	err = batcher.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, dbTrie.DeleteNodes())
	newHash := dbTrie.root.Hash()
	dbTrie.RollbackTrie(rootNode)
	assert.Equal(t, dbTrie.root.Weight(), uint64(26))
	assert.Equal(t, dbTrie.root.CalcHash(), rootNode.hash)
	_, err = db.Get(newHash)
	assert.Error(t, err)
	_, err = db.Get(rootNode.hash)
	assert.NoError(t, err)
	_, _, err = dbTrie.GetBlockProof(21)
	assert.NoError(t, err)
}

func TestUpdateTrie(t *testing.T) {
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	trie := New(nil, nil)
	trie.Update(keys[0], []byte("a"), 10)
	trie.Update(keys[1], []byte("b"), 10)
	h1 := trie.root.CalcHash()
	assert.Equal(t, trie.root.Weight(), uint64(20))
	trie.Update(keys[1], []byte("c"), 5)
	h2 := trie.root.CalcHash()
	assert.Equal(t, trie.root.Weight(), uint64(15))
	assert.NotEqual(t, h1, h2)
}
