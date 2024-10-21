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

// Package pbatb provides app template binding core protocol struct and convert functions.
package pbatb

import (
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
)

// AppTemplateBinding convert pb AppTemplateBinding to table AppTemplateBinding
func (m *AppTemplateBinding) AppTemplateBinding() (*table.AppTemplateBinding, error) {
	if m == nil {
		return nil, nil
	}

	return &table.AppTemplateBinding{
		ID:         m.Id,
		Spec:       m.Spec.AppTemplateBindingSpec(),
		Attachment: m.Attachment.AppTemplateBindingAttachment(),
	}, nil
}

// AppTemplateBindingSpec convert pb AppTemplateBindingSpec to table AppTemplateBindingSpec
func (m *AppTemplateBindingSpec) AppTemplateBindingSpec() *table.AppTemplateBindingSpec {
	if m == nil {
		return nil
	}

	result := make([]*table.TemplateBinding, 0)
	for _, one := range m.Bindings {
		result = append(result, one.TemplateBinding())
	}

	return &table.AppTemplateBindingSpec{
		TemplateSpaceIDs:    m.TemplateSpaceIds,
		TemplateSetIDs:      m.TemplateSetIds,
		TemplateIDs:         m.TemplateIds,
		TemplateRevisionIDs: m.TemplateRevisionIds,
		LatestTemplateIDs:   m.LatestTemplateIds,
		Bindings:            result,
	}
}

// PbAppTemplateBindingSpec convert table AppTemplateBindingSpec to pb AppTemplateBindingSpec
func PbAppTemplateBindingSpec(spec *table.AppTemplateBindingSpec) *AppTemplateBindingSpec {
	if spec == nil {
		return nil
	}

	result := make([]*TemplateBinding, 0)
	for _, one := range spec.Bindings {
		result = append(result, PbTemplateBinding(one))
	}

	return &AppTemplateBindingSpec{
		TemplateSpaceIds:    spec.TemplateSpaceIDs,
		TemplateSetIds:      spec.TemplateSetIDs,
		TemplateIds:         spec.TemplateIDs,
		TemplateRevisionIds: spec.TemplateRevisionIDs,
		LatestTemplateIds:   spec.LatestTemplateIDs,
		Bindings:            result,
	}
}

// TemplateBinding convert pb TemplateBinding to table TemplateBinding
func (m *TemplateBinding) TemplateBinding() *table.TemplateBinding {
	if m == nil {
		return nil
	}

	result := make([]*table.TemplateRevisionBinding, 0)
	for _, one := range m.TemplateRevisions {
		result = append(result, one.TemplateRevisionBinding())
	}

	return &table.TemplateBinding{
		TemplateSetID:     m.TemplateSetId,
		TemplateRevisions: result,
	}
}

// PbTemplateBinding convert table TemplateBinding to pb TemplateBinding
func PbTemplateBinding(spec *table.TemplateBinding) *TemplateBinding {
	if spec == nil {
		return nil
	}

	result := make([]*TemplateRevisionBinding, 0)
	for _, one := range spec.TemplateRevisions {
		result = append(result, PbTemplateRevisionBinding(one))
	}

	return &TemplateBinding{
		TemplateSetId:     spec.TemplateSetID,
		TemplateRevisions: result,
	}
}

// TemplateRevisionBinding convert pb TemplateRevisionBinding to table TemplateRevisionBinding
func (m *TemplateRevisionBinding) TemplateRevisionBinding() *table.TemplateRevisionBinding {
	if m == nil {
		return nil
	}

	return &table.TemplateRevisionBinding{
		TemplateID:         m.TemplateId,
		TemplateRevisionID: m.TemplateRevisionId,
		IsLatest:           m.IsLatest,
	}
}

// PbTemplateRevisionBinding convert table TemplateRevisionBinding to pb TemplateRevisionBinding
func PbTemplateRevisionBinding(spec *table.TemplateRevisionBinding) *TemplateRevisionBinding {
	if spec == nil {
		return nil
	}

	return &TemplateRevisionBinding{
		TemplateId:         spec.TemplateID,
		TemplateRevisionId: spec.TemplateRevisionID,
		IsLatest:           spec.IsLatest,
	}
}

// AppTemplateBindingAttachment convert pb AppTemplateBindingAttachment to table AppTemplateBindingAttachment
func (m *AppTemplateBindingAttachment) AppTemplateBindingAttachment() *table.AppTemplateBindingAttachment {
	if m == nil {
		return nil
	}

	return &table.AppTemplateBindingAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbAppTemplateBindingAttachment convert table AppTemplateBindingAttachment to pb AppTemplateBindingAttachment
func PbAppTemplateBindingAttachment(at *table.AppTemplateBindingAttachment) *AppTemplateBindingAttachment {
	if at == nil {
		return nil
	}

	return &AppTemplateBindingAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbAppTemplateBindings convert table AppTemplateBinding to pb AppTemplateBinding
func PbAppTemplateBindings(s []*table.AppTemplateBinding) []*AppTemplateBinding {
	if s == nil {
		return make([]*AppTemplateBinding, 0)
	}

	result := make([]*AppTemplateBinding, 0)
	for _, one := range s {
		result = append(result, PbAppTemplateBinding(one))
	}

	return result
}

// PbAppTemplateBinding convert table AppTemplateBinding to pb AppTemplateBinding
func PbAppTemplateBinding(s *table.AppTemplateBinding) *AppTemplateBinding {
	if s == nil {
		return nil
	}

	return &AppTemplateBinding{
		Id:         s.ID,
		Spec:       PbAppTemplateBindingSpec(s.Spec),
		Attachment: PbAppTemplateBindingAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}

// PbReleasedAppBoundTmplRevisions convert table ReleasedAppTemplate to pb ReleasedAppBoundTmplRevision
func PbReleasedAppBoundTmplRevisions(s []*table.ReleasedAppTemplate) []*ReleasedAppBoundTmplRevision {
	if s == nil {
		return make([]*ReleasedAppBoundTmplRevision, 0)
	}

	result := make([]*ReleasedAppBoundTmplRevision, 0)
	for _, one := range s {
		result = append(result, PbReleasedAppBoundTmplRevision(one))
	}

	return result
}

// PbReleasedAppBoundTmplRevision convert table ReleasedAppTemplate to pb ReleasedAppBoundTmplRevision
func PbReleasedAppBoundTmplRevision(s *table.ReleasedAppTemplate) *ReleasedAppBoundTmplRevision {
	if s == nil {
		return nil
	}

	return &ReleasedAppBoundTmplRevision{
		TemplateSpaceId:      s.Spec.TemplateSpaceID,
		TemplateSpaceName:    s.Spec.TemplateSpaceName,
		TemplateSetId:        s.Spec.TemplateSetID,
		TemplateSetName:      s.Spec.TemplateSetName,
		TemplateId:           s.Spec.TemplateID,
		Name:                 s.Spec.Name,
		Path:                 s.Spec.Path,
		TemplateRevisionId:   s.Spec.TemplateRevisionID,
		IsLatest:             s.Spec.IsLatest,
		TemplateRevisionName: s.Spec.TemplateRevisionName,
		TemplateRevisionMemo: s.Spec.TemplateRevisionMemo,
		FileType:             s.Spec.FileType,
		FileMode:             s.Spec.FileMode,
		Signature:            s.Spec.Signature,
		ByteSize:             s.Spec.ByteSize,
		OriginSignature:      s.Spec.OriginSignature,
		OriginByteSize:       s.Spec.OriginByteSize,
		Creator:              s.Revision.Creator,
		Reviser:              s.Revision.Reviser,
		CreateAt:             s.Revision.CreatedAt.Format(time.RFC3339),
		UpdateAt:             s.Revision.UpdatedAt.Format(time.RFC3339),
		Md5:                  s.Spec.Md5,
		Permission:           pbci.PbFilePermission(s.Spec.Permission),
	}
}

// PbAppBoundTmplRevisionsFromReleased convert table ReleasedAppTemplate to pb AppBoundTmplRevision
func PbAppBoundTmplRevisionsFromReleased(releasedTmpls []*table.ReleasedAppTemplate) []*AppBoundTmplRevision {
	tmplRevisions := make([]*AppBoundTmplRevision, len(releasedTmpls))
	for idx, r := range releasedTmpls {
		tmplRevisions[idx] = &AppBoundTmplRevision{
			TemplateSpaceId:      r.Spec.TemplateSpaceID,
			TemplateSpaceName:    r.Spec.TemplateSpaceName,
			TemplateSetId:        r.Spec.TemplateSetID,
			TemplateSetName:      r.Spec.TemplateSetName,
			TemplateId:           r.Spec.TemplateID,
			Name:                 r.Spec.Name,
			Path:                 r.Spec.Path,
			TemplateRevisionId:   r.Spec.TemplateRevisionID,
			IsLatest:             r.Spec.IsLatest,
			TemplateRevisionName: r.Spec.TemplateRevisionName,
			TemplateRevisionMemo: r.Spec.TemplateRevisionMemo,
			FileType:             r.Spec.FileType,
			FileMode:             r.Spec.FileMode,
			Signature:            r.Spec.Signature,
			ByteSize:             r.Spec.ByteSize,
			Creator:              r.Revision.Creator,
			CreateAt:             r.Revision.CreatedAt.Format(time.RFC3339),
			Permission:           pbci.PbFilePermission(r.Spec.Permission),
		}
	}
	return tmplRevisions
}
