package wmpt

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeHashNode(t *testing.T) {
	hash := sha256.Sum256([]byte("hello"))
	node := hashNode{hash: hash[:], weight: 100000000000}
	data, err := node.Serialize()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(data))
}
