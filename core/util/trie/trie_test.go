package trie

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/0chain/common/core/encryption"
	"github.com/stretchr/testify/assert"
)

func TestEmptyTrie(t *testing.T) {
	trie := New()
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))

	val, found := trie.Get([]byte("00000"))

	assert.True(t, found)
	assert.Equal(t, []byte("hello"), val)

}

func TestRichTrie(t *testing.T) {
	trie := New()
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("hello"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00000"), 6, []byte("hi"))

	val, found := trie.Get([]byte("00000"))

	assert.True(t, found)
	assert.Equal(t, []byte("hi"), val)

}

func TestDeleteTrie(t *testing.T) {
	trie := New()
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("hello"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00000"), 6, []byte("hi"))

	n := trie.Delete([]byte("00000"))
	assert.NotNil(t, n)

	_, found := trie.Get([]byte("00000"))

	assert.False(t, found)
}

func TestMerklelize(t *testing.T) {
	trie := New()
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("hello"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00000"), 6, []byte("hi"))

	weight_00000 := []byte(strconv.FormatUint(6, 10))
	weight_00200 := []byte(strconv.FormatUint(9, 10))
	weight_00300 := []byte(strconv.FormatUint(8, 10))
	weight_00220 := []byte(strconv.FormatUint(7, 10))

	hash_00000 := encryption.RawHash(append(weight_00000, "hi"...))
	hash_00200 := encryption.RawHash(append(weight_00200, "hi"...))
	hash_00300 := encryption.RawHash(append(weight_00300, "hello"...))
	hash_00220 := encryption.RawHash(append(weight_00220, "hello"...))

	weight_00200_00220 := []byte(strconv.FormatUint(16, 10))
	hash00200_00220 := encryption.RawHash(append(weight_00200_00220, append(hash_00200, hash_00220...)...))
	weight_root := []byte(strconv.FormatUint(30, 10))
	h := encryption.RawHash(append(weight_root, append(append(hash_00000, hash00200_00220...), hash_00300...)...))

	fmt.Println(hex.EncodeToString(h))
	fmt.Println(hex.EncodeToString(hash_00000))
	fmt.Println(hex.EncodeToString(hash00200_00220))
	fmt.Println(hex.EncodeToString(hash_00200))
	fmt.Println(hex.EncodeToString(hash_00220))
	fmt.Println(hex.EncodeToString(hash_00300))

	assert.Equal(t, h, trie.root.Hash())
}

//	    00
//	  /  |  \
//	[0   2   3]
//	/    \    \
//
// 00  [0 2]  00
//
//	 /  \
//	0    0
func TestFixedLengthHexKeyMerkleTrie_Values(t *testing.T) {
	trie := New()

	trie.InsertOrUpdate([]byte("00000"), 10, []byte("1"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("2"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("3"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("4"))

	values := trie.Values()

	for _, v := range values {
		fmt.Println(string(v))
	}
	weights := trie.Weights()

	for _, w := range weights {
		fmt.Println(w)
	}
}

//	    00
//	  /  |  \
//	[0   2   3]
//	/    \    \
//
// 00  [0 2]  00
//
//	 /  \
//	0    0
func TestFixedLengthHexKeyMerkleTrie_Values2(t *testing.T) {
	trie := New()

	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("hello"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00000"), 6, []byte("hi"))

	values := trie.Values()

	for _, v := range values {
		fmt.Println(string(v))
	}
	weights := trie.Weights()

	for _, w := range weights {
		fmt.Println(w)
	}
	hashes := trie.Hashes()

	for _, h := range hashes {
		fmt.Println(hex.EncodeToString(h))
	}
}

//	    00
//	  /  |  \
//	[0   2   3]
//	/    \    \
//
// 00  [0 2]  00
//
//	 /  \
//	0    0
func TestFixedLengthHexKeyMerkleTrie_Delete(t *testing.T) {
	trie := New()

	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("hello"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00000"), 6, []byte("hi"))
	trie.Delete([]byte("00220"))
	trie.Delete([]byte("00200"))

	values := trie.Values()

	for _, v := range values {
		fmt.Println(string(v))
	}
	weights := trie.Weights()

	for _, w := range weights {
		fmt.Println(w)
	}
	hashes := trie.Hashes()

	for _, h := range hashes {
		fmt.Println(hex.EncodeToString(h))
	}
}

func TestFixedLengthHexKeyMerkleTrie_FloorNodeValue(t *testing.T) {
	trie := New()

	trie.InsertOrUpdate([]byte("00000"), 10, []byte("10"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("9"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("8"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("7"))

	value, err := trie.FloorNodeValue(22)
	assert.NoError(t, err)
	assert.Equal(t, "7", string(value))

	value, err = trie.FloorNodeValue(35)
	assert.Error(t, err)

	value, err = trie.FloorNodeValue(30)
	assert.NoError(t, err)
	assert.Equal(t, "8", string(value))

}
