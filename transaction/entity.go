package transaction

import (
	"github.com/0chain/common/common"
	"github.com/0chain/common/currency"
	"github.com/0chain/common/datastore"
)

/*Transaction type for capturing the transaction data */
type Transaction struct {
	datastore.HashIDField
	datastore.CollectionMemberField `json:"-" msgpack:"-"`
	datastore.VersionField

	ClientID  string `json:"client_id" msgpack:"cid,omitempty"`
	PublicKey string `json:"public_key,omitempty" msgpack:"puk,omitempty"`

	ToClientID      string           `json:"to_client_id,omitempty" msgpack:"tcid,omitempty"`
	ChainID         string           `json:"chain_id,omitempty" msgpack:"chid"`
	TransactionData string           `json:"transaction_data" msgpack:"d"`
	Value           currency.Coin    `json:"transaction_value" msgpack:"v"` // The value associated with this transaction
	Signature       string           `json:"signature" msgpack:"s"`
	CreationDate    common.Timestamp `json:"creation_date" msgpack:"ts"`
	Fee             currency.Coin    `json:"transaction_fee" msgpack:"f"`
	Nonce           int64            `json:"transaction_nonce" msgpack:"n"`

	TransactionType   int    `json:"transaction_type" msgpack:"tt"`
	TransactionOutput string `json:"transaction_output,omitempty" msgpack:"o,omitempty"`
	OutputHash        string `json:"txn_output_hash" msgpack:"oh"`
	Status            int    `json:"transaction_status" msgpack:"sot"`
}
