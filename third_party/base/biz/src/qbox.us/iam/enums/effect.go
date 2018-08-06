package enums

type Effect string

const (
	EffectUnknown Effect = "Unknown"
	EffectAllow   Effect = "Allow"
	EffectDeny    Effect = "Deny"
)

func MakeEffect(effect string) Effect {
	switch effect {
	case "Allow":
		return EffectAllow
	case "Deny":
		return EffectDeny
	}
	return EffectUnknown
}

func (e Effect) String() string {
	switch e {
	case EffectAllow:
		return "Allow"
	case EffectDeny:
		return "Deny"
	}
	return "Unknown"
}

func (e Effect) IsAllow() bool {
	return e == EffectAllow
}

func (e Effect) IsDeny() bool {
	return e == EffectDeny
}

func (e Effect) IsUnknown() bool {
	return e == EffectUnknown
}
