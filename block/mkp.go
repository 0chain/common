package block

//go:generate msgp -io=false -tests=false -v

type MPK struct {
	ID  string
	Mpk []string
}

// swagger:model Mpks
type Mpks struct {
	Mpks map[string]*MPK
}
