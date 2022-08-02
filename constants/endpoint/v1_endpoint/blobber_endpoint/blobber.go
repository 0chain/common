package blobber_endpoint

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint"
)

const (
	fileResource        string = "/file"
	connectionResource  string = "/connection"
	directoryResource   string = "/dir"
	allocationResource  string = "/allocation"
	getStatsResource    string = "/getstats"
	marketplaceResource string = "/marketplace"
	writeMarkerResource string = "/writemarker"
	hashNodeResource    string = "/hashnode"
	debugResource              = v1_endpoint.InternalEndpoint + "debug"
	configResource             = v1_endpoint.InternalEndpoint + "config"
	statsResource              = v1_endpoint.InternalEndpoint + "stats"
	jsonStstsResource          = v1_endpoint.InternalEndpoint + "statsJSON"
	cleanUpDiskResource        = v1_endpoint.InternalEndpoint + "cleanupdisk"
)

var (
	Debug       = endpoint.New(debugResource)       // /_debug
	Config      = endpoint.New(configResource)      // /_config
	Stats       = endpoint.New(statsResource)       // /_stats
	JsonStats   = endpoint.New(jsonStstsResource)   // /_statsJSON
	CleanupDisk = endpoint.New(cleanUpDiskResource) // /_cleanupdisk
	GetStats    = endpoint.New(getStatsResource)    // /getstats
)

var (
	File              = endpoint.New(v1_endpoint.ApiVersion + fileResource)                 // /v1/file
	FileUpload        = endpoint.JoinWithPathVariable(File, "/upload", "allocation")        // /v1/file/upload/{allocation}
	FileDownload      = endpoint.JoinWithPathVariable(File, "/download", "allocation")      // /v1/file/download/{allocation}
	FileRename        = endpoint.JoinWithPathVariable(File, "/rename", "allocation")        // /v1/file/download/{rename}
	FileCopy          = endpoint.JoinWithPathVariable(File, "/copy", "allocation")          // /v1/file/copy/{rename}
	FileCollaborator  = endpoint.JoinWithPathVariable(File, "/collaborator", "allocation")  // /v1/file/collaborator/{rename}
	FileMeta          = endpoint.JoinWithPathVariable(File, "/meta", "allocation")          // /v1/file/meta/{rename}
	FileStats         = endpoint.JoinWithPathVariable(File, "/stats", "allocation")         // /v1/file/stats/{rename}
	FileList          = endpoint.JoinWithPathVariable(File, "/list", "allocation")          // /v1/file/list/{rename}
	FileObjectPath    = endpoint.JoinWithPathVariable(File, "/objectpath", "allocation")    // /v1/file/objectpath/{rename}
	FileReferencePath = endpoint.JoinWithPathVariable(File, "/referencepath", "allocation") // /v1/file/referencepath/{rename}
	FileObjectTree    = endpoint.JoinWithPathVariable(File, "/objecttree", "allocation")    // /v1/file/objecttree/{rename}
	FileRefs          = endpoint.JoinWithPathVariable(File, "/refs", "allocation")          // /v1/file/refs/{rename}
	FileCommitMetaTxn = endpoint.JoinWithPathVariable(File, "/commitmetatxn", "allocation") // /v1/file/commitmetatxn/{rename}
)

var (
	Marketplace          = endpoint.New(v1_endpoint.ApiVersion + marketplaceResource)             // /v1/marketplace
	MarketplaceShareInfo = endpoint.JoinWithPathVariable(Marketplace, "/shareinfo", "allocation") // /v1/marketplace/shareinfo/{allocation}
)

var (
	WriteMarker               = endpoint.New(v1_endpoint.ApiVersion + writeMarkerResource)        // /v1/writemarker
	WriteMarkerLock           = endpoint.JoinWithPathVariable(WriteMarker, "/lock", "allocation") // /v1/writemarker/lock/{allocation}
	WriteMarkerLockConnection = endpoint.JoinWithPathVariable(WriteMarkerLock, "", "connection")  // /v1/writemarker/lock/{allocation}/{connection}
)

var (
	Hashnode     = endpoint.New(v1_endpoint.ApiVersion + hashNodeResource)       // /v1/hashnode
	HashnodeRoot = endpoint.JoinWithPathVariable(Hashnode, "root", "allocation") // /v1/hashnode/root/{allocation}
)

var (
	Dir = endpoint.NewWithPathVariable(v1_endpoint.ApiVersion+directoryResource, "allocation") // /v1/dir/{allocation}
)

var (
	Allocation = endpoint.New(allocationResource) // /allocation
)
var (
	ConnectionCommit = endpoint.NewWithPathVariable(v1_endpoint.ApiVersion+connectionResource+"/commit", "allocation") // /v1/connection/commit/{allocation}
)
