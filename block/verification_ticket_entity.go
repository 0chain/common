package block

import (
	"github.com/0chain/common/datastore"
)

/*VerificationTicket - verification ticket for the block */
type VerificationTicket struct {
	VerifierID datastore.Key `json:"verifier_id" msgpack:"v_id"`
	Signature  string        `json:"signature" msgpack:"sig"`
}
