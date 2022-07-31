package blobber_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	fileResource       string = "/file"
	connectionResource string = "/connection"
)

var (
	FileUpload       = endpoint.NewWithPathVariable(v1_endpoint.ApiVersion+fileResource+"/upload", "allocation")       // /v1/file/upload/{allocation}
	FileDownload     = endpoint.NewWithPathVariable(v1_endpoint.ApiVersion+fileResource+"/download", "allocation")     // /v1/file/download/{allocation}
	ConnectionCommit = endpoint.NewWithPathVariable(v1_endpoint.ApiVersion+connectionResource+"/commit", "allocation") // /v1/connection/commit/{allocation}
)
