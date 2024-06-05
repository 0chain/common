package verkletrie

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethereum/go-verkle"
	"github.com/stretchr/testify/assert"
)

var keys = [][]byte{
	HexToBytes("a3b45d6e7f890123456789abcdef0123456789abcdef0123456789abcdef0123"),
	HexToBytes("123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0"),
	HexToBytes("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"),
	HexToBytes("f10a5b26c19e3d94b6a87fe0c41269abd4e29a935fb7e4cd9a51f8b1272d3a68"),
	HexToBytes("43e1b7c9f20d5e76b4a1c8f62e9a03d57b9c16e3fa12b7a589d20c4f7e5a8c3b"),
	HexToBytes("e6d2a1f94c10b3e87f5d2b9c4a3ea62f709b52d68e13c4d7a8b61f9eb3c2475a"),
	HexToBytes("a58c3e7b12f4d96a3b8e27f0a1c76e5d49f2b6a13d0c5e28f7b19d4a2c07b831"),
	HexToBytes("f094e7a3b8d2c5f1093b4e76a8d2b5c4e1f7a3965d2c0e4b8a5d19f2b3c7e8b4"),
	HexToBytes("1a9f4c2d3b6e7a5f8d29c3b14e07a6d5c8b3f2d7a1e4b9c5d702b8f1c3a9d4e6"),
	HexToBytes("7e8c3b2f6a1d5e0c8b29f4a713d6e5c8f7a2b1d0e9c5a4b8f36e7d12c8b0a4f9"),
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
	value, err := vt.GetWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), value)

	value, err = vt.GetWithHashedKey(keys[1])
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
	_, err = vt.DeleteWithHashedKey(keys[0])
	assert.Nil(t, err)

	// Check that the data is no longer there
	value, err := vt.GetWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = vt.GetWithHashedKey(keys[1])
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
	vt.Flush()

	// fmt.Println(string(verkle.ToDot(vt.root)))

	// Create a new tree with the db
	newVt := New("alloc_1", db)
	// Check if the data can be acquired
	value, err := newVt.GetWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, keys[0], value)

	value, err = newVt.GetWithHashedKey(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, keys[1], value)
}

func TestTreeKeyStorage(t *testing.T) {
	// Createverkletrie
	vt := New("alloc_1", NewInMemoryVerkleDB())

	filepathHash := keys[0]
	rootHash := keys[1]
	// insert file: alloc1/testfile.txt
	key := GetTreeKeyForFileHash(filepathHash)
	err := vt.Insert(key, rootHash)
	assert.Nil(t, err)

	vt.Flush()

	v, err := vt.GetWithHashedKey(key)
	assert.Nil(t, err)

	assert.Equal(t, rootHash, v)

	bigValue := append(keys[0], keys[1]...)
	err = vt.InsertValue(filepathHash, bigValue)
	assert.Nil(t, err)

	vt.Flush()

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

	vt.Flush()

	v, err := vt.GetValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, values, v)
}

func TestInsertsNodeChanges(t *testing.T) {
	vt := New("alloc_1", NewInMemoryVerkleDB())
	for i := 0; i < len(keys[:7]); i++ {
		err := vt.InsertValue(keys[i], keys[i])
		assert.Nil(t, err)
	}

	vt.Flush()

	vt.Insert(keys[7], keys[7])
	vt.Flush()
}

func TestProof(t *testing.T) {
	vt := New("alloc_1", NewInMemoryVerkleDB())
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.Insert(keys[i], keys[i])
		assert.Nil(t, err)
	}

	root := vt.Commit()
	dproof, stateDiff, err := MakeProof(vt, keys[:3])
	assert.Nil(t, err)

	dproofBytes, err := json.Marshal(dproof)
	assert.Nil(t, err)

	stateDiffBytes, err := json.Marshal(stateDiff)
	assert.Nil(t, err)

	// deserialize dproof
	dproof2 := &verkle.VerkleProof{}
	err = json.Unmarshal(dproofBytes, dproof2)
	assert.Nil(t, err)

	// deserialize stateDiff
	stateDiff2 := verkle.StateDiff{}
	err = json.Unmarshal(stateDiffBytes, &stateDiff2)
	assert.Nil(t, err)

	err = VerifyProofPresence(dproof2, stateDiff2, root[:], keys[:3])
	assert.Nil(t, err)
}

func TestProofNotExistKey(t *testing.T) {
	vt := New("alloc_1", NewInMemoryVerkleDB())
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.Insert(keys[i], keys[i])
		assert.Nil(t, err)
	}

	root := vt.Commit()

	t.Run("proof no key exists", func(t *testing.T) {
		dp, sdiff, err := MakeProof(vt, keys[3:])
		assert.Nil(t, err)

		err = VerifyProofAbsence(dp, sdiff, root[:], keys[3:])
		fmt.Println("err:", err)
		assert.Nil(t, err)
	})

	t.Run("proof absence of exist key - should fail", func(t *testing.T) {
		dp, sdiff, err := MakeProof(vt, keys[2:])
		assert.Nil(t, err)

		err = VerifyProofAbsence(dp, sdiff, root[:], keys[2:])
		assert.EqualError(t, err, "verkle proof contains value")
	})
}

func TestFileRootHash(t *testing.T) {
	vt := New("alloc_1", NewInMemoryVerkleDB())
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.InsertFileRootHash(keys[i], keys[i])
		assert.Nil(t, err)
	}

	vt.Commit()
	_, err := vt.DeleteFileRootHash(keys[2])
	assert.Nil(t, err)
	vt.Commit()

	// Verify that the root hash of the file is deleted
	// v, err := vt.(GetTreeKeyForFileHash(keys[2]))
	v2, err := vt.GetFileRootHash(keys[2])
	assert.Nil(t, err)
	assert.Nil(t, v2)

	v1, err := vt.GetFileRootHash(keys[1])
	assert.Nil(t, err)
	assert.NotNil(t, v1)
	fmt.Printf("%x\n", v1)
}
