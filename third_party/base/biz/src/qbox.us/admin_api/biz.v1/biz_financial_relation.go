package biz

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
)

type FinancialRelation struct {
	Uid       uint32                `json:"uid,string"`        // 用户 Uid
	ParentUid uint32                `json:"parent_uid,string"` // 父账户 Uid
	Type      FinancialRelationType `json:"type"`              // 合并类型
}

func (d *FinancialRelation) IsParent() bool {
	return d.ParentUid == 0
}

func (d *FinancialRelation) IsChild() bool {
	return d.ParentUid > 0
}

// 返回 合帐-用户的信息，无合帐财务信息则返回 not found
func (s *BizService) FinancialRelationUserInfo(l rpc.Logger, uid uint32) (res *FinancialRelation, err error) {
	url := fmt.Sprintf("/admin/financial/relation/%d", uid)
	err = s.rpc.GetCall(l, &res, s.host+url)
	return
}

// 返回 合帐-父账户的子账户列表
func (s *BizService) FinancialRelationChildren(l rpc.Logger, uid uint32) (res []*FinancialRelation, err error) {
	url := fmt.Sprintf("/admin/financial/relation/%d/children", uid)
	err = s.rpc.GetCall(l, &res, s.host+url)
	return
}

// 返回 合帐-全部账户列表
func (s *BizService) FinancialRelationList(l rpc.Logger, offset, limit int) (res []*FinancialRelation, err error) {
	url := s.host + "/admin/financial/relation?"
	url += fmt.Sprintf("offset=%d&limit=%d", offset, limit)
	err = s.rpc.GetCall(l, &res, url)
	return
}

// 返回 合帐-全部子账户列表
func (s *BizService) FinancialRelationListChildren(l rpc.Logger, offset, limit int) (res []*FinancialRelation, err error) {
	url := s.host + "/admin/financial/relation/children?"
	url += fmt.Sprintf("offset=%d&limit=%d", offset, limit)
	err = s.rpc.GetCall(l, &res, url)
	return
}
