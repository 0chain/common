package database

import (
	"errors"
)

const IdealBatchSize = 100 * 1024

var ErrMissingNode = errors.New("missing node")

type BatchWriter interface {
	Put(key []byte, value []byte) error
	Write() error
	Reset()
	Size() int
}

type DB interface {
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	NewBatch() BatchWriter
	Close()
}

type InMemoryDB struct {
	store map[string][]byte
}

func NewInMemoryVerkleDB() *InMemoryDB {
	return &InMemoryDB{
		store: make(map[string][]byte),
	}
}

func (m *InMemoryDB) Set(key []byte, value []byte) error {
	m.store[string(key)] = value
	return nil
}

func (m *InMemoryDB) Get(key []byte) ([]byte, error) {
	v, ok := m.store[string(key)]
	if !ok {
		return nil, ErrMissingNode
	}

	return v, nil
}

func (m *InMemoryDB) NewBatch() BatchWriter {
	return &InMemoryBatch{
		store: make(map[string][]byte),
		db:    m,
	}
}

func (m *InMemoryDB) Close() {
}

type InMemoryBatch struct {
	store map[string][]byte
	db    *InMemoryDB
}

func (m *InMemoryBatch) Put(key []byte, value []byte) error {
	m.store[string(key)] = value
	return nil
}

func (m *InMemoryBatch) Write() error {
	for k, v := range m.store {
		m.db.Set([]byte(k), v)
	}
	return nil
}

func (m *InMemoryBatch) Reset() {
	m.store = make(map[string][]byte)
}

func (m *InMemoryBatch) Size() int {
	return len(m.store)
}
