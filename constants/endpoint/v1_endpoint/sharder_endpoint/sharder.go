package sharder_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	sharderResource                     string = "/sharder"
	blockResource                       string = "/block"
	transactionResource                 string = "/transaction"
	smartContractResource               string = "/screst"
	stateResource                       string = "/state"
	healthCheckResource                 string = "/healthcheck"
	minerToSharderInternalFunction             = v1_endpoint.InternalEndpoint + "m2s"
	sharderToSharderInternalFunction           = v1_endpoint.InternalEndpoint + "s2s"
	anyServiceToSharderInternalFunction        = v1_endpoint.InternalEndpoint + "x2s"
)

var (
	NodesState = endpoint.New(v1_endpoint.ApiVersion + stateResource + "/nodes") // /v1/state/nodes
)

var (
	GetSharderStats = endpoint.New(v1_endpoint.ApiVersion + sharderResource + v1_endpoint.GetAction + "/stats") // /v1/sharder/get/stats
)

var (
	HealthCheckFunction = endpoint.New(v1_endpoint.InternalEndpoint + healthCheckResource) // /_healthcheck
	HealthCheck         = endpoint.New(v1_endpoint.ApiVersion + healthCheckResource)       // /v1/healthcheck
)

var (
	SmartContractFunction = endpoint.New(v1_endpoint.ApiVersion + smartContractResource) // /v1/screst
)

var (
	GetMagicBlock = endpoint.New(v1_endpoint.ApiVersion + blockResource + "/magic" + v1_endpoint.GetAction) // /v1/block/magic/get
)

var (
	GetTransaction             = endpoint.New(v1_endpoint.ApiVersion + transactionResource + v1_endpoint.GetAction) // /v1/transaction/get
	GetTransactionConfirmation = endpoint.Join(GetTransaction, "/confirmation")                                     // /v1/transaction/get/confirmation
)

var (
	MinerToSharderBlock                   = endpoint.New(v1_endpoint.ApiVersion + minerToSharderInternalFunction + "/block") // /v1/_m2s/block
	MinerToSharderGetFinalizedBlock       = endpoint.Join(MinerToSharderBlock, "/finalized")                                 // /v1/_m2s/block/finalized
	MinerToSharderGetNotarisedBlock       = endpoint.Join(MinerToSharderBlock, "/notarized")                                 // /v1/_m2s/block/notarized
	MinerToSharderKickNotarisedBlock      = endpoint.Join(MinerToSharderGetNotarisedBlock, "/kick")                          // /v1/_m2s/block/notarized/kick
	MinerToSharderGetLatestFinalizedBlock = endpoint.Join(MinerToSharderBlock, "/latest_finalized"+v1_endpoint.GetAction)    // /v1/_m2s/block/latest_finalized/get
)

var (
	SharderToSharderGetRound          = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/round" + v1_endpoint.GetAction)          // /v1/_s2s/round/get
	SharderToSharderGetLatestRound    = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/latest_round" + v1_endpoint.GetAction)   // /v1/_s2s/latest_round/get
	SharderToSharderGetBlock          = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/block" + v1_endpoint.GetAction)          // /v1/_s2s/block/get
	SharderToSharderGetBlockSummary   = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/blocksummary" + v1_endpoint.GetAction)   // /v1/_s2s/blocksummary/get
	SharderToSharderGetBlockSummaries = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/blocksummaries" + v1_endpoint.GetAction) // /v1/_s2s/blocksummaries/get
	SharderToSharderGetRoundSummaries = endpoint.New(v1_endpoint.ApiVersion + sharderToSharderInternalFunction + "/roundsummaries" + v1_endpoint.GetAction) // /v1/_s2s/roundsummaries/get
)

var (
	AnyServiceToSharderGetBlock = endpoint.New(v1_endpoint.ApiVersion + anyServiceToSharderInternalFunction + "/block" + v1_endpoint.GetAction) // /v1/_x2s/block/get
)
