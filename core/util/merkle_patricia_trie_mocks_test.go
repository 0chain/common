package util_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/0chain/common/core/logging"
	"github.com/0chain/common/core/mocks"
	"github.com/0chain/common/core/statecache"
	. "github.com/0chain/common/core/util"
	"github.com/stretchr/testify/mock"
)

func init() {
	logging.Logger = zap.NewNop()
}

func TestMPTSaveChanges(t *testing.T) {
	mockNodeDB1 := &mocks.NodeDB{}
	mockNodeDB1.On("PutNode", mock.Anything, mock.Anything).Return(nil)
	mpt := NewMerklePatriciaTrie(mockNodeDB1, Sequence(0), nil, statecache.NewEmpty())
	_, err := mpt.Insert(Path("key"), &Txn{"value"})
	require.NoError(t, err)
	mockNodeDB2 := &mocks.NodeDB{}
	mockNodeDB2.On("MultiPutNode", mock.Anything, mock.Anything).Return(errors.New("Failure"))
	err2 := mpt.SaveChanges(context.TODO(), mockNodeDB2, false)
	require.Error(t, err2) // expected error
}
