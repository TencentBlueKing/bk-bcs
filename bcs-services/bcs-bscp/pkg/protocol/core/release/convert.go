/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbrelease

import (
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/protocol/core/base"
)

// ReleaseSpec convert pb ReleaseSpec to table ReleaseSpec
func (m *ReleaseSpec) ReleaseSpec() *table.ReleaseSpec {
	if m == nil {
		return nil
	}

	return &table.ReleaseSpec{
		Name:       m.Name,
		Memo:       m.Memo,
		Deprecated: m.Deprecated,
		PublishNum: m.PublishNum,
		Hook:       m.Hook.ReleaseSpecHook(),
	}
}

// ReleaseSpecHook convert pb ReleaseSpecHook to table ReleaseSpecHook
func (h *Hook) ReleaseSpecHook() *table.ReleaseHook {

	if h == nil {
		return &table.ReleaseHook{}
	}

	return &table.ReleaseHook{
		PreHookID:         h.PreHookId,
		PreHookReleaseID:  h.PreHookReleaseId,
		PostHookID:        h.PostHookId,
		PostHookReleaseID: h.PostHookReleaseId,
	}

}

// PbReleaseSpec convert table ReleaseSpec to pb ReleaseSpec
func PbReleaseSpec(spec *table.ReleaseSpec) *ReleaseSpec {
	if spec == nil {
		return nil
	}

	return &ReleaseSpec{
		Name:       spec.Name,
		Memo:       spec.Memo,
		Deprecated: spec.Deprecated,
		PublishNum: spec.PublishNum,
		Hook: &Hook{
			PreHookId:         spec.Hook.PreHookID,
			PreHookReleaseId:  spec.Hook.PreHookReleaseID,
			PostHookId:        spec.Hook.PostHookID,
			PostHookReleaseId: spec.Hook.PostHookReleaseID,
		},
	}
}

// ReleaseAttachment convert pb ReleaseAttachment to table ReleaseAttachment
func (m *ReleaseAttachment) ReleaseAttachment() *table.ReleaseAttachment {
	if m == nil {
		return nil
	}

	return &table.ReleaseAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbReleaseAttachment convert table ReleaseAttachment to pb ReleaseAttachment
func PbReleaseAttachment(at *table.ReleaseAttachment) *ReleaseAttachment {
	if at == nil {
		return nil
	}

	return &ReleaseAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbReleases convert table Release to pb Release
func PbReleases(rls []*table.Release) []*Release {
	if rls == nil {
		return make([]*Release, 0)
	}

	result := make([]*Release, 0)
	for _, r := range rls {
		result = append(result, PbRelease(r))
	}

	return result
}

// PbRelease convert table Release to pb Release
func PbRelease(rl *table.Release) *Release {
	if rl == nil {
		return nil
	}

	return &Release{
		Id:         rl.ID,
		Spec:       PbReleaseSpec(rl.Spec),
		Attachment: PbReleaseAttachment(rl.Attachment),
		Revision:   pbbase.PbCreatedRevision(rl.Revision),
	}
}
