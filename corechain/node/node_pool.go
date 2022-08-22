package node

import (
	"sync"
)
/*Pool - a pool of nodes used for the same purpose */
type Pool struct {
	Type NodeType `json:"type"`

	// ---------------------------------------------
	mmx      sync.RWMutex     `json:"-" msgpack:"-" msg:"-"`
	Nodes    []*Node          `json:"-" msgpack:"-" msg:"-"`
	NodesMap map[string]*Node `json:"nodes"`
	// ---------------------------------------------

	medianNetworkTime uint64 `msg:"-"` // float64
}
