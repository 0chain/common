package miner_endpoint

import "github.com/0chain/common/constants/endpoint/v1_endpoint"

const (
	MinerResource                     string = "/miner"
	ClientResource                    string = "/client"
	TransactionResource               string = "/transaction"
	AnyServiceToMinerInternalFunction        = v1_endpoint.InternalEndpoint + "x2m"
	MinerToMinerInternalFunction             = v1_endpoint.InternalEndpoint + "m2m"
)

const (
	GetMinerStats = v1_endpoint.ApiVersion + MinerResource + v1_endpoint.GetAction + "/stats" // /v1/miner/get/stats
)

const (
	WalletStatsFunction = v1_endpoint.InternalEndpoint + "diagnostics/wallet_stats" // /_diagnostics/wallet_stats
)

const (
	GetClient        = v1_endpoint.ApiVersion + ClientResource + v1_endpoint.GetAction // /v1/client/get
	PutClient        = v1_endpoint.ApiVersion + ClientResource + v1_endpoint.PutAction // /v1/client/put
	GetClientBalance = GetClient + "/balance"                                          // /v1/client/get/balance
)

const (
	PutTransaction = v1_endpoint.ApiVersion + TransactionResource + v1_endpoint.PutAction // /v1/transaction/put
)

const (
	MinerToMinerRound              = v1_endpoint.ApiVersion + MinerToMinerInternalFunction + "/round"     // /v1/_m2m/round
	MinerToMinerBlock              = v1_endpoint.ApiVersion + MinerToMinerInternalFunction + "/block"     // /v1/_m2m/block
	MinerToMinerChain              = v1_endpoint.ApiVersion + MinerToMinerInternalFunction + "/chain"     // /v1/_m2m/chain
	MinerToMinerDkgShare           = v1_endpoint.ApiVersion + MinerToMinerInternalFunction + "/dkg/share" // /v1/_m2m/dkg/share
	MinerToMinerRoundVRFSender     = MinerToMinerRound + "/vrf_share"                                     // /v1/_m2m/round/vrf_share
	MinerToMinerVerifyBlock        = MinerToMinerBlock + "/verify"                                        // /v1/_m2m/block/verify
	MinerToMinerNotarizedBlock     = MinerToMinerBlock + "/notarized_block"                               // /v1/_m2m/block/notarized_block
	MinerToMinerVerificationTicket = MinerToMinerBlock + "/verification_ticket"                           // /v1/_m2m/block/verification_ticket
	MinerToMinerNotarization       = MinerToMinerBlock + "/notarization"                                  // /v1/_m2m/block/notarization
	MinerToMinerChainStart         = MinerToMinerChain + "/start"                                         // /v1/_m2m/chain/start
)

const (
	AnyServiceToMinerBlock             = v1_endpoint.ApiVersion + AnyServiceToMinerInternalFunction + "/block"                         // /v1/_x2m/block
	AnyServiceToMinerGetNotarizedBlock = AnyServiceToMinerBlock + "/notarized_block" + v1_endpoint.GetAction                           // /v1/_x2m/block/notarized_block/get
	AnyServiceToMinerGetStateChange    = AnyServiceToMinerBlock + "/state_change" + v1_endpoint.GetAction                              // /v1/_x2m/block/state_change/get
	AnyServiceToMinerGetState          = v1_endpoint.ApiVersion + AnyServiceToMinerInternalFunction + "/state" + v1_endpoint.GetAction // /v1/_x2m/state/get
)
