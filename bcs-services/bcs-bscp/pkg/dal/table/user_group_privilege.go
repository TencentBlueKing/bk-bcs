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

package table

import (
	"fmt"
)

// UserGroupPrivilege 用户组权限
type UserGroupPrivilege struct {
	ID         uint32                        `json:"id" gorm:"primaryKey"`
	Spec       *UserGroupPrivilegeSpec       `json:"spec" gorm:"embedded"`
	Attachment *UserGroupPrivilegeAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                     `json:"revision" gorm:"embedded"`
}

// TableName is the user_group_privilege's database table name.
func (u *UserGroupPrivilege) TableName() string {
	return "user_group_privileges"
}

// AppID AuditRes interface
func (u *UserGroupPrivilege) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (u *UserGroupPrivilege) ResID() uint32 {
	return u.ID
}

// ResType AuditRes interface
func (u *UserGroupPrivilege) ResType() string {
	return "user_group_privileges"
}

// UserGroupPrivilegeSpec 用户权限组
type UserGroupPrivilegeSpec struct {
	UserGroup     string        `gorm:"column:user_group" json:"user_group"`
	PrivilegeType PrivilegeType `gorm:"column:privilege_type" json:"privilege_type"` // 权限类型：system、custom
	ReadOnly      bool          `gorm:"column:read_only" json:"read_only"`           // 只读
}

// UserGroupPrivilegeAttachment 用户组附加结构体
type UserGroupPrivilegeAttachment struct {
	BizID           uint32 `gorm:"column:biz_id" json:"biz_id"`
	AppID           uint32 `gorm:"column:app_id" json:"app_id"`
	Gid             uint32 `gorm:"column:gid" json:"gid"`
	TemplateSpaceID uint32 `gorm:"column:template_space_id" json:"template_space_id"`
}

// PrivilegeType defines the type of a system or a custom.
type PrivilegeType string

const (
	// PrivilegeTypeSystem is the type of system privilege
	PrivilegeTypeSystem PrivilegeType = "system"

	// PrivilegeTypeCustom is the type for custom privilege
	PrivilegeTypeCustom PrivilegeType = "custom"
)

// Validate the privilege type is valid or not.
func (pt PrivilegeType) Validate() error {
	switch pt {
	case PrivilegeTypeSystem:
	case PrivilegeTypeCustom:
	default:
		return fmt.Errorf("unknown %s privilege type", pt)
	}

	return nil
}
