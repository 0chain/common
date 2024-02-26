package statecache

import (
	"sync"

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
	mu              sync.RWMutex
	lastBreakRound  int64
	lastCommitRound int64
	// cache map[string]map[string]valueNode
	cache     *lru.Cache
	hashCache *lru.Cache
}

func NewStateCache() *StateCache {
	cache, err := lru.New(100 * 1024)
	if err != nil {
		panic(err)
	}

	hCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}

	return &StateCache{
		// cache: make(map[string]map[string]valueNode),
		cache:     cache,
		hashCache: hCache,
	}
}

func (sc *StateCache) commitRound(round int64, prevHash, blockHash string) {
	sc.mu.Lock()
	if sc.lastCommitRound+1 != round {
		sc.lastBreakRound = round // any valueNode with Round < lastBreakRound is stale
	}
	sc.lastCommitRound = round
	sc.mu.Unlock()
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
			// break if the value is not found in previous 100 rounds
			if count >= 100 {
				return nil, false
			}

			continue
		}

		v := vv.(valueNode)
		sc.mu.Lock()
		if v.round < sc.lastBreakRound {
			sc.mu.Unlock()
			// break if value round is less than last break round
			return nil, false
		}
		sc.mu.Unlock()

		if v.deleted {
			return nil, false
		}

		return v.data.Clone(), true
	}
}

// func (sc *StateCache) getValue(key, blockHash string) (valueNode, bool) {
// 	// sc.mu.RLock()
// 	// defer sc.mu.RUnlock()

// 	blockValues, ok := sc.cache.Get(key)
// 	if !ok {
// 		return valueNode{}, false
// 	}

// 	vv, ok := blockValues.(*lru.Cache).Get(blockHash)
// 	v := vv.(valueNode)
// 	if ok && !v.deleted {
// 		v.data = v.data.Clone()
// 		return v, true
// 	}
// 	return valueNode{}, false
// }

// shift copy the value from previous block to current
// func (sc *StateCache) shift(prevHash, blockHash string) {
// 	if prevHash == "" || blockHash == "" {
// 		return
// 	}
// 	tm := time.Now()

// 	keys := sc.cache.Keys()
// 	for _, key := range keys {
// 		blockValues, ok := sc.cache.Get(key)
// 		if ok {
// 			bvs := blockValues.(*lru.Cache)
// 			vv, ok := bvs.Get(prevHash)
// 			if ok {
// 				// shift the value from previous block when it does not exist in present block
// 				if _, exist := bvs.Get(blockHash); !exist {
// 					v := vv.(valueNode)
// 					v.data = v.data.Clone()
// 					bvs.Add(blockHash, v)
// 					sc.cache.Add(key, bvs)
// 				}
// 			}
// 		}
// 	}

// for key, blockValues := range sc.cache {
// 	v, ok := blockValues[prevHash]
// 	if ok {
// 		if _, exists := blockValues[blockHash]; !exists {
// 			if sc.cache[key] == nil {
// 				sc.cache[key] = make(map[string]valueNode)
// 			}
// 			v.data = v.data.Clone()
// 			sc.cache[key][blockHash] = v
// 		}
// 	}
// }

// logging.Logger.Debug("state cache - shift", zap.Any("duration", time.Since(tm)))
// }

// Remove removes the values map with the given key
func (sc *StateCache) Remove(key string) {
	// sc.mu.Lock()
	// defer sc.mu.Unlock()

	sc.cache.Remove(key)
	// delete(sc.cache, key)
}

// PruneRoundBelow removes all values that are below the given round
// func (sc *StateCache) PruneRoundBelow(round int64) {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	var count int
// 	for key, blockValues := range sc.cache {
// 		for blockHash, value := range blockValues {
// 			if value.round < round {
// 				count++
// 				delete(blockValues, blockHash)
// 			}
// 		}

// 		// Delete the map if it becomes empty
// 		if len(blockValues) == 0 {
// 			delete(sc.cache, key)
// 		}
// 	}

// 	logging.Logger.Debug("state cache - prune_round_below", zap.Int64("round", round), zap.Int("count", count))
// }

// PrettyPrint prints the state cache in a pretty format
// func (sc *StateCache) PrettyPrint() {
// 	sc.mu.RLock()
// 	defer sc.mu.RUnlock()

// 	// Sort keys in alphabetical order
// 	var keys []string
// 	for key := range sc.cache {
// 		keys = append(keys, key)
// 	}
// 	sort.Strings(keys)

// 	// Print values for each key
// 	for _, key := range keys {
// 		fmt.Printf("Key: %s\n", key)

// 		blockValues := sc.cache[key]

// 		// Sort block hashes by round number in descending order
// 		var rounds []int64
// 		for _, value := range blockValues {
// 			rounds = append(rounds, value.round)
// 		}
// 		sort.Slice(rounds, func(i, j int) bool {
// 			return rounds[i] > rounds[j]
// 		})

// 		// Print values for each round
// 		for _, round := range rounds {
// 			fmt.Printf("  Round: %d\n", round)

// 			// Sort block hashes for the same round
// 			var hashes []string
// 			for hash, value := range blockValues {
// 				if value.round == round {
// 					hashes = append(hashes, hash)
// 				}
// 			}
// 			sort.Strings(hashes)

// 			// Print values for each hash
// 			for _, hash := range hashes {
// 				value := blockValues[hash]
// 				fmt.Printf("    Hash: %s\n", hash)
// 				fmt.Printf("      Deleted: %v\n", value.deleted)
// 			}
// 		}
// 	}
// }
