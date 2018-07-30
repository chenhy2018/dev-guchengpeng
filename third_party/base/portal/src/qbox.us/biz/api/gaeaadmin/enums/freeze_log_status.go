package enums

type FreezeLogStatus int

const (
	FreezeLogStatusRunning FreezeLogStatus = iota + 1
	FreezeLogStatusSuccess
	FreezeLogStatusFailed
)

func (s FreezeLogStatus) IsRunning() bool {
	return s == FreezeLogStatusRunning
}

func (s FreezeLogStatus) IsSuccess() bool {
	return s == FreezeLogStatusSuccess
}

func (s FreezeLogStatus) IsFailed() bool {
	return s == FreezeLogStatusFailed
}
