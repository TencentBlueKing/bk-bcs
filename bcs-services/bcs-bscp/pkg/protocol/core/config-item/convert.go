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

package pbci

import (
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/protocol/core/base"
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

// PbConfigItems convert table ConfigItems to pb ConfigItems
func PbConfigItems(cis []*table.ConfigItem) []*ConfigItem {
	if cis == nil {
		return make([]*ConfigItem, 0)
	}

	result := make([]*ConfigItem, 0)
	for _, ci := range cis {
		result = append(result, PbConfigItem(ci))
	}

	return result
}

// PbConfigItem convert table ConfigItem to pb ConfigItem
func PbConfigItem(ci *table.ConfigItem) *ConfigItem {
	if ci == nil {
		return nil
	}

	return &ConfigItem{
		Id:         ci.ID,
		Spec:       PbConfigItemSpec(ci.Spec),
		Attachment: PbConfigItemAttachment(ci.Attachment),
		Revision:   pbbase.PbRevision(ci.Revision),
	}
}

// PbConfigItemCounts
func PbConfigItemCounts(ccs []*table.ListConfigItemCounts) []*ListConfigItemCounts {
	if ccs == nil {
		return make([]*ListConfigItemCounts, 0)
	}

	result := make([]*ListConfigItemCounts, 0)
	for _, cc := range ccs {
		result = append(result, PbConfigItemCount(cc))
	}

	return result
}

// PbConfigItemCount
func PbConfigItemCount(cc *table.ListConfigItemCounts) *ListConfigItemCounts {
	if cc == nil {
		return nil
	}

	return &ListConfigItemCounts{
		AppId:    cc.AppId,
		Count:    cc.Count,
		UpdateAt: cc.UpdateAt,
	}
}
