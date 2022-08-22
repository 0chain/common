package state

import (
	"github.com/0chain/common/datastore"
	"github.com/0chain/common/util"
	"github.com/0chain/common/block"
	"github.com/0chain/common/transaction"
	// "github.com/0chain/common/chain/state"
)

/*
* The state context is available to the smart contract logic.
* The smart contract logic can use
*    GetClientBalance - to get the balance of a client at the beginning of executing the transaction.
*    AddTransfer - to add transfer of tokens from one client to another.
*  Restrictions:
*    1) The total transfer out from the txn.ClientID should be <= txn.Value
*    2) The only from clients valid are txn.ClientID and txn.ToClientID (which will be the smart contract's client id)
 */

 type CommonStateContextI interface {
	GetTrieNode(key datastore.Key, v util.MPTSerializable) error
	GetBlock() *block.Block
	GetLatestFinalizedBlock() *block.Block
}

//go:generate mockery --case underscore --name=QueryStateContextI --output=./mocks
type QueryStateContextI interface {
	CommonStateContextI
	GetEventDB() *event.EventDb
}

type Appender func(events []event.Event, current event.Event) []event.Event

type StateContextI interface {
	QueryStateContextI
	GetLastestFinalizedMagicBlock() *block.Block
	GetChainCurrentMagicBlock() *block.MagicBlock
	SetMagicBlock(block *block.MagicBlock)    // cannot use in smart contracts or REST endpoints
	GetState() util.MerklePatriciaTrieI       // cannot use in smart contracts or REST endpoints
	GetTransaction() *transaction.Transaction // cannot use in smart contracts or REST endpoints
	GetClientBalance(clientID datastore.Key) (currency.Coin, error)
	SetStateContext(st *state.State) error // cannot use in smart contracts or REST endpoints
	InsertTrieNode(key datastore.Key, node util.MPTSerializable) (datastore.Key, error)
	DeleteTrieNode(key datastore.Key) (datastore.Key, error)
	AddTransfer(t *state.Transfer) error
	AddSignedTransfer(st *state.SignedTransfer)
	AddMint(m *state.Mint) error
	GetTransfers() []*state.Transfer // cannot use in smart contracts or REST endpoints
	GetSignedTransfers() []*state.SignedTransfer
	GetMints() []*state.Mint // cannot use in smart contracts or REST endpoints
	Validate() error
	GetBlockSharders(b *block.Block) []string
	GetSignatureScheme() encryption.SignatureScheme
	GetLatestFinalizedBlock() *block.Block
	EmitEvent(event.EventType, event.EventTag, string, interface{}, ...Appender)
	EmitError(error)
	GetEvents() []event.Event // cannot use in smart contracts or REST endpoints
}
