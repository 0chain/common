package chain_endpoint

import "github.com/0chain/common/constants/endpoint/v1_endpoint"
import "github.com/0chain/common/constants/endpoint"

const (
	blockResource                          string = "/block"
	chainResource                          string = "/chain"
	configResource                         string = "/config"
	smartContractStateResource             string = "/scstate"
	smartContractStatsResource             string = "/scstats"
	debugResource                          string = "/debug"
	nodeToNodeInternalFunction                    = v1_endpoint.InternalEndpoint + "n2n"
	anyServiceToAnyServiceInternalFunction        = v1_endpoint.InternalEndpoint + "x2x"
	nodeHost                                      = v1_endpoint.InternalEndpoint + "nh"
)

var (
	HashFunction               = endpoint.New(v1_endpoint.InternalEndpoint + "hash")                 // /_hash
	SignFunction               = endpoint.New(v1_endpoint.InternalEndpoint + "sign")                 // /_sign
	ChainStatsFunction         = endpoint.New(v1_endpoint.InternalEndpoint + "chain_stats")          // /_chain_stats
	SmartContractStatsFunction = endpoint.New(v1_endpoint.InternalEndpoint + "smart_contract_stats") // /_smart_contract_stats
)

var (
	ListNodes      = endpoint.New(nodeHost + "/list")           // /_nh/list
	WhoAmI         = endpoint.New(nodeHost + "/whoami")         // /_nh/whoami
	Status         = endpoint.New(nodeHost + "/status")         // /_nh/status
	GetPoolMembers = endpoint.New(nodeHost + "/getpoolmembers") // /_nh/getpoolmembers
	ListMiners     = endpoint.Join(ListNodes, "/m")             // /_nh/list/m
	ListSharders   = endpoint.Join(ListNodes, "/s")             // /_nh/list/s
)

var (
	Diagnostics               = endpoint.New(v1_endpoint.InternalEndpoint + "diagnostics")                                    // /_diagnostics
	DiagnosticsInfoJson       = endpoint.New(v1_endpoint.InternalEndpoint + "/diagnostics" + v1_endpoint.GetAction + "/info") // /v1/diagnostics/info
	DiagnosticsInfo           = endpoint.Join(Diagnostics, "/info")                                                           // /_diagnostics/info
	WalletStatsDiagnostics    = endpoint.Join(Diagnostics, "/wallet_stats")                                                   // /_diagnostics/wallet_stats
	CurrentMbNodesDiagnostics = endpoint.Join(Diagnostics, "/current_mb_nodes")                                               // /_diagnostics/current_mb_nodes
	DkgProcessDiagnostics     = endpoint.Join(Diagnostics, "/dkg_process")                                                    // /_diagnostics/dkg_process
	RoundInfoDiagnostics      = endpoint.Join(Diagnostics, "/round_info")                                                     // /_diagnostics/round_info
	StateDumpDiagnostics      = endpoint.Join(Diagnostics, "/state_dump")                                                     // /_diagnostics/state_dump
	MinerStatsDiagnostics     = endpoint.Join(Diagnostics, "/miner_stats")                                                    // /_diagnostics/miner_stats
	DiagnosticsLogs           = endpoint.Join(Diagnostics, "/logs")                                                           // /_diagnostics/logs
	DiagnosticsNodeToNodeLogs = endpoint.Join(Diagnostics, "/n2n_logs")                                                       // /_diagnostics/n2n_logs
	DiagnosticsNodeToNodeInfo = endpoint.Join(Diagnostics, "/n2n/info")                                                       // /_diagnostics/n2n/info
	DiagnosticsMemoryLogs     = endpoint.Join(Diagnostics, "/mem_logs")                                                       // /_diagnostics/mem_logs
	BlockChainDiagnostics     = endpoint.Join(Diagnostics, "/block_chain")                                                    // /_diagnostics/block_chain
)

var (
	Debug = endpoint.New(v1_endpoint.ApiVersion + debugResource + "/pprof")
)

var (
	GetConfig       = endpoint.New(v1_endpoint.ApiVersion + configResource + v1_endpoint.GetAction) // /v1/config/get
	UpdateConfig    = endpoint.Join(GetConfig, "/update")                                           // /v1/config/update
	UpdateAllConfig = endpoint.Join(GetConfig, "/update_all")                                       // /v1/config/update
)

var (
	GetSmartContractState = endpoint.New(v1_endpoint.ApiVersion + smartContractStateResource + v1_endpoint.GetAction) // /v1/scstate/get
	GetSmartContractStats = endpoint.New(v1_endpoint.ApiVersion + smartContractStatsResource)                         // /v1/scstats
)

var (
	GetBlock                            = endpoint.New(v1_endpoint.ApiVersion + blockResource + v1_endpoint.GetAction) // /v1/block/get
	GetBlockStateChange                 = endpoint.New(v1_endpoint.ApiVersion + blockResource + "/state_change")       // /v1/block/state_change
	GetLatestFinalizedBlock             = endpoint.Join(GetBlock, "/latest_finalized")                                 // /v1/block/get/latest_finalized
	GetLatestFinalizedTicket            = endpoint.Join(GetBlock, "/latest_finalized_ticket")                          // /v1/block/get/latest_finalized_ticket
	GetLatestFinalizedMagicBlock        = endpoint.Join(GetBlock, "/latest_finalized_magic_block")                     // /v1/block/get/latest_finalized_magic_block
	GetLatestFinalizedMagicBlockSummary = endpoint.Join(GetBlock, "/latest_finalized_magic_block_summary")             // /v1/block/get/latest_finalized_magic_block_summary
	GetRecentFinalizedBlock             = endpoint.Join(GetBlock, "/recent_finalized")                                 // /v1/block/get/recent_finalized
	GetBlockFeeStats                    = endpoint.Join(GetBlock, "/fee_stats")                                        // /v1/block/get/fee_stats
)

var (
	GetChain      = endpoint.New(v1_endpoint.ApiVersion + chainResource + v1_endpoint.GetAction) // /v1/chain/get
	GetChainStats = endpoint.Join(GetChain, "/stats")                                            // /v1/chain/get/stats
	PutChain      = endpoint.New(v1_endpoint.ApiVersion + chainResource + v1_endpoint.PutAction) // /v1/chain/put
)

var (
	NodeToNodePostEntity = endpoint.New(v1_endpoint.ApiVersion + nodeToNodeInternalFunction + "/entity" + v1_endpoint.PostAction)     // /v1/_n2n/entity/post
	NodeToNodeGetEntity  = endpoint.New(v1_endpoint.ApiVersion + nodeToNodeInternalFunction + "/entity_pull" + v1_endpoint.GetAction) // /v1/_n2n/entity_pull/get
)

var (
	AnyServiceToAnyServiceGetBlockStateChange = endpoint.New(v1_endpoint.ApiVersion + anyServiceToAnyServiceInternalFunction + "/block/state_change" + v1_endpoint.GetAction) // /v1/_x2x/block/state_change
	AnyServiceToAnyServiceGetNodes            = endpoint.New(v1_endpoint.ApiVersion + anyServiceToAnyServiceInternalFunction + "/state/get_nodes")                            // /v1/_x2x/state/get_nodes
)
