package verkletrie

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-verkle"
	"github.com/holiman/uint256"
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

	err := vt.Insert(keys[0], append(keys[0], keys[1]...))
	assert.Nil(t, err)
	err = vt.Insert(keys[1], keys[1])
	assert.Nil(t, err)

	// Commit the tree
	_, err = vt.Commit(true)
	assert.Nil(t, err)

	fmt.Println(string(verkle.ToDot(vt.root)))

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

func TestTreeKey(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	filepathHash := keys[0]
	rootHash := keys[1]
	// insert file: alloc1/testfile.txt
	key := GetTreeKeyForFileRootHash(filepathHash)
	err := vt.Insert(key, rootHash)
	assert.Nil(t, err)

	vt.Commit(true)

	v, err := vt.Get(key)
	assert.Nil(t, err)

	assert.Equal(t, rootHash, v)

	bigValue := append(keys[0], keys[1]...)
	size := len(bigValue)
	sk := GetTreeKeyForStorageSize(filepathHash)
	sizeV := uint256.NewInt(uint64(size))
	vt.Insert(sk, sizeV.Bytes())
	for i := 0; i < 2; i++ {
		ssk := GetTreeKeyForStorageSlot(filepathHash, uint64(i))
		fmt.Println(ssk)
		vt.Insert(ssk, bigValue[i*32:(i+1)*32])
	}

	vt.Commit(true)

	vv, err := vt.GetValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, bigValue, vv)
}
