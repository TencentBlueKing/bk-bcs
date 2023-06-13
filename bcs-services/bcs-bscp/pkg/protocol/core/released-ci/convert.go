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

package pbrci

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	"bscp.io/pkg/types"
)

// PbReleasedConfigItems convert table ReleasedConfigItems to pb ReleasedConfigItems
func PbReleasedConfigItems(rcis []*table.ReleasedConfigItem) []*ReleasedConfigItem {
	if rcis == nil {
		return make([]*ReleasedConfigItem, 0)
	}

	result := make([]*ReleasedConfigItem, len(rcis))
	for idx := range rcis {
		result[idx] = PbReleasedConfigItem(rcis[idx])
	}

	return result
}

// PbConfigItems convert table ReleasedConfigItems to pb ConfigItems
func PbConfigItems(rcis []*table.ReleasedConfigItem) []*pbci.ConfigItem {
	if rcis == nil {
		return make([]*pbci.ConfigItem, 0)
	}

	result := make([]*pbci.ConfigItem, len(rcis))
	for idx := range rcis {
		result[idx] = PbConfigItem(rcis[idx], "")
	}

	return result
}

// PbReleasedCIFromCache convert types ReleaseCICache to pb ReleasedConfigItems
func PbReleasedCIFromCache(rs []*types.ReleaseCICache) []*ReleasedConfigItem {
	list := make([]*ReleasedConfigItem, len(rs))

	for index, one := range rs {
		list[index] = &ReleasedConfigItem{
			Id:        one.ID,
			ReleaseId: one.ReleaseID,
			CommitId:  one.CommitID,
			CommitSpec: &pbcommit.CommitSpec{
				ContentId: one.CommitSpec.ContentID,
				Content: &pbcontent.ContentSpec{
					Signature: one.CommitSpec.Signature,
					ByteSize:  one.CommitSpec.ByteSize,
				},
			},
			ConfigItemId: one.ConfigItemID,
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     one.ConfigItemSpec.Name,
				Path:     one.ConfigItemSpec.Path,
				FileType: string(one.ConfigItemSpec.FileType),
				FileMode: string(one.ConfigItemSpec.FileMode),
				Permission: &pbci.FilePermission{
					User:      one.ConfigItemSpec.Permission.User,
					UserGroup: one.ConfigItemSpec.Permission.UserGroup,
					Privilege: one.ConfigItemSpec.Permission.Privilege,
				},
			},
			Attachment: &pbci.ConfigItemAttachment{
				BizId: one.Attachment.BizID,
				AppId: one.Attachment.AppID,
			},
		}
	}

	return list
}

// PbReleasedConfigItem convert table ReleasedConfigItem to pb ReleasedConfigItem
func PbReleasedConfigItem(rci *table.ReleasedConfigItem) *ReleasedConfigItem {
	if rci == nil {
		return nil
	}

	return &ReleasedConfigItem{
		Id:             rci.ID,
		ReleaseId:      rci.ReleaseID,
		CommitId:       rci.CommitID,
		CommitSpec:     pbcommit.PbCommitSpec(rci.CommitSpec),
		ConfigItemId:   rci.ConfigItemID,
		ConfigItemSpec: pbci.PbConfigItemSpec(rci.ConfigItemSpec),
		Attachment:     pbci.PbConfigItemAttachment(rci.Attachment),
		Revision:       pbbase.PbRevision(rci.Revision),
	}
}

// PbConfigItem convert table ReleasedConfigItem to pb ConfigItem
func PbConfigItem(rci *table.ReleasedConfigItem, fileState string) *pbci.ConfigItem {
	if rci == nil {
		return nil
	}

	return &pbci.ConfigItem{
		Id:         rci.ConfigItemID,
		FileState:  fileState,
		Spec:       pbci.PbConfigItemSpec(rci.ConfigItemSpec),
		Attachment: pbci.PbConfigItemAttachment(rci.Attachment),
		Revision:   pbbase.PbRevision(rci.Revision),
	}
}

// PbConfigItemState convert config item state
func PbConfigItemState(cis []*table.ConfigItem, fileRelease []*table.ReleasedConfigItem) (
	[]*pbci.ConfigItem, []*pbci.ConfigItem) {
	if cis == nil {
		return make([]*pbci.ConfigItem, 0), nil
	}

	releaseMap := make(map[uint32]*table.ReleasedConfigItem, len(fileRelease))
	for _, release := range fileRelease {
		releaseMap[release.ConfigItemID] = release
	}

	result := make([]*pbci.ConfigItem, 0)
	deleted := make([]*pbci.ConfigItem, 0)
	for _, ci := range cis {
		var fileState string
		if len(fileRelease) == 0 {
			fileState = ADD
		} else {
			if _, ok := releaseMap[ci.ID]; ok {
				if ci.Revision.UpdatedAt == releaseMap[ci.ID].Revision.UpdatedAt {
					fileState = UNCHANGE
				} else {
					fileState = REVISE
				}
				delete(releaseMap, ci.ID)
			}
		}
		if len(fileState) == 0 {
			fileState = ADD
		}
		result = append(result, pbci.PbConfigItem(ci, fileState))
	}

	if len(releaseMap) != 0 {
		for _, file := range releaseMap {
			deleted = append(deleted, PbConfigItem(file, DELETE))
		}
	}

	return result, deleted
}

// 文件状态
const (
	// 增加
	ADD = "ADD"
	//删除
	DELETE = "DELETE"
	//修改
	REVISE = "REVISE"
	//不变
	UNCHANGE = "UNCHANGE"
)
