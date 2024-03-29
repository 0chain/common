//go:build dev
// +build dev

package util

import (
	"github.com/linxGnu/grocksdb"
)

func init() {
	// https://github.com/facebook/rocksdb/issues/814
	// there is an issue to build rocksdb with lz4 support on MacOS. so let's disable it on local debugging
	PNodeDBCompression = grocksdb.NoCompression
}
