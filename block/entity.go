package block

import (
	"github.com/0chain/common/datastore"
	"github.com/0chain/common/util"
	"github.com/0chain/common/transaction"
	"sync"
)


// UnverifiedBlockBody - used to compute the signature
// This is what is used to verify the correctness of the block & the associated signature
type UnverifiedBlockBody struct {
	datastore.VersionField
	datastore.CreationDateField

	LatestFinalizedMagicBlockHash  string                `json:"latest_finalized_magic_block_hash"`
	LatestFinalizedMagicBlockRound int64                 `json:"latest_finalized_magic_block_round"`
	PrevHash                       string                `json:"prev_hash"`
	PrevBlockVerificationTickets   []*VerificationTicket `json:"prev_verification_tickets,omitempty"`

	MinerID           datastore.Key `json:"miner_id"`
	Round             int64         `json:"round"`
	RoundRandomSeed   int64         `json:"round_random_seed"`
	RoundTimeoutCount int           `json:"round_timeout_count"`

	ClientStateHash util.Key `json:"state_hash"`

	// The entire transaction payload to represent full block
	Txns []*transaction.Transaction `json:"transactions,omitempty"`
}

/*Block - data structure that holds the block data */
type Block struct {
	UnverifiedBlockBody
	VerificationTickets []*VerificationTicket `json:"verification_tickets,omitempty"`

	datastore.HashIDField
	Signature string `json:"signature"`

	ChainID   datastore.Key `json:"chain_id"`
	RoundRank int           `json:"-" msgpack:"-"` // rank of the block in the round it belongs to
	PrevBlock *Block        `json:"-" msgpack:"-"`
	Events    []event.Event

	TxnsMap   map[string]bool `json:"-" msgpack:"-"`
	mutexTxns sync.RWMutex    `json:"-" msgpack:"-"`

	ClientState           util.MerklePatriciaTrieI `json:"-" msgpack:"-"`
	stateStatus           int8
	stateStatusMutex      sync.RWMutex `json:"-" msgpack:"-"`
	stateMutex            sync.RWMutex `json:"-" msgpack:"-"`
	blockState            int8
	isNotarized           bool
	ticketsMutex          sync.RWMutex `json:"-" msgpack:"-"`
	verificationStatus    int
	RunningTxnCount       int64           `json:"running_txn_count"`
	UniqueBlockExtensions map[string]bool `json:"-" msgpack:"-"`
	*MagicBlock           `json:"magic_block,omitempty" msgpack:"mb,omitempty"`
	// StateChangesCount represents the state changes number in client state of current block.
	// this will be used to verify the state changes acquire from remote
	StateChangesCount int `json:"state_changes_count"`
}

