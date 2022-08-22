package block

import (
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"sync"

	"0chain.net/chaincore/node"
	"github.com/0chain/common/datastore"
	"0chain.net/core/encryption"
	"0chain.net/core/util"
)

//go:generate msgp -io=false -tests=false -v

// swagger:model MagicBlock
type MagicBlock struct {
	datastore.HashIDField
	mutex                  sync.RWMutex        `json:"-" msgpack:"-" msg:"-"`
	PreviousMagicBlockHash string              `json:"previous_hash"`
	MagicBlockNumber       int64               `json:"magic_block_number"`
	StartingRound          int64               `json:"starting_round"`
	Miners                 *node.Pool          `json:"miners"`   //this is the pool of miners participating in the blockchain
	Sharders               *node.Pool          `json:"sharders"` //this is the pool of sharders participaing in the blockchain
	ShareOrSigns           *GroupSharesOrSigns `json:"share_or_signs"`
	Mpks                   *Mpks               `json:"mpks"`
	T                      int                 `json:"t"`
	K                      int                 `json:"k"`
	N                      int                 `json:"n"`
}