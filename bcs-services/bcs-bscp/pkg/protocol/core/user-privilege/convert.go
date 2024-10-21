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

// Package pbup xxx
package pbup

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// UserPrivilegeSpec convert pb UserPrivilege to table UserPrivilege
// nolint:revive
func (up *UserPrivilege) UserPrivilege() *table.UserPrivilege {
	if up == nil {
		return nil
	}

	return &table.UserPrivilege{
		ID:         up.GetId(),
		Spec:       &table.UserPrivilegeSpec{},
		Attachment: &table.UserPrivilegeAttachment{},
		Revision:   &table.Revision{},
	}
}

// UserPrivilegeSpec convert pb UserPrivilegeSpec to table UserPrivilegeSpec
// nolint:revive
func (up *UserPrivilegeSpec) UserPrivilegeSpec() *table.UserPrivilegeSpec {
	if up == nil {
		return nil
	}

	return &table.UserPrivilegeSpec{
		User:          up.User,
		PrivilegeType: table.PrivilegeType(up.PrivilegeType),
		ReadOnly:      up.ReadOnly,
	}
}

// UserPrivilegeAttachment convert pb UserPrivilegeAttachment to table UserPrivilegeAttachment
// nolint:revive
func (up *UserPrivilegeAttachment) UserPrivilegeAttachment() *table.UserPrivilegeAttachment {
	if up == nil {
		return nil
	}

	return &table.UserPrivilegeAttachment{
		BizID: up.BizId,
		AppID: up.AppId,
		Uid:   up.Uid,
	}
}

// PbUserPrivilegeSpec convert table UserPrivilegeSpec to pb UserPrivilegeSpec
// nolint:revive
func PbUserPrivilegeSpec(spec *table.UserPrivilegeSpec) *UserPrivilegeSpec {
	if spec == nil {
		return nil
	}

	return &UserPrivilegeSpec{
		User:          spec.User,
		PrivilegeType: string(spec.PrivilegeType),
		ReadOnly:      spec.ReadOnly,
	}
}

// PbUserPrivilegeAttachment convert table UserPrivilegeAttachment to pb UserPrivilegeAttachment
// nolint:revive
func PbUserPrivilegeAttachment(attachment *table.UserPrivilegeAttachment) *UserPrivilegeAttachment {
	if attachment == nil {
		return nil
	}
	return &UserPrivilegeAttachment{
		BizId: attachment.BizID,
		AppId: attachment.AppID,
		Uid:   attachment.Uid,
	}
}

// PbUserPrivilege convert table UserPrivilege to pb UserPrivilege
func PbUserPrivilege(up *table.UserPrivilege) *UserPrivilege {
	if up == nil {
		return nil
	}

	return &UserPrivilege{
		Id:         up.ID,
		Spec:       PbUserPrivilegeSpec(up.Spec),
		Attachment: PbUserPrivilegeAttachment(up.Attachment),
	}
}

// PbUserPrivileges convert table UserPrivilege to pb UserPrivilege
func PbUserPrivileges(up []*table.UserPrivilege) []*UserPrivilege {
	if up == nil {
		return make([]*UserPrivilege, 0)
	}
	result := make([]*UserPrivilege, 0)
	for _, v := range up {
		result = append(result, PbUserPrivilege(v))
	}
	return result
}
