package types

import (
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/runtime/filter"
)

type ListHookReleasesOption struct {
	BizID  uint32             `json:"biz_id"`
	HookID uint32             `json:"hook_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// ListHookReleaseDetails defines the response details of requested ListHooksReleaseOption.
type ListHookReleaseDetails struct {
	Count   uint32               `json:"count"`
	Details []*table.HookRelease `json:"details"`
}

// Validate the list release options
func (opt *ListHookReleasesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid hook id id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id, hook_id because it's a required field in the option.
		RuleFields: table.ReleaseColumns.WithoutColumn("biz_id", "hook_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}
