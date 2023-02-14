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

package pbgc

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// GroupCategorySpec convert pb GroupCategorySpec to table GroupCategorySpec
func (m *GroupCategorySpec) GroupCategorySpec() *table.GroupCategorySpec {
	if m == nil {
		return nil
	}

	return &table.GroupCategorySpec{
		Name: m.Name,
	}
}

// PbGroupCategorySpec convert table GroupCategorySpec to pb GroupCategorySpec
func PbGroupCategorySpec(spec *table.GroupCategorySpec) *GroupCategorySpec {
	if spec == nil {
		return nil
	}

	return &GroupCategorySpec{
		Name: spec.Name,
	}
}

// GroupCategoryAttachment convert pb GroupCategoryAttachment to table GroupCategoryAttachment
func (m *GroupCategoryAttachment) GroupCategoryAttachment() *table.GroupCategoryAttachment {
	if m == nil {
		return nil
	}

	return &table.GroupCategoryAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbGroupCategoryAttachment convert table GroupCategoryAttachment to pb GroupCategoryAttachment
func PbGroupCategoryAttachment(at *table.GroupCategoryAttachment) *GroupCategoryAttachment {
	if at == nil {
		return nil
	}

	return &GroupCategoryAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbGroupCategories convert table GroupCategory to pb GroupCategory
func PbGroupCategories(s []*table.GroupCategory) []*GroupCategory {
	if s == nil {
		return make([]*GroupCategory, 0)
	}

	result := make([]*GroupCategory, 0)
	for _, one := range s {
		result = append(result, PbGroupCategory(one))
	}

	return result
}

// PbGroupCategory convert table GroupCategory to pb GroupCategory
func PbGroupCategory(s *table.GroupCategory) *GroupCategory {
	if s == nil {
		return nil
	}

	return &GroupCategory{
		Id:         s.ID,
		Spec:       PbGroupCategorySpec(s.Spec),
		Attachment: PbGroupCategoryAttachment(s.Attachment),
		Revision:   pbbase.PbCreatedRevision(s.Revision),
	}
}
