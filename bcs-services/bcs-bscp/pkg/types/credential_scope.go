package types

import (
	"bscp.io/pkg/dal/table"
)

type ListCredentialScopeDetails struct {
	Count   uint32                   `json:"count"`
	Details []*table.CredentialScope `json:"details"`
}

type ListCredentialScopesOption struct {
	BizID        uint32 `json:"biz_id"`
	CredentialId uint32 `json:"credential_id"`
}
