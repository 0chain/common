package block

import (
	"github.com/0chain/common/corechain/threshold/bls"
)

//go:generate msgp -io=false -tests=false -v

type ShareOrSigns struct {
	ID           string                      `json:"id"`
	ShareOrSigns map[string]*bls.DKGKeyShare `json:"share_or_sign"`
}
