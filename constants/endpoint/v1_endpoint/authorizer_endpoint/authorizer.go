package authorizer_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	etherResource string = "/ether"
	zChainResource   string = "/0chain"
)

var (
	Ether = endpoint.New(v1_endpoint.ApiVersion + etherResource) // /v1/ether
	GetBurnEtherTicket = endpoint.Join(Ether, "/burnticket" + v1_endpoint.GetAction) // /v1/ether/burnticket/get
)

var (
	Native = endpoint.New(v1_endpoint.ApiVersion + zChainResource) // /v1/0chain
	GetBurnNativeTicket = endpoint.Join(Native, "/burnticket" + v1_endpoint.GetAction) // /v1/0chain/burnticket/get
)