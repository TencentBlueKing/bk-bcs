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
