package v3

const (
	ResourceGroupTypeInvalid = iota
	ResourceGroupTypeFusion
)

func NewResourceGroupForType(typ int) interface{} {
	switch typ {
	case ResourceGroupTypeFusion:
		return &ResourceGroupFusion{}
	default:
		return nil
	}
}

type ResourceGroupMeta struct {
	Name  string             `json:"name"`
	Desc  string             `json:"desc"`
	Price ModelItemBasePrice `json:"price"`
}

type ResourceGroupFusion struct {
	ResourceGroupMeta
	Domains []string `json:"domains"`
}
