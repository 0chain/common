package sharder_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	SharderResource                     string = "/sharder"
	BlockResource                       string = "/block"
	TransactionResource                 string = "/transaction"
	SmartContractResource               string = "/screst"
	MinerToSharderInternalFunction             = v1_endpoint.InternalEndpoint + "m2s"
	SharderToSharderInternalFunction           = v1_endpoint.InternalEndpoint + "s2s"
	AnyServiceToSharderInternalFunction        = v1_endpoint.InternalEndpoint + "x2s"
)

var (
	GetSharderStats = endpoint.New(v1_endpoint.ApiVersion + SharderResource + v1_endpoint.GetAction + "/stats") // /v1/sharder/get/stats
)

var (
	HealthCheckFunction = endpoint.New(v1_endpoint.InternalEndpoint + "healthcheck") // /_healthcheck
)

var (
	SmartContractFunction = endpoint.New(v1_endpoint.ApiVersion + SmartContractResource) // /v1/screst
)

var (
	GetMagicBlock = endpoint.New(v1_endpoint.ApiVersion + BlockResource + "/magic" + v1_endpoint.GetAction) // /v1/block/magic/get
)

var (
	GetTransaction             = endpoint.New(v1_endpoint.ApiVersion + TransactionResource + v1_endpoint.GetAction) // /v1/transaction/get
	GetTransactionConfirmation = endpoint.Join(GetTransaction, "/confirmation")                                     // /v1/transaction/get/confirmation
)

var (
	MinerToSharderBlock                   = endpoint.New(v1_endpoint.ApiVersion + MinerToSharderInternalFunction + "/block") // /v1/_m2s/block
	MinerToSharderGetFinalizedBlock       = endpoint.Join(MinerToSharderBlock, "/finalized")                                 // /v1/_m2s/block/finalized
	MinerToSharderGetNotarisedBlock       = endpoint.Join(MinerToSharderBlock, "/notarized")                                 // /v1/_m2s/block/notarized
	MinerToSharderKickNotarisedBlock      = endpoint.Join(MinerToSharderGetNotarisedBlock, "/kick")                          // /v1/_m2s/block/notarized/kick
	MinerToSharderGetLatestFinalizedBlock = endpoint.Join(MinerToSharderBlock, "/latest_finalized"+v1_endpoint.GetAction)    // /v1/_m2s/block/latest_finalized/get
)

var (
	SharderToSharderGetRound          = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/round" + v1_endpoint.GetAction)          // /v1/_s2s/round/get
	SharderToSharderGetLatestRound    = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/latest_round" + v1_endpoint.GetAction)   // /v1/_s2s/latest_round/get
	SharderToSharderGetBlock          = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/block" + v1_endpoint.GetAction)          // /v1/_s2s/block/get
	SharderToSharderGetBlockSummary   = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/blocksummary" + v1_endpoint.GetAction)   // /v1/_s2s/blocksummary/get
	SharderToSharderGetBlockSummaries = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/blocksummaries" + v1_endpoint.GetAction) // /v1/_s2s/blocksummaries/get
	SharderToSharderGetRoundSummaries = endpoint.New(v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/roundsummaries" + v1_endpoint.GetAction) // /v1/_s2s/roundsummaries/get
)

var (
	AnyServiceToSharderGetBlock = endpoint.New(v1_endpoint.ApiVersion + AnyServiceToSharderInternalFunction + "/block" + v1_endpoint.GetAction) // /v1/_x2s/block/get
)
