
package bls

import (
	"github.com/0chain/common/datastore"
)

//go:generate msgp -io=false

type DKGKeyShare struct {
	datastore.IDField
	Message string `json:"message"`
	Share   string `json:"share"`
	Sign    string `json:"sign"`
}