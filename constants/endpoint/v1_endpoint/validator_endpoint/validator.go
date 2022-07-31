package validator_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	storageResource string = "/storage"
	debugResource   string = "/debug"
)

var (
	Debug = endpoint.New(debugResource) // /debug
)

var (
	Challenge    = endpoint.New(v1_endpoint.ApiVersion + storageResource + "/challenge") // /v1/storage/challenge
	NewChallenge = endpoint.Join(Challenge, "/new")                                      // /v1/storage/challenge/new
)
