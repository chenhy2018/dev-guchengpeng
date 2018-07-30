package enums

type FreezeSource int

const (
	FreezeSourceUnknown FreezeSource = iota
	FreezeSourcePortalIo
	FreezeSourcePayment
	FreezeSourceKodo
	FreezeSourceFusion
	FreezeSourcePili
	FreezeSourceSofa
	FreezeSourcePortal
)
