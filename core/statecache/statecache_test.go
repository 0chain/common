package statecache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/0chain/common/core/logging"
	"github.com/stretchr/testify/require"
)

func init() {
	logging.InitLogging("development", "")
}

func TestStateCacheSetGet(t *testing.T) {
	sc := NewStateCache()

	// Test Get method when cache is empty
	_, ok := sc.Get("key1", "hash1")
	if ok {
		t.Error("Expected false, got ", ok)
	}

	bc := NewBlockCache(sc, Block{Hash: "hash1"})
	bc.Set("key1", String("data1"))
	bc.Commit()

	// Test Get method when cache has a value
	v, ok := sc.Get("key1", "hash1")
	if !ok {
		t.Error("Expected true, got ", ok)
	}
	if v.(String) != "data1" {
		t.Error("Expected data1, got ", v)
	}
}

func TestBlockCache(t *testing.T) {
	sc := NewStateCache()
	ct := NewBlockCache(sc, Block{Hash: "hash1"})

	// Test Get method when cache is empty
	_, ok := ct.Get("key1")
	if ok {
		t.Error("Expected false, got ", ok)
	}

	// Test Set method
	ct.Set("key1", String("value1"))
	value, ok := ct.Get("key1")
	require.True(t, ok)
	require.EqualValues(t, "value1", value)

	// Test Commit method
	ct.Set("key2", String("value2"))

	ct.Commit()
	value, ok = sc.Get("key2", "hash1")
	require.True(t, ok)
	require.EqualValues(t, "value2", value)

	// Add a new value to the cache for key1 in hash2 block
	value2 := String("data2")
	ct2 := NewBlockCache(sc, Block{PrevHash: "hash1", Hash: "hash2"})
	ct2.Set("key1", value2)
	ct2.Commit()

	// Get cache in current block
	v2, ok := sc.Get("key1", "hash2")
	require.True(t, ok)
	require.EqualValues(t, "data2", v2)

	// get data that is updated in hash1, no changes in hash2
	vv2, ok := sc.Get("key2", "hash2")
	require.True(t, ok)
	require.EqualValues(t, "value2", vv2)

	// Get cache in prior block
	v1, ok := sc.Get("key1", "hash1")
	require.True(t, ok)
	require.EqualValues(t, "value1", v1)
}

func TestCacheTx_NotCommitted(t *testing.T) {
	sc := NewStateCache()
	ct := NewBlockCache(sc, Block{Round: 1, Hash: "hash1"})

	// Test Get method when cache is empty
	_, ok := ct.Get("key1")
	require.False(t, ok)

	// Test Set method
	ct.Set("key1", String("value1"))
	value, ok := ct.Get("key1")
	require.True(t, ok)
	require.EqualValues(t, "value1", value)

	// Test Get method in state cache before committing
	_, ok = sc.Get("key1", "hash1")
	require.False(t, ok)

	ct.Commit()
	_, ok = sc.Get("key1", "hash1")
	require.True(t, ok)

	ct = NewBlockCache(sc, Block{Round: 2, Hash: "hash2", PrevHash: "hash1"})
	_, ok = ct.Get("key1")
	require.True(t, ok)

	// Test Remove method
	ct.remove("key1")
	_, ok = ct.Get("key1")
	require.False(t, ok)
	if ok {
		t.Error("Expected false, got ", ok)
	}

	ct.Commit()

	_, ok = sc.Get("key1", "hash2")
	require.False(t, ok)

	// should be exist in hash1
	v, ok := sc.Get("key1", "hash1")
	require.True(t, ok)
	require.EqualValues(t, "value1", v)
}

func TestCacheTx_SkipBlock(t *testing.T) {
	sc := NewStateCache()
	ct := NewBlockCache(sc, Block{Hash: "hash1"})

	// Add values to the cache in block "hash1"
	ct.Set("key1", String("value1"))

	// Commit the changes to the main cache
	ct.Commit()

	// Skip one block
	ct = NewBlockCache(sc, Block{PrevHash: "hash2", Hash: "hash3"})

	_, ok := ct.Get("key1")
	require.False(t, ok)

	// Add a new value to the cache in block "hash3"
	ct.Set("key1", String("value3"))

	_, ok = ct.Get("key1")
	require.True(t, ok)

	ct.Commit()
	v, ok := sc.Get("key1", "hash3")
	require.True(t, ok)
	require.EqualValues(t, "value3", v)
}

// func TestCacheTx_Shift(t *testing.T) {
// 	sc := NewStateCache()
// 	ct := NewBlockCache(sc, Block{Hash: "hash1"})

// 	// Add values to the cache in block "hash1"
// 	ct.Set("key1", String("value1_h1"))
// 	ct.Set("key2", String("value2_h1"))

// 	// Commit the changes to the main cache
// 	ct.Commit()

// 	// New block that update key1 only
// 	ct = NewBlockCache(sc, Block{PrevHash: "hash1", Hash: "hash2"})
// 	ct.Set("key1", String("value1_h2"))
// 	ct.Commit()

// 	// Commit should trigger shift of key2 from hash1 to hash2
// 	v, ok := sc.Get("key2", "hash2")
// 	require.True(t, ok)
// 	require.EqualValues(t, "value2_h1", v)

// 	// New block to update both key1 and key2
// 	ct = NewBlockCache(sc, Block{PrevHash: "hash2", Hash: "hash3"})
// 	ct.Set("key1", String("value1_h3"))
// 	ct.Set("key2", String("value2_h3"))

// 	v1, ok := ct.Get("key1")
// 	require.True(t, ok)
// 	require.EqualValues(t, "value1_h3", v1)

// 	v2, ok := ct.Get("key2")
// 	require.True(t, ok)
// 	require.EqualValues(t, "value2_h3", v2)

// 	ct.Commit()

// 	v1, ok = sc.Get("key1", "hash3")
// 	require.True(t, ok)
// 	require.EqualValues(t, "value1_h3", v1)

// 	v2, ok = sc.Get("key2", "hash3")
// 	require.True(t, ok)
// 	require.EqualValues(t, "value2_h3", v2)
// }

func TestConcurrentExecutionAndCommit(t *testing.T) {
	sc := NewStateCache()

	// Create two concurrent CacheTx instances for the same block
	ct1 := NewBlockCache(sc, Block{Hash: "hash1"})
	ct2 := NewBlockCache(sc, Block{Hash: "hash1"})

	// Set values in both CacheTx instances
	ct1.Set("key1", String("value1_h1"))
	ct2.Set("key1", String("value1_h1"))

	// Concurrently commit both CacheTx instances
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ct1.Commit()
	}()

	go func() {
		defer wg.Done()
		ct2.Commit()
	}()

	wg.Wait()

	// Verify that the cache is the same after concurrent execution and commit
	v1, ok := sc.Get("key1", "hash1")
	require.True(t, ok)
	require.EqualValues(t, "value1_h1", v1)
}

func TestAddRemoveAdd(t *testing.T) {
	sc := NewStateCache()

	// Create a CacheTx instance
	ct := NewBlockCache(sc, Block{Hash: "hash1"})

	// Add a value to the CacheTx
	ct.Set("key1", String(string([]byte("value1"))))

	// Commit the CacheTx
	ct.Commit()

	// Verify that the value is added to the StateCache
	v1, ok := sc.Get("key1", "hash1")
	require.True(t, ok)
	require.EqualValues(t, "value1", v1)

	// Create another CacheTx instance
	ct2 := NewBlockCache(sc, Block{PrevHash: "hash1", Hash: "hash2"})

	// Remove the value from the CacheTx
	ct2.remove("key1")

	// Commit the CacheTx
	ct2.Commit()

	// Verify that the value is removed from the StateCache
	_, ok = sc.Get("key1", "hash2")
	require.False(t, ok)

	ct3 := NewBlockCache(sc, Block{PrevHash: "hash2", Hash: "hash3"})
	// Add the value again to the CacheTx
	ct3.Set("key1", String("value1"))

	// ct2.Set("key1", String("value1"))

	// Commit the CacheTx
	ct3.Commit()

	// Verify that the value is added back to the StateCache
	v2, ok := sc.Get("key1", "hash3")
	require.True(t, ok)
	require.EqualValues(t, "value1", v2)
}

func TestTransactionCache(t *testing.T) {
	sc := NewStateCache()

	for i := 0; i < 10; i++ {
		hash := fmt.Sprintf("hash%d", i)
		var preHash string
		if i > 0 {
			preHash = fmt.Sprintf("hash%d", i-1)
		}

		bc := NewBlockCache(sc, Block{PrevHash: preHash, Hash: hash})
		tc := NewTransactionCache(bc)

		// Test Get method when cache is empty
		_, ok := tc.Get("key1")
		if ok && i == 0 {
			t.Error("Expected false, got ", ok)
		}

		// Test Set method
		value1 := fmt.Sprintf("value1_%s", hash)
		tc.Set("key1", String(fmt.Sprintf("value1_%s", hash)))
		value, ok := tc.Get("key1")
		require.True(t, ok)
		require.EqualValues(t, value1, value)

		// Test Commit method
		value2 := fmt.Sprintf("value2_%s", hash)
		tc.Set("key2", String(value2))
		tc.Commit()

		v1, ok := bc.Get("key1")
		require.True(t, ok)
		require.EqualValues(t, value1, v1)

		v2, ok := bc.Get("key2")
		require.True(t, ok)
		require.EqualValues(t, value2, v2)

		// sc should not have the values yet before commit
		_, ok = sc.Get("key1", hash)
		require.False(t, ok)

		_, ok = sc.Get("key2", hash)
		require.False(t, ok)

		// sc should see the values after commit
		bc.Commit()
		vv1, ok := sc.Get("key1", hash)
		require.True(t, ok)
		require.EqualValues(t, value1, vv1)

		vv2, ok := sc.Get("key2", hash)
		require.True(t, ok)
		require.EqualValues(t, value2, vv2)
	}

	for i := 0; i < 10; i++ {
		hash := fmt.Sprintf("hash%d", i)
		value1 := fmt.Sprintf("value1_%s", hash)
		v, ok := sc.Get("key1", hash)
		require.True(t, ok)
		require.EqualValues(t, value1, v)
	}
}

func TestTransactionCacheRemove(t *testing.T) {
	sc := NewStateCache()
	bc := NewBlockCache(sc, Block{Hash: "hash1"})

	tc := NewTransactionCache(bc)
	tc.Set("key1", String("value1"))

	v, ok := tc.Get("key1")
	require.True(t, ok)
	require.EqualValues(t, "value1", v)

	// remove
	tc.Remove("key1")

	_, ok = tc.Get("key1")
	require.False(t, ok)

	tc.Commit()

	_, ok = bc.Get("key1")
	require.False(t, ok)
}

type Foo struct {
	V string
}

func (f *Foo) Clone() Value {
	return &Foo{V: f.V}
}

func (f *Foo) CopyFrom(v interface{}) bool {
	if v, ok := v.(*Foo); ok {
		f.V = v.V
		return true
	}
	return false
}

func (f *Foo) Add() {

}

type Bar struct {
}

func (b *Bar) Add() {
}

type MsgInterface interface {
	Add()
}

func TestEnableCache(t *testing.T) {
	f := &Foo{V: "foo"}
	fi := MsgInterface(f)

	_, ok := Cacheable(fi)
	require.True(t, ok)

	b := &Bar{}
	_, ok = Cacheable(b)
	require.False(t, ok)
}

func TestEmptyBlockCache(t *testing.T) {
	// Create a new state cache
	sc := NewStateCache()

	// Add a value to the state cache for a specific block
	blockHash := "block123"
	key := "key123"
	value := String("value123")

	bc := NewBlockCache(sc, Block{Hash: blockHash})
	bc.Set(key, value)
	bc.Commit()

	// Create an empty block cache linked to the state cache
	bc2 := NewBlockCache(sc, Block{Hash: blockHash})
	_, ok := bc2.Get(key)
	require.False(t, ok)
}

func TestBlockCacheGetFromPrevious(t *testing.T) {
	sc := NewStateCache()
	ct := NewBlockCache(sc, Block{Hash: "hash1"})

	// Test Get method when cache is empty
	_, ok := ct.Get("key1")
	if ok {
		t.Error("Expected false, got ", ok)
	}

	ct.Set("key0", String("value0"))
	ct.Commit()

	var lastBC *BlockCache
	for i := 0; i <= 100; i++ {
		bc := NewBlockCache(sc, Block{PrevHash: fmt.Sprintf("hash%d", i+1), Hash: fmt.Sprintf("hash%d", i+2)})
		if i == 50 {
			bc.Set("key1", String("value50"))
		}

		bc.Commit()
		lastBC = bc
	}

	v, ok := lastBC.Get("key1")
	require.True(t, ok)
	require.EqualValues(t, "value50", v)

	// should not find the value in 100 rounds before
	_, ok = lastBC.Get("key0")
	require.False(t, ok)
}

func TestBlockCacheGetGap(t *testing.T) {
	sc := NewStateCache()
	ct := NewBlockCache(sc, Block{Hash: "hash1"})
	ct.Set("key1", String("value1"))
	ct.Commit()

	var lastBC *BlockCache
	for i := 0; i < 100; i++ {
		bc := NewBlockCache(sc, Block{PrevHash: fmt.Sprintf("hash%d", i+1), Hash: fmt.Sprintf("hash%d", i+2)})
		if i == 50 {
			// cause break
			continue
		}

		if i == 80 {
			bc.Set("key2", String("value80"))
		}

		bc.Commit()
		lastBC = bc
	}

	// update before 50 rounds would not be found
	_, ok := lastBC.Get("key1")
	require.False(t, ok)

	v, ok := lastBC.Get("key2")
	require.True(t, ok)
	require.EqualValues(t, "value80", v)
}
