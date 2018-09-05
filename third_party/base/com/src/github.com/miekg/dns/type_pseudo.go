package dns

// dns 保留的 type，这里用于 negative cache。
const (
	TypeNoDomain = 0xFFF0
	TypeNoData   = 0xFFF1
)

type NoDomain struct {
	Hdr RR_Header
}

func (rr *NoDomain) Header() *RR_Header { return &rr.Hdr }
func (rr *NoDomain) Copy() RR           { return &NoDomain{*rr.Hdr.copyHeader()} }
func (rr *NoDomain) String() string     { return rr.Hdr.String() }
func (rr *NoDomain) len() int           { return rr.Hdr.len() }

type NoData struct {
	Hdr  RR_Header
	Type uint16
}

func (rr *NoData) Header() *RR_Header { return &rr.Hdr }
func (rr *NoData) Copy() RR           { return &NoData{*rr.Hdr.copyHeader(), rr.Type} }
func (rr *NoData) String() string     { return rr.Hdr.String() }
func (rr *NoData) len() int           { return rr.Hdr.len() }
