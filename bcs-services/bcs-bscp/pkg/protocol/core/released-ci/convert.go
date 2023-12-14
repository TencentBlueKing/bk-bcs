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

// Package pbrci provides released config_item core protocol struct and convert functions.
package pbrci

import (
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcommit "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
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
			CommitSpec: &pbcommit.ReleasedCommitSpec{
				ContentId: one.CommitSpec.ContentID,
				Content: &pbcontent.ReleasedContentSpec{
					Signature: one.CommitSpec.Signature,
					ByteSize:  one.CommitSpec.ByteSize,
				},
			},
			ConfigItemId: one.ConfigItemID,
			Spec: &pbci.ConfigItemSpec{
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
		Id:           rci.ID,
		ReleaseId:    rci.ReleaseID,
		CommitId:     rci.CommitID,
		CommitSpec:   pbcommit.PbReleasedCommitSpec(rci.CommitSpec),
		ConfigItemId: rci.ConfigItemID,
		Spec:         pbci.PbConfigItemSpec(rci.ConfigItemSpec),
		Attachment:   pbci.PbConfigItemAttachment(rci.Attachment),
		Revision:     pbbase.PbRevision(rci.Revision),
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
		Revision: &pbbase.Revision{
			Creator:  rci.Revision.Creator,
			Reviser:  rci.Revision.Creator,
			CreateAt: rci.Revision.CreatedAt.Format(time.RFC3339),
			UpdateAt: rci.Revision.CreatedAt.Format(time.RFC3339),
		},
	}
}

// PbConfigItemState convert config item state
func PbConfigItemState(cis []*table.ConfigItem, fileRelease []*table.ReleasedConfigItem) []*pbci.ConfigItem {
	releaseMap := make(map[uint32]*table.ReleasedConfigItem, len(fileRelease))
	for _, release := range fileRelease {
		releaseMap[release.ConfigItemID] = release
	}

	result := make([]*pbci.ConfigItem, 0)
	for _, ci := range cis {
		var fileState string
		if len(fileRelease) == 0 {
			fileState = constant.FileStateAdd
		} else {
			if _, ok := releaseMap[ci.ID]; ok {
				if ci.Revision.UpdatedAt.After(releaseMap[ci.ID].Revision.CreatedAt) {
					fileState = constant.FileStateRevise
				} else {
					fileState = constant.FileStateUnchange
				}
				delete(releaseMap, ci.ID)
			}
		}
		if len(fileState) == 0 {
			fileState = constant.FileStateAdd
		}
		result = append(result, pbci.PbConfigItem(ci, fileState))
	}
	for _, file := range releaseMap {
		result = append(result, PbConfigItem(file, constant.FileStateDelete))
	}
	return sortConfigItemsByState(result)
}

// sortConfigItemsByState sort as add > revise > unchange > delete
func sortConfigItemsByState(cis []*pbci.ConfigItem) []*pbci.ConfigItem {
	result := make([]*pbci.ConfigItem, 0)
	add := make([]*pbci.ConfigItem, 0)
	del := make([]*pbci.ConfigItem, 0)
	revise := make([]*pbci.ConfigItem, 0)
	unchange := make([]*pbci.ConfigItem, 0)
	for _, ci := range cis {
		switch ci.FileState {
		case constant.FileStateAdd:
			add = append(add, ci)
		case constant.FileStateDelete:
			del = append(del, ci)
		case constant.FileStateRevise:
			revise = append(revise, ci)
		case constant.FileStateUnchange:
			unchange = append(unchange, ci)
		}
	}
	result = append(result, add...)
	result = append(result, revise...)
	result = append(result, unchange...)
	result = append(result, del...)
	return result
}
