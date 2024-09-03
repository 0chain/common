package trie

import (
	"crypto/sha256"
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

func TestInsertOrderSensitive(t *testing.T) {
	trie := New()
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("01220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("02220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("12000"), 6, []byte("hi"))

	newTrie := New()
	trie.InsertOrUpdate([]byte("12000"), 6, []byte("hi"))
	trie.InsertOrUpdate([]byte("02220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("01220"), 7, []byte("hello"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("hi"))
	trie.InsertOrUpdate([]byte("00000"), 10, []byte("hello"))
	assert.Equal(t, trie.root.Hash(), newTrie.root.Hash())
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

func TestFixedLengthHexKeyMerkleTrie_Serialize(t *testing.T) {
	trie := New()

	trie.InsertOrUpdate([]byte("00000"), 10, []byte("10"))
	trie.InsertOrUpdate([]byte("00200"), 9, []byte("9"))
	trie.InsertOrUpdate([]byte("00300"), 8, []byte("8"))
	trie.InsertOrUpdate([]byte("00220"), 7, []byte("7"))
	data, err := trie.Serialize()
	assert.NoError(t, err)
	assert.Greater(t, len(data), 0)
	newTrie := New()
	err = newTrie.Deserialize(data)
	assert.NoError(t, err)
	fmt.Println("root1: ", trie.root, "root2: ", newTrie.root)
	assert.Equal(t, hex.EncodeToString(trie.root.Hash()), hex.EncodeToString(newTrie.root.Hash()))
}

func BenchmarkSerialize(b *testing.B) {
	// b.Run("SerializeDeserialize 100K", func(b *testing.B) {
	trie := New()
	keys := make([][]byte, 100000)
	sha256 := sha256.New()
	for i := 0; i < 100000; i++ {
		sha256.Write([]byte(strconv.Itoa(i)))
		key := sha256.Sum(nil)
		sha256.Reset()
		keys[i] = []byte(hex.EncodeToString(key))
		trie.InsertOrUpdate(keys[i], uint64(i), []byte(strconv.Itoa(i)))
	}
	b.ResetTimer()
	// now := time.Now()
	for i := 0; i < b.N; i++ {
		data, err := trie.Serialize()
		assert.NoError(b, err)
		// b.Log("serialize time: ", time.Since(now).Milliseconds())
		// now = time.Now()
		b.Log("data len: ", len(data))
		newTrie := New()
		err = newTrie.Deserialize(data)
		assert.NoError(b, err)
		// b.Log("deserialize time: ", time.Since(now).Milliseconds())
		assert.Equal(b, hex.EncodeToString(trie.root.Hash()), hex.EncodeToString(newTrie.root.Hash()))
	}
	// })

	// b.Run("SerializeDeserialize 500K", func(b *testing.B) {
	// 	trie := New()
	// 	keys := make([][]byte, 500000)
	// 	sha256 := sha256.New()
	// 	for i := 0; i < 500000; i++ {
	// 		sha256.Write([]byte(strconv.Itoa(i)))
	// 		key := sha256.Sum(nil)
	// 		sha256.Reset()
	// 		keys[i] = []byte(hex.EncodeToString(key))
	// 		trie.InsertOrUpdate(keys[i], uint64(i), []byte(strconv.Itoa(i)))
	// 	}
	// 	b.ResetTimer()
	// 	// now := time.Now()
	// 	data, err := trie.Serialize()
	// 	assert.NoError(b, err)
	// 	// b.Log("serialize time: ", time.Since(now).Milliseconds())
	// 	// now = time.Now()
	// 	b.Log("data len: ", len(data))
	// 	newTrie := New()
	// 	err = newTrie.Deserialize(data)
	// 	assert.NoError(b, err)
	// 	// b.Log("deserialize time: ", time.Since(now).Milliseconds())
	// 	assert.Equal(b, hex.EncodeToString(trie.root.Hash()), hex.EncodeToString(newTrie.root.Hash()))
	// })

	// b.Run("SerializeDeserialize 1M", func(b *testing.B) {
	// 	trie := New()
	// 	keys := make([][]byte, 1000000)
	// 	sha256 := sha256.New()
	// 	for i := 0; i < 1000000; i++ {
	// 		sha256.Write([]byte(strconv.Itoa(i)))
	// 		key := sha256.Sum(nil)
	// 		sha256.Reset()
	// 		keys[i] = []byte(hex.EncodeToString(key))
	// 		trie.InsertOrUpdate(keys[i], uint64(i), []byte(strconv.Itoa(i)))
	// 	}
	// 	b.ResetTimer()
	// 	// now := time.Now()
	// 	data, err := trie.Serialize()
	// 	assert.NoError(b, err)
	// 	// b.Log("serialize time: ", time.Since(now).Milliseconds())
	// 	// now = time.Now()
	// 	b.Log("data len: ", len(data))
	// 	newTrie := New()
	// 	err = newTrie.Deserialize(data)
	// 	assert.NoError(b, err)
	// 	// b.Log("deserialize time: ", time.Since(now).Milliseconds())
	// 	assert.Equal(b, hex.EncodeToString(trie.root.Hash()), hex.EncodeToString(newTrie.root.Hash()))
	// })
}

// goos: linux
// goarch: amd64
// pkg: github.com/0chain/common/core/util/trie
// cpu: 13th Gen Intel(R) Core(TM) i5-13400F
// BenchmarkSerialize/SerializeDeserialize_100K-16                 1000000000               0.4087 ns/op
// --- BENCH: BenchmarkSerialize/SerializeDeserialize_100K-16
//     trie_test.go:247: data len:  15129091
// BenchmarkSerialize/SerializeDeserialize_500K-16                        1        1861843864 ns/op
// --- BENCH: BenchmarkSerialize/SerializeDeserialize_500K-16
//     trie_test.go:272: data len:  73226557
// BenchmarkSerialize/SerializeDeserialize_1M-16                          1        3921399288 ns/op
// --- BENCH: BenchmarkSerialize/SerializeDeserialize_1M-16
//     trie_test.go:297: data len:  149861461
