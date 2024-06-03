package verkletrie

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var keys = []string{
	"a3b45d6e7f890123456789abcdef0123456789abcdef0123456789abcdef0123",
	"123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0",
	"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
}

func TestVerkleTrie_Insert(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	// Insert some data
	err := vt.Insert(hexToBytes(keys[0]), []byte("value1"))
	assert.Nil(t, err)
	err = vt.Insert(hexToBytes(keys[1]), []byte("value2"))
	assert.Nil(t, err)

	// Check that the data is there
	value, err := vt.Get(hexToBytes(keys[0]))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), value)

	value, err = vt.Get(hexToBytes(keys[1]))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Delete(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	// Insert some data
	err := vt.Insert(hexToBytes(keys[0]), []byte("value1"))
	assert.Nil(t, err)
	err = vt.Insert(hexToBytes(keys[1]), []byte("value2"))
	assert.Nil(t, err)

	// Delete some data
	_, err = vt.Delete(hexToBytes(keys[0]))
	assert.Nil(t, err)

	// Check that the data is no longer there
	value, err := vt.Get(hexToBytes(keys[0]))
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = vt.Get(hexToBytes(keys[1]))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Commit(t *testing.T) {
	// Createverkletrie
	db := NewInMemoryVerkleDB()
	vt := New("alloc_1", db)

	// Insert some data
	err := vt.Insert(hexToBytes(keys[0]), []byte("value1"))
	assert.Nil(t, err)
	err = vt.Insert(hexToBytes(keys[1]), []byte("value2"))
	assert.Nil(t, err)

	// Commit the tree
	_, err = vt.Commit(true)
	assert.Nil(t, err)

	// Create a new tree with the db

	newVt := New("alloc_1", db)

	// Check if the data can be acquired
	value, err := newVt.Get(hexToBytes(keys[0]))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), value)

	value, err = newVt.Get(hexToBytes(keys[1]))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func hexToBytes(s string) []byte {
	v, _ := hex.DecodeString(s)
	return v
}
