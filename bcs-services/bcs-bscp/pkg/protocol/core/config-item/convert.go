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

// Package pbci provides config_item core protocol struct and convert functions.
package pbci

import (
	"time"

	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// ConfigItemSpec convert pb ConfigItemSpec to table ConfigItemSpec
func (m *ConfigItemSpec) ConfigItemSpec() *table.ConfigItemSpec {
	if m == nil {
		return nil
	}

	return &table.ConfigItemSpec{
		Name:       m.Name,
		Path:       m.Path,
		FileType:   table.FileFormat(m.FileType),
		FileMode:   table.FileMode(m.FileMode),
		Memo:       m.Memo,
		Permission: m.Permission.FilePermission(),
	}
}

// PbConfigItemSpec convert table ConfigItemSpec to pb ConfigItemSpec
func PbConfigItemSpec(spec *table.ConfigItemSpec) *ConfigItemSpec {
	if spec == nil {
		return nil
	}

	return &ConfigItemSpec{
		Name:       spec.Name,
		Path:       spec.Path,
		FileType:   string(spec.FileType),
		FileMode:   string(spec.FileMode),
		Memo:       spec.Memo,
		Permission: PbFilePermission(spec.Permission),
	}
}

// FilePermission convert pb FilePermission to table FilePermission
func (m *FilePermission) FilePermission() *table.FilePermission {
	if m == nil {
		return nil
	}

	return &table.FilePermission{
		User:      m.User,
		UserGroup: m.UserGroup,
		Privilege: m.Privilege,
	}
}

// PbFilePermission convert table FilePermission to pb FilePermission
func PbFilePermission(fp *table.FilePermission) *FilePermission {
	if fp == nil {
		return nil
	}

	return &FilePermission{
		User:      fp.User,
		UserGroup: fp.UserGroup,
		Privilege: fp.Privilege,
	}
}

// ConfigItemAttachment convert pb ConfigItemAttachment to table ConfigItemAttachment
func (m *ConfigItemAttachment) ConfigItemAttachment() *table.ConfigItemAttachment {
	if m == nil {
		return nil
	}

	return &table.ConfigItemAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbConfigItemAttachment convert table ConfigItemAttachment to pb ConfigItemAttachment
func PbConfigItemAttachment(at *table.ConfigItemAttachment) *ConfigItemAttachment {
	if at == nil {
		return nil
	}

	return &ConfigItemAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbConfigItem convert table ConfigItem to pb ConfigItem
func PbConfigItem(ci *table.ConfigItem, fileState string) *ConfigItem {
	if ci == nil {
		return nil
	}

	return &ConfigItem{
		Id:           ci.ID,
		ConfigItemId: ci.ID,
		FileState:    fileState,
		Spec:         PbConfigItemSpec(ci.Spec),
		Attachment:   PbConfigItemAttachment(ci.Attachment),
		Revision:     pbbase.PbRevision(ci.Revision),
	}
}

// PbConfigItemCounts convert table ListConfigItemCounts to pb ListConfigItemCounts
func PbConfigItemCounts(ccs []*table.ListConfigItemCounts, appList []uint32) []*ListConfigItemCounts {
	if ccs == nil {
		return make([]*ListConfigItemCounts, 0)
	}

	result := make([]*ListConfigItemCounts, 0)
	ccsList := make(map[uint32]*ListConfigItemCounts, 0)
	for _, cc := range ccs {
		ccsList[cc.AppId] = PbConfigItemCount(cc)
	}

	for _, app := range appList {
		if _, ok := ccsList[app]; !ok {
			result = append(result, &ListConfigItemCounts{AppId: app})
		} else {
			result = append(result, ccsList[app])
		}
	}
	return result
}

// PbConfigItemCount convert table ListConfigItemCounts to pb ListConfigItemCounts
func PbConfigItemCount(cc *table.ListConfigItemCounts) *ListConfigItemCounts {
	if cc == nil {
		return nil
	}

	return &ListConfigItemCounts{
		AppId:    cc.AppId,
		Count:    cc.Count,
		UpdateAt: cc.UpdatedAt.Format(time.RFC3339),
	}
}

// PbConfigItemSpecs convert table Templates to pb Templates
func PbConfigItemSpecs(s []*table.ConfigItem) []*ConfigItem {
	if s == nil {
		return make([]*ConfigItem, 0)
	}

	result := make([]*ConfigItem, 0)
	for _, one := range s {
		result = append(result, PbConfigItem(one, ""))
	}

	return result
}
