package wmpt

type (
	PersistNodeBase struct {
		Branch   *PersistNodeBranch `cbor:"10,keyasint,omitempty"`
		Value    *PersistNodeValue  `cbor:"11,keyasint,omitempty"`
		Short    *PersistNodeShort  `cbor:"12,keyasint,omitempty"`
		NilNode  *PersistNilNode    `cbor:"13,keyasint,omitempty"`
		HashNode *PersistHashNode   `cbor:"14,keyasint,omitempty"`
	}

	PersistNodeBranch struct {
		_        struct{} `cbor:",toarray"`
		Hash     []byte
		Children [][]byte `cbor:"omitempty"`
	}
	PersistNodeValue struct {
		_      struct{} `cbor:",toarray"`
		Value  []byte
		Hash   []byte
		Weight uint64
	}
	PersistNodeShort struct {
		_     struct{} `cbor:",toarray"`
		Key   []byte
		Hash  []byte
		Value []byte
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
