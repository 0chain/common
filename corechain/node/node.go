package node

/*Node - a struct holding the node information */
type Node struct {
	client.Client  `yaml:",inline"`
	N2NHost        string        `json:"n2n_host" yaml:"n2n_ip"`
	Host           string        `json:"host" yaml:"public_ip"`
	Port           int           `json:"port" yaml:"port"`
	Path           string        `json:"path" yaml:"path"`
	Type           NodeType      `json:"type" yaml:"-"`
	Description    string        `json:"description" yaml:"description"`
	SetIndex       int           `json:"set_index" yaml:"set_index"`
	Status         int           `json:"status" yaml:"-"`
	InPrevMB       bool          `json:"in_prev_mb" yaml:"-"`
	LastActiveTime time.Time     `json:"-" msgpack:"-" msg:"-" yaml:"-"`
	ErrorCount     int64         `json:"-" msgpack:"-" msg:"-" yaml:"-"`
	CommChannel    chan struct{} `json:"-" msgpack:"-" msg:"-" yaml:"-"`
	//These are approximiate as we are not going to lock to update
	sent       int64 `json:"-" msgpack:"-" msg:"-" yaml:"-"` // messages sent to this node
	sendErrors int64 `json:"-" msgpack:"-" msg:"-" yaml:"-"` // failed message sent to this node
	received   int64 `json:"-" msgpack:"-" msg:"-" yaml:"-"` // messages received from this node

	TimersByURI map[string]metrics.Timer     `json:"-" msgpack:"-" msg:"-" yaml:"-"`
	SizeByURI   map[string]metrics.Histogram `json:"-" msgpack:"-" msg:"-" yaml:"-"`

	largeMessageSendTime uint64 `yaml:"-"`
	smallMessageSendTime uint64 `yaml:"-"`

	LargeMessagePullServeTime float64 `json:"-" msgpack:"-" msg:"-" yaml:"-"`
	SmallMessagePullServeTime float64 `json:"-" msgpack:"-" msg:"-" yaml:"-"`

	mutex sync.RWMutex `json:"-" msgpack:"-" msg:"-" yaml:"-"`

	ProtocolStats interface{} `json:"-" msgpack:"-" msg:"-" yaml:"-"`

	idBytes []byte `yaml:"-"`

	Info Info `json:"info"  yaml:"-"`
}
