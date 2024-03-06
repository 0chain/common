package statecache

import (
	"sync"
	"time"

	"github.com/0chain/common/core/logging"
	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
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
	lock        sync.Mutex
	hits        int64
	miss        int64
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

func (sc *StateCache) commit(bc *BlockCache) {
	sc.lock.Lock()
	defer sc.lock.Unlock()

	_, ok := sc.hashCache.Get(bc.blockHash)
	if ok {
		// block already committed
		return
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()
	ts := time.Now()
	for key, v := range bc.cache {
		bvsi, ok := sc.cache.Get(key)
		if !ok {
			var err error
			bvsi, err = lru.New(200)
			if err != nil {
				panic(err)
			}
		}

		bvs := bvsi.(*lru.Cache)

		if v.data != nil {
			v.data = v.data.Clone()
		}
		bvs.Add(bc.blockHash, v)

		sc.cache.Add(key, bvs)
	}

	sc.commitRound(bc.round, bc.prevBlockHash, bc.blockHash)

	sc.hits += bc.hits
	sc.miss += bc.miss

	// Clear the pre-commit cache
	bc.cache = make(map[string]valueNode)
	logging.Logger.Debug("statecache - commit",
		zap.String("block", bc.blockHash),
		zap.Int64("bc_hits", bc.hits),
		zap.Int64("bc_miss", bc.miss),
		zap.Int64("sc_hits", sc.hits),
		zap.Int64("sc_miss", sc.miss),
		zap.Any("duration", time.Since(ts)))
}

// Get returns the value with the given key and block hash
func (sc *StateCache) Get(key, blockHash string) (Value, bool) {
	// sc.mu.RLock()
	// defer sc.mu.RUnlock()

	blockValues, ok := sc.cache.Get(key)
	if !ok {
		logging.Logger.Debug("state cache get - key not found", zap.String("key", key))
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
			logging.Logger.Debug("state cache - see gap", zap.String("block", blockHash))
			return nil, false
		}

		blockHash = prevHash.(string)
		vv, ok = bvs.Get(blockHash)
		if !ok {
			// stop if the value is not found in previous maxHisDepth rounds
			if count >= sc.maxHisDepth {
				logging.Logger.Debug("state cache - reach max depth", zap.String("block", blockHash))
				return nil, false
			}

			continue
		}

		v := vv.(valueNode)
		if v.deleted {
			logging.Logger.Debug("state cache - is deleted")
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

func (sc *StateCache) Stats() (hits int64, miss int64) {
	sc.lock.Lock()
	hits, miss = sc.hits, sc.miss
	sc.lock.Unlock()
	return
}
