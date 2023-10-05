package util

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"

	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

/*PNodeDB - a node db that is persisted */
type PNodeDB struct {
	db *grocksdb.DB

	ro      *grocksdb.ReadOptions
	wo      *grocksdb.WriteOptions
	to      *grocksdb.TransactionOptions
	fo      *grocksdb.FlushOptions
	mutex   sync.Mutex
	version int64

	defaultCFH   *grocksdb.ColumnFamilyHandle
	deadNodesCFH *grocksdb.ColumnFamilyHandle
}

const (
	SSTTypeBlockBasedTable = 0
	SSTTypePlainTable      = 1
)

var (
	PNodeDBCompression = grocksdb.LZ4Compression
	deadNodesKey       = []byte("dead_nodes")
)

var sstType = SSTTypeBlockBasedTable

func newDefaultCFOptions(logDir string) *grocksdb.Options {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCompression(PNodeDBCompression)
	opts.OptimizeUniversalStyleCompaction(64 * 1024 * 1024)
	if sstType == SSTTypePlainTable {
		opts.SetAllowMmapReads(true)
		opts.SetPrefixExtractor(grocksdb.NewFixedPrefixTransform(6))
		opts.SetPlainTableFactory(32, 10, 0.75, 16)
	} else {
		opts.OptimizeForPointLookup(64)
		opts.SetAllowMmapReads(true)
		opts.SetPrefixExtractor(grocksdb.NewFixedPrefixTransform(6))
		opts.SetMaxBackgroundJobs(4)               // default was 2, double to 4
		opts.SetMaxWriteBufferNumber(4)            // default was 2, double to 4
		opts.SetWriteBufferSize(128 * 1024 * 1024) // default was 64M, double to 128M
		opts.SetMinWriteBufferNumberToMerge(2)     // default was 1, double to 2
	}
	opts.IncreaseParallelism(2) // pruning and saving happen in parallel
	opts.SetDbLogDir(logDir)
	opts.EnableStatistics()

	return opts
}

func newDeadNodesCFOptions() *grocksdb.Options {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))
	opts := grocksdb.NewDefaultOptions()
	opts.SetKeepLogFileNum(5)
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetCompression(PNodeDBCompression)

	opts.SetMaxBackgroundJobs(4)               // default was 2, double to 4
	opts.SetMaxWriteBufferNumber(4)            // default was 2, double to 4
	opts.SetWriteBufferSize(128 * 1024 * 1024) // default was 64M, double to 128M
	opts.SetMinWriteBufferNumberToMerge(2)     // default was 1, double to 2
	return opts
}

func newDBOptions() *grocksdb.Options {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetCreateIfMissingColumnFamilies(true)
	opts.SetCompression(PNodeDBCompression)
	return opts
}

// NewPNodeDB - create a new PNodeDB
func NewPNodeDB(stateDir, logDir string) (*PNodeDB, error) {

	var (
		defaultCFOpts = newDefaultCFOptions(logDir)
		deadNodesOpts = newDeadNodesCFOptions()

		cfs     = []string{"default", "dead_nodes"}
		cfsOpts = []*grocksdb.Options{defaultCFOpts, deadNodesOpts}
	)

	db, cfhs, err := grocksdb.OpenDbColumnFamilies(newDBOptions(), stateDir, cfs, cfsOpts)
	if err != nil {
		return nil, err
	}

	wo := grocksdb.NewDefaultWriteOptions()
	wo.SetSync(false)

	return &PNodeDB{
		db:           db,
		defaultCFH:   cfhs[0],
		deadNodesCFH: cfhs[1],
		ro:           grocksdb.NewDefaultReadOptions(),
		wo:           wo,
		to:           grocksdb.NewDefaultTransactionOptions(),
		fo:           grocksdb.NewDefaultFlushOptions(),
	}, nil
}

/*GetNode - implement interface */
func (pndb *PNodeDB) GetNode(key Key) (Node, error) {
	data, err := pndb.db.Get(pndb.ro, key)
	if err != nil {
		return nil, err
	}
	defer data.Free()
	buf := data.Data()
	if len(buf) == 0 {
		return nil, ErrNodeNotFound
	}
	return CreateNode(bytes.NewReader(buf))
}

/*PutNode - implement interface */
func (pndb *PNodeDB) PutNode(key Key, node Node) error {
	nd := node.Clone()
	data := nd.Encode()
	if !bytes.Equal(key, nd.GetHashBytes()) {
		logging.Logger.Error("put node key not match",
			zap.String("key", ToHex(key)),
			zap.String("node", ToHex(nd.GetHashBytes())))
	}

	err := pndb.db.Put(pndb.wo, key, data)

	if DebugMPTNode {
		logging.Logger.Debug("MPT - put node to PersistDB",
			zap.String("key", ToHex(key)),
			zap.String("node key", ToHex(nd.GetHashBytes())),
			zap.Int64("Origin", int64(nd.GetOrigin())),
			zap.Int64("Version", int64(nd.GetVersion())),
			zap.Error(err))
	}
	return err
}

func (pndb *PNodeDB) saveDeadNodes(dn *deadNodes, version int64) error {
	d, err := dn.encode()
	if err != nil {
		return err
	}

	return pndb.db.PutCF(pndb.wo, pndb.deadNodesCFH, uint64ToBytes(uint64(version)), d)
}

// RecordDeadNodes records dead nodes with version
func (pndb *PNodeDB) RecordDeadNodes(nodes []Node, version int64) error {
	dn := deadNodes{make(map[string]bool, len(nodes))}
	for _, n := range nodes {
		dn.Nodes[n.GetHash()] = true
	}

	return pndb.saveDeadNodes(&dn, version)
}

func (pndb *PNodeDB) PruneBelowVersion(ctx context.Context, version int64) error {
	const (
		maxPruneNodes = 1000
	)

	type deadNodesRecord struct {
		round     uint64
		nodesKeys []Key
	}

	var (
		ps    = GetPruneStats(ctx)
		count int64

		keys        = make([]Key, 0, maxPruneNodes)
		pruneRounds = make([]uint64, 0, 100)

		deadNodesC = make(chan deadNodesRecord, 1)
	)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		pndb.iteratorDeadNodes(cctx, func(key, value []byte) bool {
			roundNum := bytesToUint64(key)
			if roundNum >= uint64(version) {
				return false // break iteration
			}

			// decode node keys
			dn := deadNodes{}
			err := dn.decode(value)
			if err != nil {
				logging.Logger.Warn("prune state iterator - iterator decode node keys failed",
					zap.Error(err),
					zap.Uint64("round", roundNum))
				return true // continue
			}

			ns := make([]Key, 0, len(dn.Nodes))
			for k := range dn.Nodes {
				kk, err := fromHex(k)
				if err != nil {
					logging.Logger.Warn("prune state - iterator decode key failed",
						zap.Error(err),
						zap.Uint64("round", roundNum))
					return true // continue
				}
				ns = append(ns, kk)
			}

			deadNodesC <- deadNodesRecord{
				round:     roundNum,
				nodesKeys: ns,
			}
			return true
		})
		close(deadNodesC)
	}()

	for {
		select {
		case dn, ok := <-deadNodesC:
			if !ok {
				if len(keys) > 0 {
					if err := pndb.MultiDeleteNode(keys); err != nil {
						return err
					}

					count += int64(len(keys))
					keys = keys[:0]
				}

				if err := pndb.multiDeleteDeadNodes(pruneRounds); err != nil {
					return err
				}

				// all have been processed
				pndb.Flush()

				if ps != nil {
					ps.Deleted = count
				}

				return nil
			}

			pruneRounds = append(pruneRounds, dn.round)
			keys = append(keys, dn.nodesKeys...)
			if len(keys) >= maxPruneNodes {
				// delete nodes
				if err := pndb.MultiDeleteNode(keys); err != nil {
					return err
				}

				count += int64(len(keys))
				keys = keys[:0]
			}
		}
	}
}

func uint64ToBytes(r uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, r)
	return b
}

func bytesToUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

/*MultiDeleteNode - implement interface */
func (pndb *PNodeDB) multiDeleteDeadNodes(rounds []uint64) error {
	wb := grocksdb.NewWriteBatch()
	defer wb.Destroy()
	for _, r := range rounds {
		wb.DeleteCF(pndb.deadNodesCFH, uint64ToBytes(r))
	}
	return pndb.db.Write(pndb.wo, wb)
}

func (pndb *PNodeDB) iteratorDeadNodes(ctx context.Context, handler func(key, value []byte) bool) {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	ro.SetFillCache(false)
	it := pndb.db.NewIteratorCF(ro, pndb.deadNodesCFH)
	defer it.Close()
	for it.SeekToFirst(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return
		default:
			key := it.Key()
			value := it.Value()

			keyData := key.Data()
			valueData := value.Data()
			if !handler(keyData, valueData) {
				key.Free()
				value.Free()
				return
			}

			key.Free()
			value.Free()
		}
	}
}

/*DeleteNode - implement interface */
func (pndb *PNodeDB) DeleteNode(key Key) error {
	err := pndb.db.Delete(pndb.wo, key)
	return err
}

/*MultiGetNode - get multiple nodes */
func (pndb *PNodeDB) MultiGetNode(keys []Key) ([]Node, error) {
	var nodes []Node
	var err error
	for _, key := range keys {
		node, nerr := pndb.GetNode(key)
		if nerr != nil {
			err = nerr
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, err
}

/*MultiPutNode - implement interface */
func (pndb *PNodeDB) MultiPutNode(keys []Key, nodes []Node) error {
	ts := time.Now()
	wb := grocksdb.NewWriteBatch()
	defer wb.Destroy()
	for idx, key := range keys {
		nd := nodes[idx].Clone()
		if !bytes.Equal(key, nd.GetHashBytes()) {
			logging.Logger.Error("put node key not match",
				zap.String("key", ToHex(key)),
				zap.String("node", ToHex(nd.GetHashBytes())))
		}

		nv := nd.Encode()
		wb.Put(key, nv)
		if DebugMPTNode {
			logging.Logger.Debug("MPT - put node to PersistDB, multiple",
				zap.String("key", ToHex(key)),
				zap.String("node", ToHex(nd.GetHashBytes())),
				zap.Int64("Origin", int64(nodes[idx].GetOrigin())),
				zap.Int64("Version", int64(nodes[idx].GetVersion())))
		}
	}
	err := pndb.db.Write(pndb.wo, wb)
	if err != nil {
		logging.Logger.Error("pnode save nodes failed",
			zap.Int64("round", pndb.version),
			zap.Any("duration", ts),
			zap.Error(err))
	}
	return err
}

/*MultiDeleteNode - implement interface */
func (pndb *PNodeDB) MultiDeleteNode(keys []Key) error {
	// wb := grocksdb.NewWriteBatch()
	// defer wb.Destroy()
	// for _, key := range keys {
	// 	wb.Delete(key)
	// }
	for _, k := range keys {
		if err := pndb.DeleteNode(k); err != nil {
			return err
		}
	}
	// return pndb.db.Write(pndb.wo, wb)
	return nil
}

/*Iterate - implement interface */
func (pndb *PNodeDB) Iterate(ctx context.Context, handler NodeDBIteratorHandler) error {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	ro.SetFillCache(false)
	it := pndb.db.NewIterator(ro)
	defer it.Close()
	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()
		kdata := key.Data()
		if bytes.Equal(kdata, deadNodesKey) {
			continue
		}
		vdata := value.Data()
		node, err := CreateNode(bytes.NewReader(vdata))
		if err != nil {
			key.Free()
			value.Free()
			logging.Logger.Error("iterate - create node", zap.String("key", ToHex(kdata)), zap.Error(err))
			continue
		}
		err = handler(ctx, kdata, node)
		if err != nil {
			key.Free()
			value.Free()
			logging.Logger.Error("iterate - create node handler error", zap.String("key", ToHex(kdata)), zap.Any("data", vdata), zap.Error(err))
			return err
		}
		key.Free()
		value.Free()
	}
	return nil
}

/*Flush - flush the db */
func (pndb *PNodeDB) Flush() {
	pndb.db.Flush(pndb.fo)
}

/*Size - count number of keys in the db */
func (pndb *PNodeDB) Size(ctx context.Context) int64 {
	var count int64
	handler := func(ctx context.Context, key Key, node Node) error {
		count++
		return nil
	}
	err := pndb.Iterate(ctx, handler)
	if err != nil {
		logging.Logger.Error("count", zap.Error(err))
		return -1
	}
	return count
}

// Close closes the rocksdb
func (pndb *PNodeDB) Close() {
	pndb.defaultCFH.Destroy()
	pndb.deadNodesCFH.Destroy()
	pndb.db.Close()
}
