package encryption

import (
	"encoding/hex"

	"github.com/0chain/common/core/common"
	"golang.org/x/crypto/sha3"
)

// ErrInvalidHash - hash is invalid error
var ErrInvalidHash = common.NewError("invalid_hash", "Invalid hash")

const HASH_LENGTH = 32

type HashBytes [HASH_LENGTH]byte

/*Hash - hash the given data and return the hash as hex string */
func Hash(data interface{}) string {
	return hex.EncodeToString(RawHash(data))
}

func IsHash(str string) bool {
	bytes, err := hex.DecodeString(str)
	return err == nil && len(bytes) == HASH_LENGTH
}

// EmptyHash - hash of an empty string
var EmptyHash = Hash("")

// EmptyHashBytes - hash bytes of an empty string
var EmptyHashBytes = RawHash("")

/*RawHash - Logic to hash the text and return the hash bytes */
func RawHash(data interface{}) []byte {
	var databuf []byte
	switch dataImpl := data.(type) {
	case []byte:
		databuf = dataImpl
	case HashBytes:
		databuf = dataImpl[:]
	case string:
		databuf = []byte(dataImpl)
	default:
		panic("unknown type")
	}
	hash := sha3.New256()
	hash.Write(databuf)
	buf := make([]byte, 0, hash.Size())
	return hash.Sum(buf)
}
