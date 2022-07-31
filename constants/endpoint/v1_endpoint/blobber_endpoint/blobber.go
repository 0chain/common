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
	FileUpload   = endpoint.New(v1_endpoint.ApiVersion + fileResource + "/upload")       // /v1/file/upload
	FileDownload = endpoint.New(v1_endpoint.ApiVersion + fileResource + "/download")     // /v1/file/download
	UploadCommit = endpoint.New(v1_endpoint.ApiVersion + connectionResource + "/commit") // /v1/connection/commit
)
