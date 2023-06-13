package pbcredential

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// CredentialSpec  convert pb CredentialSpec to table CredentialSpec
func (c *CredentialSpec) CredentialSpec() *table.CredentialSpec {
	if c == nil {
		return nil
	}

	return &table.CredentialSpec{
		CredentialType: table.CredentialType(c.CredentialType),
		EncCredential:  c.EncCredential,
		EncAlgorithm:   c.EncAlgorithm,
		Memo:           c.Memo,
		Enable:         c.Enable,
	}
}

// CredentialAttachment convert pb CredentialAttachment to table CredentialAttachment
func (m *CredentialAttachment) CredentialAttachment() *table.CredentialAttachment {
	if m == nil {
		return nil
	}

	return &table.CredentialAttachment{
		BizID: m.BizId,
	}
}

// PbCredentials Credentials
func PbCredentials(s []*table.Credential) []*CredentialList {
	if s == nil {
		return make([]*CredentialList, 0)
	}

	result := make([]*CredentialList, 0)
	for _, one := range s {
		result = append(result, PbCredential(one))
	}

	return result
}

// PbCredential convert table Credential to pb Credential
func PbCredential(s *table.Credential) *CredentialList {
	if s == nil {
		return nil
	}

	return &CredentialList{
		Id:         s.ID,
		Spec:       PbCredentialSpec(s.Spec),
		Attachment: PbCredentialAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}

// PbCredentialSpec convert table CredentialSpec to pb CredentialSpec
func PbCredentialSpec(spec *table.CredentialSpec) *CredentialSpec {
	if spec == nil {
		return nil
	}

	return &CredentialSpec{
		CredentialType: string(spec.CredentialType),
		EncCredential:  spec.EncCredential,
		EncAlgorithm:   spec.EncAlgorithm,
		Enable:         spec.Enable,
		Memo:           spec.Memo,
	}
}

// PbCredentialAttachment convert table CredentialAttachment to pb CredentialAttachment
func PbCredentialAttachment(at *table.CredentialAttachment) *CredentialAttachment {
	if at == nil {
		return nil
	}

	return &CredentialAttachment{
		BizId: at.BizID,
	}
}
