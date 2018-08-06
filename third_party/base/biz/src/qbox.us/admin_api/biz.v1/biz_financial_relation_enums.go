package biz

type FinancialRelationType int

const (
	FinancialRelationTypeMeasure FinancialRelationType = 1 // 合并使用量
	FinancialRelationTypeBilling FinancialRelationType = 2 // 合并费用
)

func (f FinancialRelationType) IsMeasure() bool {
	return f == FinancialRelationTypeMeasure
}

func (f FinancialRelationType) IsBilling() bool {
	return f == FinancialRelationTypeBilling
}
