package gaeaadmin

import (
	"net/url"
	"strconv"
	"time"
)

type FinancialRelationType int

const (
	FinancialRelationTypeMeasure FinancialRelationType = iota + 1 // 合账: 合并使用量类型
	FinancialRelationTypeBilling                                  // 合账: 合并账单类型
	FinancialRelationTypeNone                                     // 未开启合账
)

type FinancialRelation struct {
	Id        string                `json:"id"`
	Uid       uint32                `json:"uid"`
	ParentUid uint32                `json:"parent_uid"`
	Memo      string                `json:"memo"`
	CreatorId string                `json:"creator_id"`
	Type      FinancialRelationType `json:"type"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updaetd_at"`
}

func (d FinancialRelation) IsParent() bool {
	return d.ParentUid == 0
}

func (d FinancialRelation) IsChild() bool {
	return d.ParentUid > 0
}

type FinancialRelationListParams struct {
	Uid       *uint32                `param:"uid"`
	ParentUid *uint32                `param:"parent_uid"`
	Type      *FinancialRelationType `param:"type"`
	CreatorId *string                `param:"creator_id"`
	Prev      *uint32                `param:"prev"`
	Next      *uint32                `param:"next"`
	PageSize  int                    `param:"page_size"`
}

type FinancialRelationListChildrenParams struct {
	Uid      *uint32 `param:"uid"`
	Prev     *uint32 `param:"prev"`
	Next     *uint32 `param:"next"`
	PageSize int     `param:"page_size"`
}

type FinancialRelationCreateParams struct {
	Uid       uint32                `json:"uid"`
	ParentUid uint32                `json:"parent_uid"`
	Type      FinancialRelationType `json:"type"`
	Memo      string                `json:"memo"`
}

type FinancialRelationUpdateParams struct {
	ParentUid *uint32                `json:"parent_uid"`
	Type      *FinancialRelationType `json:"type"`
	Memo      *string                `json:"memo"`
}

type FinancialRelationList struct {
	List []FinancialRelation `json:"list"`
	Prev uint32              `json:"prev"`
	Next uint32              `json:"next"`
}

func (p *FinancialRelationListParams) Values() (value url.Values) {
	value = make(url.Values)

	if p.Uid != nil {
		value.Set("uid", strconv.FormatUint(uint64(*p.Uid), 10))
	}

	if p.ParentUid != nil {
		value.Set("parent_uid", strconv.FormatUint(uint64(*p.ParentUid), 10))
	}

	if p.Type != nil {
		value.Set("type", strconv.Itoa(int(*p.Type)))
	}

	if p.CreatorId != nil {
		value.Set("creator_id", *p.CreatorId)
	}

	if p.Prev != nil {
		value.Set("prev", strconv.FormatUint(uint64(*p.Prev), 10))
	}

	if p.Next != nil {
		value.Set("next", strconv.FormatUint(uint64(*p.Next), 10))
	}

	value.Set("page_size", strconv.Itoa(p.PageSize))

	return
}

func (p *FinancialRelationListChildrenParams) Values() (value url.Values) {
	value = make(url.Values)

	if p.Uid != nil {
		value.Set("uid", strconv.FormatUint(uint64(*p.Uid), 10))
	}

	if p.Prev != nil {
		value.Set("prev", strconv.FormatUint(uint64(*p.Prev), 10))
	}

	if p.Next != nil {
		value.Set("next", strconv.FormatUint(uint64(*p.Next), 10))
	}

	value.Set("page_size", strconv.Itoa(p.PageSize))

	return
}
