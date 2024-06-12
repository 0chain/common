package statecheck

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Add unit tests for StateCheck

func TestAdd(t *testing.T) {
	sck := NewStateCheck()
	v := []byte("b")
	err := sck.Add([]byte("a"), &v)
	require.NoError(t, err)

	t.Run("none pointer value", func(t *testing.T) {
		v := []byte("b")
		err := sck.Add([]byte("a"), v)
		require.Error(t, err)
	})
}

func TestGet(t *testing.T) {
	sck := NewStateCheck()
	v := []byte("b")
	err := sck.Add([]byte("a"), &v)
	require.NoError(t, err)

	vb, err := sck.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, vb, &v)

	// update the acquired value should be reflected on the state check
	*vb.(*[]byte) = []byte("c")

	require.Equal(t, []byte("c"), *sck.stateNodes["a"].(*[]byte))

	// get missing value

	_, err = sck.Get([]byte("b"))
	require.Error(t, err)
}

func TestForEach(t *testing.T) {
	sck := NewStateCheck()
	v := []byte("b")
	err := sck.Add([]byte("a"), &v)
	require.NoError(t, err)
	err = sck.Add([]byte("b"), &v)
	require.NoError(t, err)

	var count int
	err = sck.ForEach(func(key []byte, value interface{}) error {
		count++
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 2, count)
}
