package pbcredential

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// CredentialSpec  convert pb CredentialSpec to table CredentialSpec
func (c *CredentialSpec) CredentialSpec() (*table.CredentialSpec, error) {
	if c == nil {
		return nil, nil
	}

	return &table.CredentialSpec{
		CredentialType: table.CredentialType(c.CredentialType),
		EncCredential:  c.EncCredential,
		EncAlgorithm:   c.EncAlgorithm,
		Memo:           c.Memo,
		Enable:         c.Enable,
	}, nil
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
func PbCredentials(s []*table.Credential) ([]*CredentialList, error) {
	if s == nil {
		return make([]*CredentialList, 0), nil
	}

	result := make([]*CredentialList, 0)
	for _, one := range s {
		credential, err := PbCredential(one)
		if err != nil {
			return nil, err
		}
		result = append(result, credential)
	}

	return result, nil
}

// PbCredential convert table Credential to pb Credential
func PbCredential(s *table.Credential) (*CredentialList, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbCredentialSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &CredentialList{
		Id:         s.ID,
		Spec:       spec,
		Attachment: PbCredentialAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}, nil
}

// PbCredentialSpec convert table CredentialSpec to pb CredentialSpec
func PbCredentialSpec(spec *table.CredentialSpec) (*CredentialSpec, error) {
	if spec == nil {
		return nil, nil
	}

	return &CredentialSpec{
		CredentialType: string(spec.CredentialType),
		EncCredential:  spec.EncCredential,
		EncAlgorithm:   spec.EncAlgorithm,
		Enable:         spec.Enable,
		Memo:           spec.Memo,
	}, nil
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
