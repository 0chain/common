package storage

import (
	"runtime"

	"github.com/cockroachdb/pebble"
)

type PebbleAdapter struct {
	db *pebble.DB
}

func NewPebbleAdapter(path string) (*PebbleAdapter, error) {
	opts := &pebble.Options{
		Cache:                    pebble.NewCache(1024 * 1024 * 1024),
		MaxConcurrentCompactions: func() int { return runtime.GOMAXPROCS(0) },
	}
	db, err := pebble.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &PebbleAdapter{db: db}, nil
}

func (p *PebbleAdapter) Get(key []byte) ([]byte, error) {
	dat, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, len(dat))
	copy(ret, dat)
	closer.Close()
	return ret, nil
}

func (p *PebbleAdapter) Close() {
	p.db.Close()
}

func (p *PebbleAdapter) Put(key []byte, value []byte) error {
	return p.db.Set(key, value, pebble.NoSync)
}

type batch struct {
	b *pebble.Batch
}

func (p *PebbleAdapter) NewBatch() Batcher {
	b := p.db.NewBatch()
	return &batch{b: b}
}

func (b *batch) Put(key []byte, value []byte) error {
	return b.b.Set(key, value, pebble.NoSync)
}

func (b *batch) Commit(sync bool) error {
	if sync {
		return b.b.Commit(pebble.Sync)
	}
	return b.b.Commit(pebble.NoSync)
}

func (b *batch) Delete(key []byte) error {
	return b.b.Delete(key, pebble.NoSync)
}
