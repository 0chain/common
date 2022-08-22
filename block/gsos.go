package block

import (
	"sync"
)

//go:generate msgp -io=false -tests=false -v

// swagger:model GroupSharesOrSigns
type GroupSharesOrSigns struct {
	mutex  sync.RWMutex             `json:"-" msgpack:"-" msg:"-"`
	Shares map[string]*ShareOrSigns `json:"shares"`
}