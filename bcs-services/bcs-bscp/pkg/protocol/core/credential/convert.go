/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

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
