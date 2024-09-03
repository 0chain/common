package trie

type (
	PersistNodeBase struct {
		Branch   *PersistNodeBranch `cbor:"10,keyasint,omitempty"`
		Value    *PersistNodeValue  `cbor:"11,keyasint,omitempty"`
		NilNode  *PersistNilNode    `cbor:"12,keyasint,omitempty"`
		HashNode *PersistHashNode   `cbor:"13,keyasint,omitempty"`
	}

	PersistNodeBranch struct {
		_      struct{} `cbor:",toarray"`
		Weight uint64
		Key    []byte
		Hash   []byte
	}
	PersistNodeValue struct {
		_      struct{} `cbor:",toarray"`
		Value  []byte
		Weight uint64
		Key    []byte
		Hash   []byte
	}
	PersistNilNode struct {
	}
	PersistHashNode struct {
		_      struct{} `cbor:",toarray"`
		Hash   []byte
		Weight uint64
	}
	PersistTrie struct {
		_     struct{} `cbor:",toarray"`
		Pairs []*PersistTriePair
	}
	PersistTriePair struct {
		_     struct{} `cbor:",toarray"`
		Value []byte
	}
)
