package miner_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	minerResource                     string = "/miner"
	clientResource                    string = "/client"
	transactionResource               string = "/transaction"
	anyServiceToMinerInternalFunction        = v1_endpoint.InternalEndpoint + "x2m"
	minerToMinerInternalFunction             = v1_endpoint.InternalEndpoint + "m2m"
)

var (
	GetMinerStats = endpoint.New(v1_endpoint.ApiVersion + minerResource + v1_endpoint.GetAction + "/stats") // /v1/miner/get/stats
)

var (
	GetClient        = endpoint.New(v1_endpoint.ApiVersion + clientResource + v1_endpoint.GetAction) // /v1/client/get
	PutClient        = endpoint.New(v1_endpoint.ApiVersion + clientResource + v1_endpoint.PutAction) // /v1/client/put
	GetClientBalance = endpoint.Join(GetClient, "/balance")                                          // /v1/client/get/balance
)

var (
	PutTransaction = endpoint.New(v1_endpoint.ApiVersion + transactionResource + v1_endpoint.PutAction) // /v1/transaction/put
)

var (
	MinerToMinerRound              = endpoint.New(v1_endpoint.ApiVersion + minerToMinerInternalFunction + "/round")     // /v1/_m2m/round
	MinerToMinerBlock              = endpoint.New(v1_endpoint.ApiVersion + minerToMinerInternalFunction + "/block")     // /v1/_m2m/block
	MinerToMinerChain              = endpoint.New(v1_endpoint.ApiVersion + minerToMinerInternalFunction + "/chain")     // /v1/_m2m/chain
	MinerToMinerDkgShare           = endpoint.New(v1_endpoint.ApiVersion + minerToMinerInternalFunction + "/dkg/share") // /v1/_m2m/dkg/share
	MinerToMinerRoundVRFSender     = endpoint.Join(MinerToMinerRound, "/vrf_share")                                     // /v1/_m2m/round/vrf_share
	MinerToMinerVerifyBlock        = endpoint.Join(MinerToMinerBlock, "/verify")                                        // /v1/_m2m/block/verify
	MinerToMinerNotarizedBlock     = endpoint.Join(MinerToMinerBlock, "/notarized_block")                               // /v1/_m2m/block/notarized_block
	MinerToMinerVerificationTicket = endpoint.Join(MinerToMinerBlock, "/verification_ticket")                           // /v1/_m2m/block/verification_ticket
	MinerToMinerNotarization       = endpoint.Join(MinerToMinerBlock, "/notarization")                                  // /v1/_m2m/block/notarization
	MinerToMinerChainStart         = endpoint.Join(MinerToMinerChain, "/start")                                         // /v1/_m2m/chain/start
)

var (
	AnyServiceToMinerBlock             = endpoint.New(v1_endpoint.ApiVersion + anyServiceToMinerInternalFunction + "/block")                         // /v1/_x2m/block
	AnyServiceToMinerGetNotarizedBlock = endpoint.Join(AnyServiceToMinerBlock, "/notarized_block"+v1_endpoint.GetAction)                             // /v1/_x2m/block/notarized_block/get
	AnyServiceToMinerGetStateChange    = endpoint.Join(AnyServiceToMinerBlock, "/state_change"+v1_endpoint.GetAction)                                // /v1/_x2m/block/state_change/get
	AnyServiceToMinerGetState          = endpoint.New(v1_endpoint.ApiVersion + anyServiceToMinerInternalFunction + "/state" + v1_endpoint.GetAction) // /v1/_x2m/state/get
)
