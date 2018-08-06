package enums

type FreezeOperation int

const (
	FreezeOperationImmidate = iota
	FreezeOperationSchedule
	FreezeOperationExecSchedule
)

type UnfreezeOperation int

const (
	UnfreezeOperationImmidate = iota
	UnfreezeOperationCancelSchedule
)
