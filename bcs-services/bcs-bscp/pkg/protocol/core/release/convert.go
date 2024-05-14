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

// Package pbrelease provides release core protocol struct and convert functions.
package pbrelease

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
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
	}
}

// PbReleaseSpec convert table ReleaseSpec to pb ReleaseSpec
//
//nolint:revive
func PbReleaseSpec(spec *table.ReleaseSpec) *ReleaseSpec {
	if spec == nil {
		return nil
	}

	return &ReleaseSpec{
		Name:       spec.Name,
		Memo:       spec.Memo,
		Deprecated: spec.Deprecated,
		PublishNum: spec.PublishNum,
	}
}

// PbReleaseStatus convert table ReleaseSpec to pb ReleaseStatus
//
//nolint:revive
func PbReleaseStatus(spec *table.ReleaseSpec) *ReleaseStatus {
	if spec == nil {
		return nil
	}

	return &ReleaseStatus{
		FullyReleased: spec.FullyReleased,
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
func PbReleaseAttachment(at *table.ReleaseAttachment) *ReleaseAttachment { //nolint:revive
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
		Status:     PbReleaseStatus(rl.Spec),
	}
}
