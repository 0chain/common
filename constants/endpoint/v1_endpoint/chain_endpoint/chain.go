package chain_endpoint

import "github.com/0chain/common/constants/endpoint/v1_endpoint"

const (
	BlockResource      string = "/block"
	ChainResource      string = "/chain"
	SmartContractState string = "/scstate"
	SmartContractStats string = "/scstats"
)

const (
	HashFunction                           = v1_endpoint.InternalEndpoint + "hash"        // /_hash
	SignFunction                           = v1_endpoint.InternalEndpoint + "sign"        // /_sign
	ChainStatsFunction                     = v1_endpoint.InternalEndpoint + "chain_stats" // /_chain_stats
	NodeToNodeInternalFunction             = v1_endpoint.InternalEndpoint + "n2n"
	AnyServiceToAnyServiceInternalFunction = v1_endpoint.InternalEndpoint + "x2x"
)

const (
	GetSmartContractState = v1_endpoint.ApiVersion + SmartContractState + v1_endpoint.GetAction // /v1/scstate/get
	GetSmartContractStats = v1_endpoint.ApiVersion + SmartContractStats + v1_endpoint.GetAction // /v1/scstats/get
)

const (
	GetBlock                            = v1_endpoint.ApiVersion + BlockResource + v1_endpoint.GetAction // /v1/block/get
	GetBlockStateChange                 = v1_endpoint.ApiVersion + BlockResource + "/state_change"       // /v1/block/state_change
	GetLatestFinalizedBlock             = GetBlock + "/latest_finalized"                                 // /v1/block/get/latest_finalized
	GetLatestFinalizedTicket            = GetBlock + "/latest_finalized_ticket"                          // /v1/block/get/latest_finalized_ticket
	GetLatestFinalizedMagicBlock        = GetBlock + "/latest_finalized_magic_block"                     // /v1/block/get/latest_finalized_magic_block
	GetLatestFinalizedMagicBlockSummary = GetBlock + "/latest_finalized_magic_block_summary"             // /v1/block/get/latest_finalized_magic_block_summary
	GetRecentFinalizedBlock             = GetBlock + "/recent_finalized"                                 // /v1/block/get/recent_finalized
	GetBlockFeeStats                    = GetBlock + "/fee_stats"                                        // /v1/block/get/fee_stats
)

const (
	GetChain      = v1_endpoint.ApiVersion + ChainResource + v1_endpoint.GetAction // /v1/chain/get
	GetChainStats = GetChain + "/stats"                                            // /v1/chain/get/stats
	PutChain      = v1_endpoint.ApiVersion + ChainResource + v1_endpoint.PutAction // /v1/chain/put
)

const (
	NodeToNodePostEntity = v1_endpoint.ApiVersion + NodeToNodeInternalFunction + "/entity" + v1_endpoint.PostAction // /v1/_n2n/entity/post
)

const (
	AnyServiceToAnyServiceGetBlockStateChange = v1_endpoint.ApiVersion + AnyServiceToAnyServiceInternalFunction + "/block/state_change" + v1_endpoint.GetAction // /v1/_x2x/block/state_change
	AnyServiceToAnyServiceGetNodes            = v1_endpoint.ApiVersion + AnyServiceToAnyServiceInternalFunction + "/state/get_nodes"                            // /v1/_x2x/state/get_nodes
)
