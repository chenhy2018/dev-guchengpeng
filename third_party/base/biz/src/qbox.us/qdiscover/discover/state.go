package discover

type State string

const (
	StatePending  State = "pending"
	StateEnabled  State = "enabled"
	StateDisabled State = "disabled"
	StateOnline   State = "online"
	StateOffline  State = "offline"
)

func ValidState(s string) (state State, ok bool) {
	state = State(s)
	switch state {
	case StatePending, StateEnabled, StateDisabled, StateOnline, StateOffline:
		return state, true
	}
	return state, false
}
