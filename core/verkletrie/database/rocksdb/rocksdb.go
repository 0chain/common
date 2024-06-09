// implements the database interface for verkle tree to store nodes
package rocksdb

// type BatchWriter interface {
// 	Put(key []byte, value []byte) error
// 	Write() error
// 	Reset()
// 	Size() int
// }

// type DB interface {
// 	Set(key []byte, value []byte) error
// 	Get(key []byte) ([]byte, error)
// 	NewBatch() BatchWriter
// }

// package main

import (
	"errors"

	"github.com/0chain/common/core/verkletrie/database"

	"github.com/linxGnu/grocksdb"
)

// BatchWriter interface definition
// type BatchWriter interface {
// 	Put(key []byte, value []byte) error
// 	Write() error
// 	Reset()
// 	Size() int
// }

// // DB interface definition
// type DB interface {
// 	Set(key []byte, value []byte) error
// 	Get(key []byte) ([]byte, error)
// 	NewBatch() BatchWriter
// }

// rocksBatchWriter implements the BatchWriter interface using RocksDB
type rocksBatchWriter struct {
	db    *grocksdb.DB
	wo    *grocksdb.WriteOptions
	batch *grocksdb.WriteBatch
	size  int
}

func (w *rocksBatchWriter) Put(key []byte, value []byte) error {
	w.batch.Put(key, value)
	w.size++
	return nil
}

func (w *rocksBatchWriter) Write() error {
	if w.size == 0 {
		return errors.New("batch is empty")
	}
	return w.db.Write(w.wo, w.batch)
}

func (w *rocksBatchWriter) Reset() {
	w.batch.Clear()
	w.size = 0
}

func (w *rocksBatchWriter) Size() int {
	return w.size
}

// rocksDB implements the DB interface
type rocksDB struct {
	db *grocksdb.DB
	wo *grocksdb.WriteOptions
	ro *grocksdb.ReadOptions
}

func (r *rocksDB) Set(key []byte, value []byte) error {
	return r.db.Put(r.wo, key, value)
}

func (r *rocksDB) Get(key []byte) ([]byte, error) {
	slice, err := r.db.Get(r.ro, key)
	if err != nil {
		return nil, database.ErrMissingNode
	}

	if !slice.Exists() {
		return nil, database.ErrMissingNode
	}

	data := slice.Data()
	if len(data) == 0 {
		return nil, database.ErrMissingNode
	}

	ret := make([]byte, len(data))
	copy(ret, data)
	defer slice.Free()
	return ret, nil
}

func (r *rocksDB) NewBatch() database.BatchWriter {
	return &rocksBatchWriter{
		db:    r.db,
		wo:    r.wo,
		batch: grocksdb.NewWriteBatch(),
		size:  0,
	}
}

func (r *rocksDB) Close() {
	r.db.Close()
}

func defaultDBOptions() *grocksdb.Options {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))
	opts := grocksdb.NewDefaultOptions()
	opts.SetKeepLogFileNum(5)
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	return opts
}

func NewRocksDB(path string) (database.DB, error) {
	// opts := grocksdb.NewDefaultOptions()
	// opts.SetCreateIfMissing(true)
	opts := defaultDBOptions()
	db, err := grocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}

	return &rocksDB{
		db: db,
		wo: grocksdb.NewDefaultWriteOptions(),
		ro: grocksdb.NewDefaultReadOptions(),
	}, nil
}
