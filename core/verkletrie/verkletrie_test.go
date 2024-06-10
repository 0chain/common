package verkletrie

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/0chain/common/core/verkletrie/database"
	"github.com/0chain/common/core/verkletrie/database/rocksdb"
	"github.com/ethereum/go-verkle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

var benchKeys = [][]byte{
	HexToBytes("de34efbdcb37c44397e5213f295ce973edfd6e3ebd586ff4be20dd308d875635"),
	HexToBytes("9d79cd3523ad43ca5849a3c93c9aeb2a2a1e09e7b129a24583eb79ee914ca685"),
	HexToBytes("8e0b76196671d514b7b26e5478e459c34339a965ad168424184b33e11c7503c3"),
	HexToBytes("a2e17cc6bedd7c8079a25ef6f7ee2f1da3f773ae8d993e7836b2d22cd1697dfc"),
	HexToBytes("dc52d23c1940f342f228aa307cb436bde167a68fca3d005a9cbfca8648c23a63"),
	HexToBytes("a262530eb4e165305ce7e70f5480a9f6f74d18b67467c3a6ce5259a2e5f79c21"),
	HexToBytes("28d75ae78e0ee004ec8ac7691129916d5aac021b80b247712b4816cb5b2245ed"),
	HexToBytes("887658a80146b0cd5fd87c3c5e01699432025425cc70be24a134331484db8bd2"),
	HexToBytes("62046f78350e648a3dab7747222b9e4822f44800432ef83b929c62748bcd3212"),
	HexToBytes("357c0778187bce682a78331d7e8496838c241345635327b53335ea1dfa69e938"),
}

var (
	mainStorageLargeValue = []byte{}
	once                  sync.Once
	// dbType can be "inmemory" or "rocksdb"
	// var dbType = "rocksdb"
	dbType = "rocksdb"
	// dbType = "inmemory"
)

var generate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&generate, "gen", false, "generate test data")
	flag.Parse()

	mainStorageLargeValue = make([]byte, 0, 128)
	for i := 0; i < 128; i++ {
		mainStorageLargeValue = append(mainStorageLargeValue, keys[0][:]...)
	}

	if generate && dbType == "rocksdb" {
		// generate the ./testdata/bench.db if it's the first time to run the benchmark
		makeBenchRocksDB1MSmall()
		makeBenchRocksDBLargeValue()
		makeBenchRocksDB1KNodes()
		makeBenchRocksDB1KLargeNodes()
	}

	os.Exit(m.Run())
}

func testPrepareDB(t testing.TB) (database.DB, func()) {
	switch {
	case dbType == "inmemory":
		return database.NewInMemoryVerkleDB(), func() {}
	case dbType == "rocksdb":
		return makeRocksDB(t)
	}
	return nil, nil
}

func makeBenchRocksDB1MSmall() {
	dbPath := "./testdata/bench.db"
	fmt.Println("dbPath:", dbPath)
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	vt := New("alloc_1", db)
	for i := 0; i < 1000000; i++ {
		var key []byte
		if i < len(benchKeys) {
			key = benchKeys[i]
		} else {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			key = randBytes[:]
		}
		// fmt.Printf("%x\n", key)
		err := vt.insertValue(key[:], key[:])
		if err != nil {
			panic(err)
		}
	}
	vt.Flush()
}

func makeBenchRocksDBLargeValue() {
	dbPath := fmt.Sprintf("./testdata/bench_large.db")
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	vt := New("alloc_1", db)
	for i := 0; i < 10000; i++ {
		key := []byte{}
		if i < len(benchKeys) {
			key = benchKeys[i]
		} else {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			key = randBytes[:]
		}
		err := vt.insertValue(key[:], mainStorageLargeValue)
		if err != nil {
			panic(err)
		}
	}
	vt.Flush()
}
func makeBenchRocksDB1KNodes() {
	dbPath := fmt.Sprintf("./testdata/bench_1k.db")
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	vt := New("alloc_1", db)
	for i := 0; i < 1000; i++ {
		key := []byte{}
		if i < len(benchKeys) {
			key = benchKeys[i]
		} else {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			key = randBytes[:]
		}
		err := vt.insertValue(key[:], mainStorageLargeValue)
		if err != nil {
			panic(err)
		}
	}
	vt.Flush()
}

func makeBenchRocksDB1KLargeNodes() {
	dbPath := fmt.Sprintf("./testdata/bench_1k_large.db")
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	vt := New("alloc_1", db)
	for i := 0; i < 1000; i++ {
		key := []byte{}
		if i < len(benchKeys) {
			key = benchKeys[i]
		} else {
			randBytes := make([]byte, 32)
			rand.Read(randBytes)
			key = randBytes[:]
		}
		err := vt.insertValue(key[:], mainStorageLargeValue)
		if err != nil {
			panic(err)
		}
	}
	vt.Flush()
}

func getBenchRocksDB1KLargeDB() database.DB {
	dbPath := "./testdata/bench_1k_large.db"
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func getBenchRocksDB1MSmall() database.DB {
	dbPath := "./testdata/bench.db"
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}

	return db
}

func getBenchRocksDBLarge() database.DB {
	dbPath := "./testdata/bench_large.db"
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func getBenchRocksDB1K() database.DB {
	dbPath := "./testdata/bench_1k.db"
	db, err := rocksdb.NewRocksDB(dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func makeRocksDB(t testing.TB) (db database.DB, clean func()) {
	dbPath := fmt.Sprintf("./testdata/%s_%d.db", t.Name(), time.Now().Nanosecond())
	dbDir := filepath.Dir(dbPath)
	os.MkdirAll(dbDir, os.ModePerm)

	var err error
	db, err = rocksdb.NewRocksDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		if err := os.RemoveAll(dbPath); err != nil {
			t.Fatal(err)
		}
	}
}

func TestVerkleTrie_Insert(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)

	// Insert some data
	err := vt.insert(keys[0], []byte("value1"))
	assert.Nil(t, err)
	err = vt.insert(keys[1], []byte("value2"))
	assert.Nil(t, err)

	// Check that the data is there
	value, err := vt.getWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), value)

	value, err = vt.getWithHashedKey(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Delete(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)

	// Insert some data
	err := vt.insert(keys[0], []byte("value1"))
	assert.Nil(t, err)
	err = vt.insert(keys[1], []byte("value2"))
	assert.Nil(t, err)

	// Delete some data
	_, err = vt.deleteWithHashedKey(keys[0])
	assert.Nil(t, err)

	// Check that the data is no longer there
	value, err := vt.getWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = vt.getWithHashedKey(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, []byte("value2"), value)
}

func TestVerkleTrie_Flush(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)

	err := vt.insert(keys[0], keys[0])
	assert.Nil(t, err)

	err = vt.insert(keys[1], keys[1])
	assert.Nil(t, err)

	// Commit the tree
	vt.Flush()
	fmt.Println("flush count:", flushCount)
	oldFlushCount := flushCount

	// fmt.Println(string(verkle.ToDot(vt.root)))

	// Create a new tree with the db
	newVt := New("alloc_1", db)
	// Check if the data can be acquired
	value, err := newVt.getWithHashedKey(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, keys[0], value)

	value, err = newVt.getWithHashedKey(keys[1])
	assert.Nil(t, err)
	assert.Equal(t, keys[1], value)

	err = newVt.insert(keys[2], keys[2])
	assert.Nil(t, err)
	newVt.Flush()
	fmt.Println("new flush count:", flushCount-oldFlushCount)
}

func TestTreeKeyStorage(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)

	filepathHash := keys[0]
	rootHash := keys[1]
	// insert file: alloc1/testfile.txt
	key := GetTreeKeyForFileHash(filepathHash)
	err := vt.insert(key, rootHash)
	assert.Nil(t, err)

	vt.Flush()

	v, err := vt.getWithHashedKey(key)
	assert.Nil(t, err)

	assert.Equal(t, rootHash, v)

	bigValue := append(keys[0], keys[1]...)
	err = vt.insertValue(filepathHash, bigValue)
	assert.Nil(t, err)

	vt.Flush()

	vv, err := vt.getValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, bigValue, vv)
}

func TestTreeStorageLargeData(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	filepathHash := keys[0]

	mainStoreChunkNum := 1000
	totalChunkNum := headerStorageCap + mainStoreChunkNum

	values := make([]byte, 0, totalChunkNum*int(ChunkSize.Uint64()))
	// test to use out all header spaces for storage
	for i := 0; i < totalChunkNum; i++ {
		values = append(values, keys[0][:]...)
	}

	err := vt.insertValue(filepathHash, values)
	assert.Nil(t, err)

	vv, err := vt.getValue(filepathHash)
	assert.Nil(t, err)
	require.Equal(t, values, vv)

	vt.Flush()

	vt = New("alloc_1", db)
	fmt.Println("-----------------------------------")

	v, err := vt.getValue(filepathHash)
	assert.Nil(t, err)

	assert.Equal(t, values, v)
}

func TestInsertsNodeChanges(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	for i := 0; i < len(keys[:7]); i++ {
		err := vt.insertValue(keys[i], keys[i])
		assert.Nil(t, err)
	}

	vt.Flush()
	oldC := flushCount
	fmt.Println("flush count:", flushCount)

	vt.insert(keys[7], keys[7])
	vt.Flush()
	fmt.Println("new flush count:", flushCount-oldC)
}

func TestProof(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.insert(keys[i], keys[i])
		assert.Nil(t, err)
	}

	root := vt.Commit()
	dproof, stateDiff, err := makeProof(vt, keys[:3])
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

	err = verifyProofPresence(dproof2, stateDiff2, root[:], keys[:3])
	assert.Nil(t, err)
}

func TestProofNotExistKey(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.insert(keys[i], keys[i])
		assert.Nil(t, err)
	}

	root := vt.Commit()

	t.Run("proof no key exists", func(t *testing.T) {
		dp, sdiff, err := makeProof(vt, keys[3:])
		assert.Nil(t, err)

		err = VerifyProofAbsence(dp, sdiff, root[:], keys[3:])
		assert.Nil(t, err)
	})

	t.Run("proof absence of exist key - should fail", func(t *testing.T) {
		dp, sdiff, err := makeProof(vt, keys[2:])
		assert.Nil(t, err)

		err = VerifyProofAbsence(dp, sdiff, root[:], keys[2:])
		assert.EqualError(t, err, "verkle proof contains value")
	})
}

func TestDeleteFileRootHash(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.insertFileRootHash(keys[i], keys[i])
		assert.Nil(t, err)
	}

	vt.Commit()
	_, err := vt.deleteFileRootHash(keys[2])
	assert.Nil(t, err)
	vt.Commit()

	// Verify that the root hash of the file is deleted
	v2, err := vt.getFileRootHash(keys[2])
	assert.Nil(t, err)
	assert.Nil(t, v2)

	v1, err := vt.getFileRootHash(keys[1])
	assert.Nil(t, err)
	assert.NotNil(t, v1)
}

func TestDeleteValue(t *testing.T) {
	// t.Parallel()
	db, clean := testPrepareDB(t)
	defer clean()
	vt := New("alloc_1", db)
	for i := 0; i < len(keys[:3]); i++ {
		err := vt.insertValue(keys[i], mainStorageLargeValue[:])
		assert.Nil(t, err)
	}

	vt.Commit()

	vb, err := vt.getValue(keys[0])
	assert.Nil(t, err)
	assert.Equal(t, mainStorageLargeValue[:], vb)

	err = vt.deleteValue(keys[0])
	assert.Nil(t, err)

	// verify that the value is deleted
	vv, err := vt.getValue(keys[0])
	assert.Nil(t, err)
	assert.Nil(t, vv)

	// assert that all related nodes are deleted
	storageSizeKey := GetTreeKeyForStorageSize(keys[0])
	sv, err := vt.getWithHashedKey(storageSizeKey)
	assert.Nil(t, err)
	assert.Nil(t, sv)

	// assert that all chunks are deleted
	size := len(mainStorageLargeValue)
	chunkNum := size / int(ChunkSize.Uint64())
	if size%int(ChunkSize.Uint64()) > 0 {
		chunkNum++
	}
	for i := 0; i < chunkNum; i++ {
		chunkKey := GetTreeKeyForStorageSlot(keys[0], uint64(i))
		cv, err := vt.getWithHashedKey(chunkKey)
		assert.Nil(t, err)
		assert.Nil(t, cv)
	}
}

func BenchmarkInsertSmallValue(b *testing.B) {
	db := getBenchRocksDB1MSmall()
	// db := getBenchRocksDBLarge()
	defer db.Close()

	vt := New("alloc_1", db)
	for i := 0; i < b.N; i++ {
		randBytes := make([]byte, 32)
		rand.Read(randBytes)
		key := randBytes[:]

		err := vt.insertValue(key, key[:])
		assert.Nil(b, err)
		vt.Flush()
	}
}

func BenchmarkInsertLargeValue(b *testing.B) {
	db := getBenchRocksDBLarge()
	defer db.Close()

	vt := New("alloc_1", db)
	for i := 0; i < b.N; i++ {
		randBytes := make([]byte, 32)
		rand.Read(randBytes)
		key := randBytes[:]

		err := vt.insertValue(key, mainStorageLargeValue)
		assert.Nil(b, err)
		vt.Flush()
	}
}

// func BenchmarkMakeProof2(b *testing.B) {
// 	db := getBenchRocksDB()
// 	defer db.Close()

// 	vt := New("alloc_1", db)
// 	for i := 0; i < b.N; i++ {
// 		// randBytes := make([]byte, 32)
// 		// rand.Read(randBytes)
// 		key := smallValueKeys[i%len(smallValueKeys)]
// 		// err := vt.InsertValue(key, key[:])
// 		// assert.Nil(b, err)
// 		// vt.Flush()
// 		MakeProof(vt, Keylist{key})
// 	}
// }

// var largeValue1KNodeKeys = [][]byte{
// 	HexToBytes("cdb8b26ac25ececa8bd26aa67a6dbb11af4a75b303fb17bf644e08ba4276ea07"),
// 	HexToBytes("dc5a39e1e3494e87e06a9647b3e0c33657226410a06a2d61c530c68e2cbbe1ab"),
// 	HexToBytes("199ffb5ce1205c8a60e1b905c7730154d2812113224b9321a9362da61745ee73"),
// 	HexToBytes("ceb9d23653fd5163ac8409ce8878e0c82e46ee4d281f0e23b96e9f89ed5091d5"),
// 	HexToBytes("bfb6896cbbb07b273b6616a5ba9c4928e17e4bf812b7804a1c93baa0ca4f48bf"),
// 	HexToBytes("3e16ee78e18c62c7c9fb52535e4fe10958c519b549039a62621cf9f32cea203d"),
// 	HexToBytes("aeac233ad49984e3c96d53caf0e0f4b760a15bab01c21a233e4f904578a096f5"),
// 	HexToBytes("88eb70d652ef57e7f873e4b03f38a2144be905ae90ff54111b2ac62e7b93829a"),
// 	HexToBytes("c653552478a06c7075e21f64f166fc4a687499843bdc9e521fa106a5bb012d15"),
// 	HexToBytes("37d3367def3a1b8d75237c7099d44ff38a6b1a70250ac4fde7dc6f0378c08856"),
// }

func BenchmarkMakeProof(b *testing.B) {
	b.Run("1k small nodes", func(b *testing.B) {
		db := getBenchRocksDB1K()
		// db := getBenchRocksDB1KLargeDB()
		// db := getBenchRocksDBLarge()
		defer db.Close()
		// db, clean := testPrepareDB(b)
		// defer clean()
		vt := New("alloc_1", db)
		for i := 0; i < b.N; i++ {
			key := benchKeys[i%len(benchKeys)]
			makeProof(vt, Keylist{key})
		}
	})

	b.Run("1k large nodes", func(b *testing.B) {
		db := getBenchRocksDB1KLargeDB()
		defer db.Close()
		vt := New("alloc_1", db)
		for i := 0; i < b.N; i++ {
			key := benchKeys[i%len(benchKeys)]
			makeProof(vt, Keylist{key})
		}
	})

	b.Run("10k large nodes", func(b *testing.B) {
		db := getBenchRocksDBLarge()
		defer db.Close()
		vt := New("alloc_1", db)
		for i := 0; i < b.N; i++ {
			key := benchKeys[i%len(benchKeys)]
			makeProof(vt, Keylist{key})
		}
	})

	b.Run("1M small nodes", func(b *testing.B) {
		db := getBenchRocksDB1MSmall()
		defer db.Close()
		vt := New("alloc_1", db)
		for i := 0; i < b.N; i++ {
			key := benchKeys[i%len(benchKeys)]
			makeProof(vt, Keylist{key})
		}

	})

	b.Run("2 small nodes", func(b *testing.B) {
		db, clean := testPrepareDB(b)
		defer clean()
		vt := New("alloc_1", db)
		// Insert some data
		err := vt.insert(keys[0], []byte("value1"))
		assert.Nil(b, err)
		err = vt.insert(keys[1], []byte("value2"))
		assert.Nil(b, err)
		vt.Commit()

		for i := 0; i < b.N; i++ {
			key := keys[i%2]
			makeProof(vt, Keylist{key})
		}
	})
}

func BenchmarkVerifyProof(b *testing.B) {
	db, clean := testPrepareDB(b)
	defer clean()

	vt := New("alloc_1", db)
	for i := 0; i < len(keys); i++ {
		err := vt.insertFileRootHash(keys[i], keys[i])
		assert.Nil(b, err)
	}

	vt.Flush()
	root := vt.Hash()

	key := GetTreeKeyForFileHash(keys[0])
	vp, sd, err := makeProof(vt, Keylist{key})
	assert.Nil(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err = verifyProofPresence(vp, sd, root[:], Keylist{key})
		assert.Nil(b, err)
	}
}

func TestMakeProofFileMeta(t *testing.T) {
	db, clean := testPrepareDB(t)
	defer clean()

	vt := New("alloc_1", db)
	for i := 0; i < len(keys)-1; i++ {
		err := vt.InsertFileMeta(keys[i], keys[i], keys[i])
		assert.Nil(t, err)
	}

	vt.Flush()
	root := vt.Hash()
	vp, sd, err := MakeProofFileMeta(vt, keys)
	assert.Nil(t, err)

	t.Run("proof key exists", func(t *testing.T) {
		err = VerifyProofPresenceFileMeta(vp, sd, root[:], Keylist{keys[0]})
		assert.Nil(t, err)
	})

	t.Run("proof key does not exist", func(t *testing.T) {
		noneExistKey := keys[len(keys)-1]
		err = VerifyProofPresenceFileMeta(vp, sd, root[:], Keylist{noneExistKey})
		assert.NotNil(t, err)
	})
}
