/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package pbcrs provides credential scope core protocol struct and convert functions.
package pbcrs

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// CredentialAttachment convert pb CredentialAttachment to table CredentialScopeAttachment
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
		Revision:   pbbase.PbRevision(s.Revision),
	}, nil
}

// PbCredentialScopeSpec convert table CredentialScopeSpec to pb CredentialScopeSpec
func PbCredentialScopeSpec(spec *table.CredentialScopeSpec) (*CredentialScopeSpec, error) {
	if spec == nil {
		return nil, nil
	}

	app, scope, err := spec.CredentialScope.Split()
	if err != nil {
		return nil, err
	}

	return &CredentialScopeSpec{
		App:   app,
		Scope: scope,
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
