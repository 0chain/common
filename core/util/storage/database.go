package storage

type StorageAdapter interface {
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Has([]byte) bool
	Delete([]byte) error
	Close()
	NewBatch() Batcher
}

type Batcher interface {
	Put([]byte, []byte) error
	Delete([]byte) error
	Commit(bool) error
}
