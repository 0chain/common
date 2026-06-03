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

// TestCommitCollapsePruneDangling guards against the dangling-node bug where a
// routing node collapsed at collapseLevel was Save()d but never registered on
// createdChan. A node whose hash was queued for deletion in an earlier commit
// (delete) and then re-appears at the collapse boundary in a later commit
// (re-add) would be pruned by the deferred DeleteNodes() while the new root
// still references it -> reloading and GetPath hits "pebble: not found".
// Delete/re-add churn shifts tree depth around COLLAPSE_DEPTH and triggers it.
func TestCommitCollapsePruneDangling(t *testing.T) {
	const (
		N        = 40
		collapse = 2 // mirrors a shallow COLLAPSE_DEPTH so boundary is hit often
	)
	keys := make([][]byte, N)
	for i := 0; i < N; i++ {
		h := sha256.Sum256([]byte("collapse-" + strconv.Itoa(i)))
		// Force a shared 2-byte prefix so all keys live under a deep common
		// path; the random sha256 tail then branches into a multi-level trie.
		h[0], h[1] = 0xAB, 0xCD
		keys[i] = h[:]
	}

	wd, err := os.Getwd()
	assert.NoError(t, err)
	pebDir := filepath.Join(wd, "pebble_storage_collapse")
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

	trie := New(nil, db)
	val := func(i int) []byte { return []byte("v" + strconv.Itoa(i)) }
	for i := 0; i < N; i++ {
		assert.NoError(t, trie.Update(keys[i], val(i), uint64(i+1)))
	}
	commit := func() {
		trie.SaveRoot()
		batcher, cerr := trie.Commit(collapse)
		assert.NoError(t, cerr)
		assert.NoError(t, batcher.Commit(true))
		assert.NoError(t, trie.DeleteNodes())
	}
	commit()

	// After each round all N keys are present; reload from the persisted root
	// and prove every path resolves (no pruned/dangling node).
	checkIntact := func(round int) {
		reloaded := New(&hashNode{hash: trie.Root(), weight: trie.GetRoot().Weight()}, db)
		_, perr := reloaded.GetPath(keys)
		assert.NoErrorf(t, perr, "round %d: GetPath hit a pruned collapsed node", round)
	}

	for round := 0; round < N; round++ {
		i := round
		assert.NoError(t, trie.Update(keys[i], nil, 0)) // delete -> hash queued for deferred delete
		commit()
		assert.NoError(t, trie.Update(keys[i], val(i), uint64(i+1))) // re-add -> may re-appear at boundary
		commit()
		checkIntact(round)
	}
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

func TestCommitAndRollback(t *testing.T) {
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
	t1 := New(nil, db)
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	t1.SaveRoot()
	t1.Update(keys[0], []byte("a"), 10)
	b, err := t1.Commit(3)
	assert.NoError(t, err)
	err = b.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, t1.DeleteNodes())
	t1.SaveRoot()
	assert.NoError(t, t1.Update(keys[0], nil, 0))
	b, err = t1.Commit(3)
	assert.NoError(t, err)
	err = b.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, t1.DeleteNodes())
	h := t1.Root()
	assert.Equal(t, h, emptyState)
	data, err := t1.GetPath(nil)
	assert.NoError(t, err)
	t2 := New(nil, nil)
	assert.NoError(t, t2.Deserialize(data))
	h2 := t2.GetRoot().CalcHash()
	assert.Equal(t, h2, h)
	t1.SaveRoot()
	t1.Rollback()
	assert.Equal(t, t1.Root(), h)
}

func TestUploadDelete(t *testing.T) {
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
	t1 := New(nil, db)
	keys := make([][]byte, 0, 5)
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
	}
	t1.SaveRoot()
	t1.Update(keys[0], []byte("a"), 10)
	b, err := t1.Commit(3)
	assert.NoError(t, err)
	err = b.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, t1.DeleteNodes())
	t1.SaveRoot()
	assert.NoError(t, t1.Update(keys[0], nil, 0))
	b, err = t1.Commit(3)
	assert.NoError(t, err)
	err = b.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, t1.DeleteNodes())
	t1.Update(keys[0], []byte("a"), 10)
	b, err = t1.Commit(3)
	assert.NoError(t, err)
	err = b.Commit(true)
	assert.NoError(t, err)
	assert.NoError(t, t1.DeleteNodes())
	t2 := New(&hashNode{
		weight: 10,
		hash:   t1.GetRoot().CalcHash(),
	}, db)
	_, _, err = t2.GetBlockProof(6)
	assert.NoError(t, err)
}
