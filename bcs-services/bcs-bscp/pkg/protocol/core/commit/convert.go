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

// Package pbcommit provides commit core protocol struct and convert functions.
package pbcommit

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
)

// CommitSpec convert pb CommitSpec to table CommitSpec
func (m *CommitSpec) CommitSpec() *table.CommitSpec {
	if m == nil {
		return nil
	}

	return &table.CommitSpec{
		ContentID: m.ContentId,
		Memo:      m.Memo,
		Content:   m.Content.ContentSpec(),
	}
}

// PbCommitSpec convert table CommitSpec to pb CommitSpec
func PbCommitSpec(spec *table.CommitSpec) *CommitSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &CommitSpec{
		ContentId: spec.ContentID,
		Memo:      spec.Memo,
		Content:   pbcontent.PbContentSpec(spec.Content),
	}
}

// ReleasedCommitSpec convert pb ReleasedCommitSpec to table ReleasedCommitSpec
func (m *ReleasedCommitSpec) ReleasedCommitSpec() *table.ReleasedCommitSpec {
	if m == nil {
		return nil
	}

	return &table.ReleasedCommitSpec{
		ContentID: m.ContentId,
		Memo:      m.Memo,
		Content:   m.Content.ReleasedContentSpec(),
	}
}

// PbReleasedCommitSpec convert table ReleasedCommitSpec to pb ReleasedCommitSpec
func PbReleasedCommitSpec(spec *table.ReleasedCommitSpec) *ReleasedCommitSpec {
	if spec == nil {
		return nil
	}

	return &ReleasedCommitSpec{
		ContentId: spec.ContentID,
		Memo:      spec.Memo,
		Content:   pbcontent.PbReleasedContentSpec(spec.Content),
	}
}

// CommitAttachment convert pb CommitAttachment to table CommitAttachment
func (m *CommitAttachment) CommitAttachment() *table.CommitAttachment {
	if m == nil {
		return nil
	}

	return &table.CommitAttachment{
		BizID:        m.BizId,
		AppID:        m.AppId,
		ConfigItemID: m.ConfigItemId,
	}
}

// PbCommitAttachment convert table CommitAttachment to pb CommitAttachment
func PbCommitAttachment(at *table.CommitAttachment) *CommitAttachment { //nolint:revive
	if at == nil {
		return nil
	}

	return &CommitAttachment{
		BizId:        at.BizID,
		AppId:        at.AppID,
		ConfigItemId: at.ConfigItemID,
	}
}

// PbCommits convert table Commits to pb Commits
func PbCommits(cs []*table.Commit) []*Commit {
	if cs == nil {
		return make([]*Commit, 0)
	}

	result := make([]*Commit, 0)
	for _, c := range cs {
		result = append(result, PbCommit(c))
	}

	return result
}

// PbCommit convert table Commits to pb Commits
func PbCommit(c *table.Commit) *Commit {
	if c == nil {
		return nil
	}

	return &Commit{
		Id:         c.ID,
		Spec:       PbCommitSpec(c.Spec),
		Attachment: PbCommitAttachment(c.Attachment),
		Revision:   pbbase.PbCreatedRevision(c.Revision),
	}
}
