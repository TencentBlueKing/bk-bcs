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

package pbtr

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/protocol/core/config-item"
	"bscp.io/pkg/protocol/core/content"
)

// TemplateRelease convert pb TemplateRelease to table TemplateRelease
func (m *TemplateRelease) TemplateRelease() *table.TemplateRelease {
	if m == nil {
		return nil
	}

	return &table.TemplateRelease{
		ID:         m.Id,
		Spec:       m.Spec.TemplateReleaseSpec(),
		Attachment: m.Attachment.TemplateReleaseAttachment(),
	}
}

// TemplateReleaseSpec convert pb TemplateReleaseSpec to table TemplateReleaseSpec
func (m *TemplateReleaseSpec) TemplateReleaseSpec() *table.TemplateReleaseSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateReleaseSpec{
		ReleaseName: m.ReleaseName,
		ReleaseMemo: m.ReleaseMemo,
		Name:        m.Name,
		Path:        m.Path,
		FileType:    table.FileFormat(m.FileType),
		FileMode:    table.FileMode(m.FileMode),
		Permission:  m.Permission.FilePermission(),
		ContentSpec: m.ContentSpec.ContentSpec(),
	}
}

// PbTemplateReleaseSpec convert table TemplateReleaseSpec to pb TemplateReleaseSpec
func PbTemplateReleaseSpec(spec *table.TemplateReleaseSpec) *TemplateReleaseSpec {
	if spec == nil {
		return nil
	}

	return &TemplateReleaseSpec{
		ReleaseName: spec.ReleaseName,
		ReleaseMemo: spec.ReleaseMemo,
		Name:        spec.Name,
		Path:        spec.Path,
		FileType:    string(spec.FileType),
		FileMode:    string(spec.FileMode),
		Permission:  pbci.PbFilePermission(spec.Permission),
		ContentSpec: pbcontent.PbContentSpec(spec.ContentSpec),
	}
}

// TemplateReleaseAttachment convert pb TemplateReleaseAttachment to table TemplateReleaseAttachment
func (m *TemplateReleaseAttachment) TemplateReleaseAttachment() *table.TemplateReleaseAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateReleaseAttachment{
		BizID:           m.BizId,
		TemplateSpaceID: m.TemplateSpaceId,
		TemplateID:      m.TemplateId,
	}
}

// PbTemplateReleaseAttachment convert table TemplateReleaseAttachment to pb TemplateReleaseAttachment
func PbTemplateReleaseAttachment(at *table.TemplateReleaseAttachment) *TemplateReleaseAttachment {
	if at == nil {
		return nil
	}

	return &TemplateReleaseAttachment{
		BizId:           at.BizID,
		TemplateSpaceId: at.TemplateSpaceID,
		TemplateId:      at.TemplateID,
	}
}

// PbTemplateReleases convert table TemplateRelease to pb TemplateRelease
func PbTemplateReleases(s []*table.TemplateRelease) []*TemplateRelease {
	if s == nil {
		return make([]*TemplateRelease, 0)
	}

	result := make([]*TemplateRelease, 0)
	for _, one := range s {
		result = append(result, PbTemplateRelease(one))
	}

	return result
}

// PbTemplateRelease convert table TemplateRelease to pb TemplateRelease
func PbTemplateRelease(s *table.TemplateRelease) *TemplateRelease {
	if s == nil {
		return nil
	}

	return &TemplateRelease{
		Id:         s.ID,
		Spec:       PbTemplateReleaseSpec(s.Spec),
		Attachment: PbTemplateReleaseAttachment(s.Attachment),
		Revision:   pbbase.PbCreatedRevision(s.Revision),
	}
}
