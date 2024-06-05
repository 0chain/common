package verkletrie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var keys = [][]byte{
	HexToBytes("a3b45d6e7f890123456789abcdef0123456789abcdef0123456789abcdef0123"),
	HexToBytes("123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0"),
	HexToBytes("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"),
}

func TestVerkleTrie_Insert(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	// Insert some data
	err := vt.Insert(keys[0], []byte("value1"))
	assert.Nil(t, err)
	err = vt.Insert(keys[1], []byte("value2"))
	assert.Nil(t, err)

	// Check that the data is there
	value, err := vt.Get(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), value)

	value, err = vt.Get(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Delete(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	// Insert some data
	err := vt.Insert(keys[0], []byte("value1"))
	assert.Nil(t, err)
	err = vt.Insert(keys[1], []byte("value2"))
	assert.Nil(t, err)

	// Delete some data
	_, err = vt.Delete(keys[0])
	assert.Nil(t, err)

	// Check that the data is no longer there
	value, err := vt.Get(keys[0])
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = vt.Get(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Commit(t *testing.T) {
	// Createverkletrie
	db := NewInMemoryVerkleDB()
	vt := New("alloc_1", db)

	err := vt.Insert(keys[0], keys[0])
	assert.Nil(t, err)
	err = vt.Insert(keys[1], keys[1])
	assert.Nil(t, err)

	// Commit the tree
	vt.CommitAndFlush()

	// fmt.Println(string(verkle.ToDot(vt.root)))

	// Create a new tree with the db
	newVt := New("alloc_1", db)
	// Check if the data can be acquired
	value, err := newVt.Get(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, keys[0], value)

	value, err = newVt.Get(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, keys[1], value)
}

func TestTreeKeyStorage(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	filepathHash := keys[0]
	rootHash := keys[1]
	// insert file: alloc1/testfile.txt
	key := GetTreeKeyForFileRootHash(filepathHash)
	err := vt.Insert(key, rootHash)
	assert.Nil(t, err)

	vt.Commit()

	v, err := vt.Get(key)
	assert.Nil(t, err)

	assert.Equal(t, rootHash, v)

	bigValue := append(keys[0], keys[1]...)
	err = vt.InsertValue(filepathHash, bigValue)
	assert.Nil(t, err)

	vt.Commit()

	vv, err := vt.GetValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, bigValue, vv)
}

func TestTreeStorageLargeData(t *testing.T) {
	vt := New("alloc_1", NewInMemoryVerkleDB())
	filepathHash := keys[0]

	mainStoreChunkNum := 1000
	totalChunkNum := headerStorageCap + mainStoreChunkNum

	values := make([]byte, 0, totalChunkNum*int(ChunkSize.Uint64()))
	// test to use out all header spaces for storage
	for i := 0; i < totalChunkNum; i++ {
		values = append(values, keys[0]...)
	}

	err := vt.InsertValue(filepathHash, values)
	assert.Nil(t, err)

	vt.Commit()

	v, err := vt.GetValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, values, v)
}
