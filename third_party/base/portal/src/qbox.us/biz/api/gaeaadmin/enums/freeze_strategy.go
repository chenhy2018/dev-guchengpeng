package enums

type FreezeStrategy int

const (
	FreezeStrategyNotSet FreezeStrategy = iota
	FreezeStrategyA
	FreezeStrategyB
	FreezeStrategyC
)

func (c FreezeStrategy) Humanize() string {
	switch c {
	case FreezeStrategyNotSet:
		return "未设置"
	case FreezeStrategyA:
		return "A"
	case FreezeStrategyB:
		return "B"
	case FreezeStrategyC:
		return "C"
	}
	return "未知"
}
