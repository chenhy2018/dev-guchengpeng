package enums

type FreezeTicketStatus int

const (
	FreezeTicketStatusWaiting FreezeTicketStatus = iota + 1
	FreezeTicketStatusRunning
	FreezeTicketStatusSuccess
	FreezeTicketStatusFailed
	FreezeTicketStatusCanceled
)

func (f FreezeTicketStatus) IsWaiting() bool {
	return f == FreezeTicketStatusWaiting
}

func (f FreezeTicketStatus) IsRunning() bool {
	return f == FreezeTicketStatusRunning
}

func (f FreezeTicketStatus) IsSuccess() bool {
	return f == FreezeTicketStatusSuccess
}

func (f FreezeTicketStatus) IsFailed() bool {
	return f == FreezeTicketStatusFailed
}

func (f FreezeTicketStatus) IsCanceled() bool {
	return f == FreezeTicketStatusCanceled
}
