package types

import (
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/runtime/filter"
)

type ListCredentialsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list group options
func (opt *ListCredentialsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}
	exprOpt := &filter.ExprOption{
		RuleFields: table.CommitsColumns.WithoutColumn("biz_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}
	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

type ListCredentialDetails struct {
	Count   uint32              `json:"count"`
	Details []*table.Credential `json:"details"`
}
