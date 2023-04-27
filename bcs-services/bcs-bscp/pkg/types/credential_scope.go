package types

import (
	"bscp.io/pkg/dal/table"
)

// ListCredentialScopeDetails list credential scope details
type ListCredentialScopeDetails struct {
	Count   uint32                   `json:"count"`
	Details []*table.CredentialScope `json:"details"`
}

// ListCredentialScopesOption credential scopes option
type ListCredentialScopesOption struct {
	BizID        uint32 `json:"biz_id"`
	CredentialId uint32 `json:"credential_id"`
}

// // UpdateCredentialScopesOption update credential scopes option
// type UpdateCredentialScopesOption struct {
// 	BizID        uint32 `json:"biz_id"`
// 	CredentialId uint32 `json:"credential_id"`
// 	Updated      []*pbcrs.UpdateScopeSpec
// 	Created      []string
// 	Deleted      []uint32
// }

// // Validate validate update credential scopes option
// func (option *UpdateCredentialScopesOption) Validate() error {
// 	if option.BizID == 0 {
// 		return errf.Newf(errf.InvalidParameter, "biz id cannot be empty")
// 	}
// 	if option.CredentialId == 0 {
// 		return errf.Newf(errf.InvalidParameter, "credential id cannot be empty")
// 	}
// 	if len(option.Updated) == 0 && len(option.Created) == 0 && len(option.Deleted) == 0 {
// 		return errf.Newf(errf.InvalidParameter, "no updated credential scopes")
// 	}
// 	return nil
// }
