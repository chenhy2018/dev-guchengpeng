package enums

import (
	"fmt"
)

type FinancialRelationType int

const (
	FinancialRelationTypeMeasure FinancialRelationType = 1
	FinancialRelationTypeBilling FinancialRelationType = 2
)

func (f FinancialRelationType) IsMeasure() bool {
	return f == FinancialRelationTypeMeasure
}

func (f FinancialRelationType) IsBilling() bool {
	return f == FinancialRelationTypeBilling
}

func (f FinancialRelationType) IsValid() bool {
	switch f {
	case FinancialRelationTypeMeasure, FinancialRelationTypeBilling:
		return true
	}
	return false
}

func (f FinancialRelationType) Humanize() string {
	switch f {
	case FinancialRelationTypeMeasure:
		return "合并使用量"
	case FinancialRelationTypeBilling:
		return "合并费用"
	default:
		return fmt.Sprintf("未设置合并类型")
	}
}
