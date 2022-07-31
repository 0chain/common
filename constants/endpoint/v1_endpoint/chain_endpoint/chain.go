package chain_endpoint

import "github.com/0chain/common/constants/endpoint/v1_endpoint"
import "github.com/0chain/common/constants/endpoint"

const (
	BlockResource                          string = "/block"
	ChainResource                          string = "/chain"
	SmartContractState                     string = "/scstate"
	SmartContractStats                     string = "/scstats"
	NodeToNodeInternalFunction                    = v1_endpoint.InternalEndpoint + "n2n"
	AnyServiceToAnyServiceInternalFunction        = v1_endpoint.InternalEndpoint + "x2x"
)

var (
	HashFunction               = endpoint.New(v1_endpoint.InternalEndpoint + "hash")                 // /_hash
	SignFunction               = endpoint.New(v1_endpoint.InternalEndpoint + "sign")                 // /_sign
	ChainStatsFunction         = endpoint.New(v1_endpoint.InternalEndpoint + "chain_stats")          // /_chain_stats
	SmartContractStatsFunction = endpoint.New(v1_endpoint.InternalEndpoint + "smart_contract_stats") // /_smart_contract_stats
)

var (
	GetSmartContractState = endpoint.New(v1_endpoint.ApiVersion + SmartContractState + v1_endpoint.GetAction) // /v1/scstate/get
	GetSmartContractStats = endpoint.New(v1_endpoint.ApiVersion + SmartContractStats)                         // /v1/scstats
)

var (
	GetBlock                            = endpoint.New(v1_endpoint.ApiVersion + BlockResource + v1_endpoint.GetAction) // /v1/block/get
	GetBlockStateChange                 = endpoint.New(v1_endpoint.ApiVersion + BlockResource + "/state_change")       // /v1/block/state_change
	GetLatestFinalizedBlock             = endpoint.Join(GetBlock, "/latest_finalized")                                 // /v1/block/get/latest_finalized
	GetLatestFinalizedTicket            = endpoint.Join(GetBlock, "/latest_finalized_ticket")                          // /v1/block/get/latest_finalized_ticket
	GetLatestFinalizedMagicBlock        = endpoint.Join(GetBlock, "/latest_finalized_magic_block")                     // /v1/block/get/latest_finalized_magic_block
	GetLatestFinalizedMagicBlockSummary = endpoint.Join(GetBlock, "/latest_finalized_magic_block_summary")             // /v1/block/get/latest_finalized_magic_block_summary
	GetRecentFinalizedBlock             = endpoint.Join(GetBlock, "/recent_finalized")                                 // /v1/block/get/recent_finalized
	GetBlockFeeStats                    = endpoint.Join(GetBlock, "/fee_stats")                                        // /v1/block/get/fee_stats
)

var (
	GetChain      = endpoint.New(v1_endpoint.ApiVersion + ChainResource + v1_endpoint.GetAction) // /v1/chain/get
	GetChainStats = endpoint.Join(GetChain, "/stats")                                            // /v1/chain/get/stats
	PutChain      = endpoint.New(v1_endpoint.ApiVersion + ChainResource + v1_endpoint.PutAction) // /v1/chain/put
)

var (
	NodeToNodePostEntity = endpoint.New(v1_endpoint.ApiVersion + NodeToNodeInternalFunction + "/entity" + v1_endpoint.PostAction) // /v1/_n2n/entity/post
)

var (
	AnyServiceToAnyServiceGetBlockStateChange = endpoint.New(v1_endpoint.ApiVersion + AnyServiceToAnyServiceInternalFunction + "/block/state_change" + v1_endpoint.GetAction) // /v1/_x2x/block/state_change
	AnyServiceToAnyServiceGetNodes            = endpoint.New(v1_endpoint.ApiVersion + AnyServiceToAnyServiceInternalFunction + "/state/get_nodes")                            // /v1/_x2x/state/get_nodes
)
