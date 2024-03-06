package statecache

import (
	"sync"
	"sync/atomic"
)

type TransactionCache struct {
	main  BlockCacher
	cache map[string]valueNode
	mu    sync.RWMutex
	round int64
	hit   int64
	miss  int64
}

func NewTransactionCache(main BlockCacher) *TransactionCache {
	return &TransactionCache{
		main:  main,
		cache: make(map[string]valueNode),
		round: main.Round(),
	}
}

// NewEmpty creates a new empty transaction cache
func NewEmpty() *TransactionCache {
	_, tc := NewBlockTxnCaches(NewStateCache(), Block{})
	return tc
}

func (tc *TransactionCache) Set(key string, e Value) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// logging.Logger.Debug("txn cache set", zap.String("key", key))
	tc.cache[key] = valueNode{
		data:  e.Clone(),
		round: tc.round,
	}
}

func (tc *TransactionCache) Get(key string) (Value, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	value, ok := tc.cache[key]
	if ok {
		// logging.Logger.Debug("txn cache get", zap.String("key", key))
		return value.data.Clone(), ok
	}

	return tc.main.Get(key)
}

type EmptyValue struct{}

func (e *EmptyValue) Clone() Value {
	return e
}

func (e *EmptyValue) CopyFrom(interface{}) bool {
	return true
}

func (tc *TransactionCache) Remove(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	// logging.Logger.Debug("txn cache remove", zap.String("key", key))

	value, ok := tc.cache[key]
	if ok {
		value.deleted = true
		value.data = value.data.Clone()
		tc.cache[key] = value
	} else {
		tc.cache[key] = valueNode{
			deleted: true,
			round:   tc.round,
			data:    &EmptyValue{},
		}
	}
}

func (tc *TransactionCache) Commit() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// var count int
	for key, value := range tc.cache {
		tc.main.setValue(key, value)
		// logging.Logger.Debug("transaction cache commit",
		// 	zap.String("key", key),
		// 	zap.Int64("round", value.round),
		// 	zap.Bool("deleted", value.deleted))
		// count++
	}

	// logging.Logger.Debug("transaction cache commit - total", zap.Int("count", count))

	// Clear the transaction cache
	tc.main.addStats(tc.hit, tc.miss)
	tc.cache = make(map[string]valueNode)
}

func (tc *TransactionCache) AddHit() {
	atomic.AddInt64(&tc.hit, 1)
}

func (tc *TransactionCache) AddMiss() {
	atomic.AddInt64(&tc.miss, 1)
}

func (tc *TransactionCache) Stats() (hit, miss int64) {
	return atomic.LoadInt64(&tc.hit), atomic.LoadInt64(&tc.miss)
}
