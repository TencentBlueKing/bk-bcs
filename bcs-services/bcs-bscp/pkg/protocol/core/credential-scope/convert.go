package pbcrs

import (
	"bscp.io/pkg/dal/table"
	pbcredential "bscp.io/pkg/protocol/core/credential"
)

// CredentialScopeAttachment convert pb CredentialAttachment to table CredentialScopeAttachment
func (m *CredentialScopeAttachment) CredentialAttachment() *table.CredentialScopeAttachment {
	if m == nil {
		return nil
	}

	return &table.CredentialScopeAttachment{
		BizID:        m.BizId,
		CredentialId: m.CredentialId,
	}
}

// PbCredentialScopes convert pb CredentialScope to table CredentialScope
func PbCredentialScopes(s []*table.CredentialScope) ([]*CredentialScopeList, error) {
	if s == nil {
		return make([]*CredentialScopeList, 0), nil
	}

	result := make([]*CredentialScopeList, 0)
	for _, one := range s {
		credentialScope, err := PbCredentialScope(one)
		if err != nil {
			return nil, err
		}
		result = append(result, credentialScope)
	}

	return result, nil
}

// PbCredentialScope convert table CredentialScope to pb PbCredentialScope
func PbCredentialScope(s *table.CredentialScope) (*CredentialScopeList, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbCredentialScopeSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &CredentialScopeList{
		Id:         s.ID,
		Spec:       spec,
		Attachment: PbCredentialScopeAttachment(s.Attachment),
		Revision:   pbcredential.PbCredentialRevision(s.Revision),
	}, nil
}

// PbCredentialScopeSpec convert table CredentialScopeSpec to pb CredentialScopeSpec
func PbCredentialScopeSpec(spec *table.CredentialScopeSpec) (*CredentialScopeSpec, error) {
	if spec == nil {
		return nil, nil
	}

	return &CredentialScopeSpec{
		CredentialScope: spec.CredentialScope,
	}, nil
}

// PbCredentialScopeAttachment convert table CredentialScopeAttachment to pb CredentialScopeAttachment
func PbCredentialScopeAttachment(at *table.CredentialScopeAttachment) *CredentialScopeAttachment {
	if at == nil {
		return nil
	}

	return &CredentialScopeAttachment{
		BizId:        at.BizID,
		CredentialId: at.CredentialId,
	}
}
