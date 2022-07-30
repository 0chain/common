package v1_endpoint

const (
	ApiVersion       string = "/v1"
	InternalEndpoint string = "/_"
)

const (
	GetAction  string = "/get"
	PutAction  string = "/put"
	PostAction string = "/post"
)

const (
	NodeToNodeInternalFunction             = InternalEndpoint + "n2n"
	AnyServiceToAnyServiceInternalFunction = InternalEndpoint + "x2x"
)

const (
	NodeToNodePostEntity = ApiVersion + NodeToNodeInternalFunction + "/entity" + PostAction
)

const (
	AnyServiceToAnyServiceGetBlockStateChange = ApiVersion + AnyServiceToAnyServiceInternalFunction + "/block/state_change" + GetAction
	AnyServiceToAnyServiceGetNodes            = ApiVersion + AnyServiceToAnyServiceInternalFunction + "/state/get_nodes"
)
