package account

import (
	"fmt"
)

const (
	FamilyNormal FamilyType = iota + 1
	FamilyParent
	FamilyChild
)

type FamilyType int

func (f FamilyType) Valid() bool {
	switch f {
	case FamilyNormal, FamilyParent, FamilyChild:
		return true
	}
	return false
}

func (f FamilyType) Humanize() string {
	switch f {
	case FamilyNormal:
		return "普通账户"
	case FamilyParent:
		return "父账户"
	case FamilyChild:
		return "子账户"
	}
	return fmt.Sprintf("未定义的家族类型: %d", f)
}

func (f FamilyType) IsParent() bool {
	return f == FamilyParent
}

func (f FamilyType) IsChild() bool {
	return f == FamilyChild
}

func (f FamilyType) IsNormal() bool {
	return f == FamilyNormal
}
