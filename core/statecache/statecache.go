package statecache

import (
	lru "github.com/hashicorp/golang-lru"
)

// NewBlockTxnCaches creates a new block cache and a transaction cache for the given block
func NewBlockTxnCaches(sc *StateCache, b Block) (*BlockCache, *TransactionCache) {
	bc := NewBlockCache(sc, b)
	tc := NewTransactionCache(bc)
	return bc, tc
}

// Value is an interface that all values in the state cache must implement
type Value interface {
	Clone() Value
	CopyFrom(v interface{}) bool
}

type String string

func (se String) Clone() Value {
	return se
}

func (set String) CopyFrom(v interface{}) bool {
	return false
}

// Cacheable checks if the given value is able to be cached
func Cacheable(v interface{}) (Value, bool) {
	cv, ok := v.(Value)
	return cv, ok
}

type Copyer interface {
	// CopyFrom copies the value from the given value, returns false if not able to copy
	CopyFrom(v interface{}) bool
}

// Copyable checks if the given value is able to be copied
func Copyable(v interface{}) (Copyer, bool) {
	cv, ok := v.(Copyer)
	return cv, ok
}

type valueNode struct {
	data          Value
	deleted       bool   // indicates the value was removed
	round         int64  // round number when this value is updated
	prevBlockHash string // previous block hash
}

type StateCache struct {
	maxHisDepth int
	cache       *lru.Cache
	hashCache   *lru.Cache
}

func NewStateCache() *StateCache {
	cache, err := lru.New(100 * 1024)
	if err != nil {
		panic(err)
	}

	maxHisDepth := 100
	hCache, err := lru.New(maxHisDepth)
	if err != nil {
		panic(err)
	}

	return &StateCache{
		maxHisDepth: maxHisDepth,
		cache:       cache,
		hashCache:   hCache,
	}
}

func (sc *StateCache) commitRound(round int64, prevHash, blockHash string) {
	sc.hashCache.Add(blockHash, prevHash)
}

// Get returns the value with the given key and block hash
func (sc *StateCache) Get(key, blockHash string) (Value, bool) {
	// sc.mu.RLock()
	// defer sc.mu.RUnlock()

	blockValues, ok := sc.cache.Get(key)
	if !ok {
		// logging.Logger.Debug("state cache get - not found", zap.String("key", key))
		return nil, false
	}

	bvs := blockValues.(*lru.Cache)
	vv, ok := bvs.Get(blockHash)
	if ok {
		v := vv.(valueNode)

		if !v.deleted {
			// logging.Logger.Debug("state cache get", zap.String("key", key))
			return v.data.Clone(), true
		}

		return nil, false
	}

	var count int
	for {
		count++
		// get previous block hash
		prevHash, ok := sc.hashCache.Get(blockHash)
		if !ok {
			// could not find previous hash
			return nil, false
		}

		blockHash = prevHash.(string)
		vv, ok = bvs.Get(blockHash)
		if !ok {
			// stop if the value is not found in previous maxHisDepth rounds
			if count >= sc.maxHisDepth {
				return nil, false
			}

			continue
		}

		v := vv.(valueNode)
		if v.deleted {
			return nil, false
		}

		return v.data.Clone(), true
	}
}

// Remove removes the values map with the given key
func (sc *StateCache) Remove(key string) {
	// sc.mu.Lock()
	// defer sc.mu.Unlock()

	sc.cache.Remove(key)
	// delete(sc.cache, key)
}
