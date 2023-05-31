package types

import (
	"errors"

	"bscp.io/pkg/dal/table"
)

type ListHookReleasesOption struct {
	BizID     uint32    `json:"biz_id"`
	HookID    uint32    `json:"hook_id"`
	Page      *BasePage `json:"page"`
	SearchKey string    `json:"search_key"`
}

// ListHookReleaseDetails defines the response details of requested ListHooksReleaseOption.
type ListHookReleaseDetails struct {
	Count   uint32               `json:"count"`
	Details []*table.HookRelease `json:"details"`
}

// Validate the list release options
func (opt *ListHookReleasesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errors.New("invalid hook id id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

type GetByPubStateOption struct {
	BizID  uint32
	HookID uint32
	State  table.ReleaseStatus
}

func (opt *GetByPubStateOption) Validate() error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errors.New("invalid hook id id, should >= 1")
	}

	if err := opt.State.Validate(); err != nil {
		return err
	}

	return nil
}
