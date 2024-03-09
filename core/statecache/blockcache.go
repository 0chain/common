package statecache

import (
	"sync"
	"sync/atomic"

	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

type BlockCacher interface {
	Get(key string) (Value, bool)
	Round() int64
	Commit()
	setValue(key string, v valueNode)
	addStats(hit, miss int64)
}

// BlockCache is a pre commit cache for all changes in a block.
// This is mainly for caching values in current block when executing blocks.
//
// Call `Commit()` method to merge
// the changes to the StateCache when the block is executed.
type BlockCache struct {
	mu            sync.Mutex
	cache         map[string]valueNode
	main          *StateCache
	blockHash     string
	prevBlockHash string
	round         int64
	hits          int64
	miss          int64
}

type Block struct {
	Round    int64  // round number when this block cache is created
	Hash     string // block hash
	PrevHash string // previous hash of the block
}

func NewBlockCache(main *StateCache, b Block) *BlockCache {
	return &BlockCache{
		cache:         make(map[string]valueNode),
		main:          main,
		blockHash:     b.Hash,
		prevBlockHash: b.PrevHash,
		round:         b.Round,
	}
}

// Set sets the value with the given key in the pre-commit cache
func (pcc *BlockCache) Set(key string, e Value) {
	pcc.mu.Lock()

	pcc.cache[key] = valueNode{
		data: e.Clone(),
	}
	pcc.mu.Unlock()
}

func (pcc *BlockCache) Round() int64 {
	return pcc.round
}

func (pcc *BlockCache) setValue(key string, v valueNode) {
	pcc.mu.Lock()
	defer pcc.mu.Unlock()

	v.data = v.data.Clone()
	pcc.cache[key] = v
}

// Get returns the value with the given key
func (pcc *BlockCache) Get(key string) (Value, bool) {
	pcc.mu.Lock()
	defer pcc.mu.Unlock()

	// Check the cache first
	value, ok := pcc.cache[key]
	if ok && !value.deleted {
		// logging.Logger.Debug("block cache get", zap.String("key", key))
		return value.data.Clone(), true
	}

	// Should not return deleted value
	if ok && value.deleted {
		// logging.Logger.Debug("block cache get - deleted", zap.String("key", key))
		logging.Logger.Debug("block state cache - deleted", zap.String("block", pcc.blockHash))
		return nil, false
	}

	return pcc.main.Get(key, pcc.prevBlockHash)

	// v, ok := pcc.main.Get(key, pcc.prevBlockHash)
	// if ok {
	// 	// load the value from the state cache and store it in the block cache
	// 	vn := valueNode{
	// 		data:          v,
	// 		round:         pcc.round,
	// 		prevBlockHash: pcc.prevBlockHash,
	// 	}

	// 	pcc.cache[key] = vn
	// 	return v, true
	// }

	// return nil, false
}

// Remove marks the value with the given key as deleted in the pre-commit cache
func (pcc *BlockCache) remove(key string) {
	pcc.mu.Lock()
	defer pcc.mu.Unlock()

	value, ok := pcc.cache[key]
	if ok {
		value.deleted = true
		pcc.cache[key] = value
		return
	} else {
		pcc.cache[key] = valueNode{
			deleted: true,
		}
	}
}

func (pcc *BlockCache) addStats(hit, miss int64) {
	atomic.AddInt64(&pcc.hits, hit)
	atomic.AddInt64(&pcc.miss, miss)
}

func (pcc *BlockCache) Stats() (hit, miss int64) {
	return atomic.LoadInt64(&pcc.hits), atomic.LoadInt64(&pcc.miss)
}

// SetBlockHash sets the block hash, which is used after miners generating the block.
// miners generator does not know the block hash until the block is generated.
func (pcc *BlockCache) SetBlockHash(hash string) {
	pcc.mu.Lock()
	pcc.blockHash = hash
	pcc.mu.Unlock()
}

// Commit moves the values from the pre-commit cache to the main cache
func (pcc *BlockCache) Commit() {
	pcc.main.commit(pcc)
}
