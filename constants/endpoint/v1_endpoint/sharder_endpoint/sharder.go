package sharder_endpoint

import "github.com/0chain/common/constants/endpoint/v1_endpoint"

const (
	SharderResource                     string = "/sharder"
	BlockResource                       string = "/block"
	TransactionResource                 string = "/transaction"
	SmartContractFunction               string = "/screst"
	MinerToSharderInternalFunction             = v1_endpoint.InternalEndpoint + "m2s"
	SharderToSharderInternalFunction           = v1_endpoint.InternalEndpoint + "s2s"
	AnyServiceToSharderInternalFunction        = v1_endpoint.InternalEndpoint + "x2s"
)

const (
	GetSharderStats = v1_endpoint.ApiVersion + SharderResource + v1_endpoint.GetAction + "/stats" // /v1/sharder/get/stats
)

const (
	HealthCheckFunction = v1_endpoint.InternalEndpoint + "healthcheck" // /_healthcheck
)

const (
	GetMagicBlock = v1_endpoint.ApiVersion + BlockResource + "/magic" + v1_endpoint.GetAction // /v1/block/magic/get
)

const (
	GetTransaction             = v1_endpoint.ApiVersion + TransactionResource + v1_endpoint.GetAction // /v1/transaction/get
	GetTransactionConfirmation = GetTransaction + "/confirmation"                                     // /v1/transaction/get/confirmation
)

const (
	MinerToSharderBlock                   = v1_endpoint.ApiVersion + MinerToSharderInternalFunction + "/block" // /v1/_m2s/block
	MinerToSharderGetFinalizedBlock       = MinerToSharderBlock + "/finalized"                                 // /v1/_m2s/block/finalized
	MinerToSharderGetNotarisedBlock       = MinerToSharderBlock + "/notarized"                                 // /v1/_m2s/block/notarized
	MinerToSharderKickNotarisedBlock      = MinerToSharderGetNotarisedBlock + "/kick"                          // /v1/_m2s/block/notarized/kick
	MinerToSharderGetLatestFinalizedBlock = MinerToSharderBlock + "/latest_finalized" + v1_endpoint.GetAction  // /v1/_m2s/block/latest_finalized/get
)

const (
	SharderToSharderGetRound          = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/round" + v1_endpoint.GetAction          // /v1/_s2s/round/get
	SharderToSharderGetLatestRound    = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/latest_round" + v1_endpoint.GetAction   // /v1/_s2s/latest_round/get
	SharderToSharderGetBlock          = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/block" + v1_endpoint.GetAction          // /v1/_s2s/block/get
	SharderToSharderGetBlockSummary   = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/blocksummary" + v1_endpoint.GetAction   // /v1/_s2s/blocksummary/get
	SharderToSharderGetBlockSummaries = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/blocksummaries" + v1_endpoint.GetAction // /v1/_s2s/blocksummaries/get
	SharderToSharderGetRoundSummaries = v1_endpoint.ApiVersion + SharderToSharderInternalFunction + "/roundsummaries" + v1_endpoint.GetAction // /v1/_s2s/roundsummaries/get
)

const (
	AnyServiceToSharderGetBlock = v1_endpoint.ApiVersion + AnyServiceToSharderInternalFunction + "/block" + v1_endpoint.GetAction // /v1/_x2s/block/get
)
