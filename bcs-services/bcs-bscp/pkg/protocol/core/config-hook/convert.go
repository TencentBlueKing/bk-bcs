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

package pbch

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// ConfigHook convert pb ConfigHook to table ConfigHook
func (m *ConfigHook) ConfigHook() (*table.ConfigHook, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.ConfigHookSpec()
	if err != nil {
		return nil, err
	}

	return &table.ConfigHook{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.ConfigHookAttachment(),
	}, nil
}

// ConfigHookSpec convert pb ConfigHookSpec to table ConfigHookSpec
func (m *ConfigHookSpec) ConfigHookSpec() (*table.ConfigHookSpec, error) {
	if m == nil {
		return nil, nil
	}

	return &table.ConfigHookSpec{
		PreHookID:         m.PreHookId,
		PreHookReleaseID:  m.PreHookReleaseId,
		PostHookID:        m.PostHookId,
		PostHookReleaseID: m.PostHookReleaseId,
	}, nil
}

// ConfigHookAttachment convert pb ConfigHookAttachment to table ConfigHookAttachment
func (m *ConfigHookAttachment) ConfigHookAttachment() *table.ConfigHookAttachment {
	if m == nil {
		return nil
	}

	return &table.ConfigHookAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbConfigHook convert table ConfigHook to pb ConfigHook
func PbConfigHook(c *table.ConfigHook) *ConfigHook {
	if c == nil {
		return nil
	}
	return &ConfigHook{
		Id:         c.ID,
		Spec:       PbConfigHookSpec(c.Spec),
		Attachment: PbConfigHookAttachment(c.Attachment),
		Revision:   pbbase.PbRevision(c.Revision),
	}
}

// PbConfigHookSpec convert table ConfigHookSpec to pb ConfigHookSpec
func PbConfigHookSpec(spec *table.ConfigHookSpec) *ConfigHookSpec {
	if spec == nil {
		return nil
	}

	return &ConfigHookSpec{
		PreHookId:         spec.PreHookID,
		PreHookReleaseId:  spec.PreHookReleaseID,
		PostHookId:        spec.PostHookID,
		PostHookReleaseId: spec.PostHookReleaseID,
	}

}

// PbConfigHookAttachment convert table ConfigHookAttachment to pb ConfigHookAttachment
func PbConfigHookAttachment(at *table.ConfigHookAttachment) *ConfigHookAttachment {
	if at == nil {
		return nil
	}

	return &ConfigHookAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}
