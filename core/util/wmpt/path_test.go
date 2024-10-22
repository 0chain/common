package wmpt

import (
	"crypto/sha256"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPath(t *testing.T) {
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
	trie.Update(keys[4], []byte("hi"), 10)
	trie.root.CalcHash()
	getKeys := [][]byte{
		keys[2],
		keys[3],
	}
	path, err := trie.GetPath(getKeys)
	assert.NoError(t, err)
	newTrie := New(nil, nil)
	err = newTrie.Deserialize(path)
	assert.NoError(t, err)
	assert.Equal(t, newTrie.root.Weight(), trie.root.Weight())
	err = newTrie.Update(keys[3], []byte("hi"), 12)
	assert.NoError(t, err)
	err = trie.Update(keys[3], []byte("hi"), 12)
	assert.NoError(t, err)
	assert.Equal(t, newTrie.root.Weight(), trie.root.Weight())
	assert.Equal(t, newTrie.root.CalcHash(), trie.root.CalcHash())
	err = newTrie.Update(keys[2], nil, 0)
	assert.NoError(t, err)
	err = trie.Update(keys[2], nil, 0)
	assert.NoError(t, err)
	assert.Equal(t, newTrie.root.Weight(), trie.root.Weight())
	assert.Equal(t, newTrie.root.CalcHash(), trie.root.CalcHash())
}

func TestPathAndDelete(t *testing.T) {
	trie := New(nil, nil)
	keys := make([][]byte, 0, 100)
	for i := 0; i < 100; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(i)))
		keys = append(keys, hash[:])
		trie.Update(hash[:], []byte{byte(i)}, uint64(i))
	}
	trie.root.CalcHash()
	path, err := trie.GetPath([][]byte{keys[50]})
	assert.NoError(t, err)
	newTrie := New(nil, nil)
	err = newTrie.Deserialize(path)
	assert.NoError(t, err)
	assert.Equal(t, newTrie.root.Weight(), trie.root.Weight())
	err = newTrie.Update(keys[50], nil, 0)
	assert.NoError(t, err)
}
