package wmpt

import (
	"crypto/sha256"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBlockProof(t *testing.T) {
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
	key, proof, err := trie.GetBlockProof(10)
	assert.NoError(t, err)
	assert.Equal(t, key, keys[4])
	newTrie := New(nil, nil)
	h, v, err := newTrie.VerifyBlockProof(10, proof)
	assert.NoError(t, err)
	assert.Equal(t, h, trie.root.Hash())
	assert.Equal(t, v, []byte("hi"))
}
