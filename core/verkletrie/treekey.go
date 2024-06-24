package verkletrie

import (
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
	"github.com/ethereum/go-verkle"
	"github.com/holiman/uint256"
)

const (
	// the position between 2 and 63 are researved in case we need to define new fields
	VERSION_LEAF_KEY      = 0   // keys ending in 00: version (could be allocation version), which is always set to 0
	FILE_HASH_LEAF_KEY    = 1   // keys ending in 01: file hash, could be the root hash of the file content
	STORAGE_SIZE_LEAF_KEY = 2   // keys ending in 10: the size of the storage
	HEADER_STORAGE_OFFSET = 16  // the offset of the storage in the tree
	VERKLE_NODE_WIDTH     = 256 // the width of the verkle node
	// MAIN_STORAGE_OFFSET   = 256 ^ 31 // the offset of the main storage in the tree
)

var (
	getTreePolyIndex0Point *verkle.Point
)

var (
	zero                = uint256.NewInt(0)
	headerStorageOffset = uint256.NewInt(HEADER_STORAGE_OFFSET)
	verkleNodeWidth     = uint256.NewInt(VERKLE_NODE_WIDTH)
	headerStorageCap    = VERKLE_NODE_WIDTH - HEADER_STORAGE_OFFSET // the capacity of the header storage, which is (256-16)*32=7168 Bytes(7KB)
	mainStorageOffset   = new(uint256.Int).Lsh(uint256.NewInt(1), 248 /* 8 * 31*/)
)

// GetTreeKeyForFileHash returns file hash
func GetTreeKeyForFileHash(filepathHash []byte) []byte {
	return GetTreeKey(filepathHash, zero, FILE_HASH_LEAF_KEY)
}

func GetTreeKeyForStorageSize(filepathHash []byte) []byte {
	return GetTreeKey(filepathHash, zero, STORAGE_SIZE_LEAF_KEY)
}

func GetTreeKeyForStorageSlot(filepathHash []byte, storageKey uint64) []byte {
	pos := uint256.NewInt(storageKey)
	if storageKey < uint64(headerStorageCap) {
		// storage in the header
		pos.Add(headerStorageOffset, pos)
	} else {
		// stroage in the main storage
		pos.Add(mainStorageOffset, pos)
	}

	subIdx := uint256.NewInt(0)
	pos.DivMod(pos, verkleNodeWidth, subIdx)
	if subIdx.Eq(zero) {
		return GetTreeKey(filepathHash, pos, 0)
	} else {
		return GetTreeKey(filepathHash, pos, subIdx.Bytes()[0])
	}
}

func init() {
	// The byte array is the Marshalled output of the point computed as such:
	//cfg, _ := verkle.GetConfig()
	//verkle.FromLEBytes(&getTreePolyIndex0Fr[0], []byte{2, 64})
	//= cfg.CommitToPoly(getTreePolyIndex0Fr[:], 1)
	getTreePolyIndex0Point = new(verkle.Point)
	err := getTreePolyIndex0Point.SetBytes([]byte{34, 25, 109, 242, 193, 5, 144, 224, 76, 52, 189, 92, 197, 126, 9, 145, 27, 152, 199, 130, 165, 3, 210, 27, 193, 131, 142, 28, 110, 26, 16, 191})
	if err != nil {
		panic(err)
	}
}

// https://github.com/gballet/go-ethereum/blob/a586f0d253bdaec8bc0ee5849fe5f3137ef0ab43/trie/utils/verkle.go#L95
// GetTreeKey performs both the work of the spec's get_tree_key function, and that
// of pedersen_hash: it builds the polynomial in pedersen_hash without having to
// create a mostly zero-filled buffer and "type cast" it to a 128-long 16-byte
// array. Since at most the first 5 coefficients of the polynomial will be non-zero,
// these 5 coefficients are created directly.
func GetTreeKey(address []byte, treeIndex *uint256.Int, subIndex byte) []byte {
	if len(address) < 32 {
		var aligned [32]byte
		address = append(aligned[:32-len(address)], address...)
	}

	// poly = [2+256*64, address_le_low, address_le_high, tree_index_le_low, tree_index_le_high]
	var poly [5]fr.Element

	// 32-byte address, interpreted as two little endian
	// 16-byte numbers.
	verkle.FromLEBytes(&poly[1], address[:16])
	verkle.FromLEBytes(&poly[2], address[16:])

	// treeIndex must be interpreted as a 32-byte aligned little-endian integer.
	// e.g: if treeIndex is 0xAABBCC, we need the byte representation to be 0xCCBBAA00...00.
	// poly[3] = LE({CC,BB,AA,00...0}) (16 bytes), poly[4]=LE({00,00,...}) (16 bytes).
	//
	// To avoid unnecessary endianness conversions for go-ipa, we do some trick:
	// - poly[3]'s byte representation is the same as the *top* 16 bytes (trieIndexBytes[16:]) of
	//   32-byte aligned big-endian representation (BE({00,...,AA,BB,CC})).
	// - poly[4]'s byte representation is the same as the *low* 16 bytes (trieIndexBytes[:16]) of
	//   the 32-byte aligned big-endian representation (BE({00,00,...}).
	trieIndexBytes := treeIndex.Bytes32()
	verkle.FromBytes(&poly[3], trieIndexBytes[16:])
	verkle.FromBytes(&poly[4], trieIndexBytes[:16])

	cfg := verkle.GetConfig()
	ret := cfg.CommitToPoly(poly[:], 0)

	// add a constant point corresponding to poly[0]=[2+256*64].
	ret.Add(ret, getTreePolyIndex0Point)

	return pointToHash(ret, subIndex)
}

func pointToHash(evaluated *verkle.Point, suffix byte) []byte {
	retb := verkle.HashPointToBytes(evaluated)
	retb[31] = suffix
	return retb[:]
}
