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

// UserPrivilege 用户权限
type UserPrivilege struct {
	ID         uint32                   `json:"id" gorm:"primaryKey"`
	Spec       *UserPrivilegeSpec       `json:"spec" gorm:"embedded"`
	Attachment *UserPrivilegeAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                `json:"revision" gorm:"embedded"`
}

// TableName is the user_privilege's database table name.
func (u *UserPrivilege) TableName() string {
	return "user_privileges"
}

// AppID AuditRes interface
func (u *UserPrivilege) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (u *UserPrivilege) ResID() uint32 {
	return u.ID
}

// ResType AuditRes interface
func (u *UserPrivilege) ResType() string {
	return "user_privileges"
}

// UserPrivilegeSpec 用户权限组
type UserPrivilegeSpec struct {
	User          string        `gorm:"column:user" json:"user"`
	PrivilegeType PrivilegeType `gorm:"column:privilege_type" json:"privilege_type"` // 权限类型：system、custom
	ReadOnly      bool          `gorm:"column:read_only" json:"read_only"`           // 只读
}

// UserPrivilegeAttachment xxx
type UserPrivilegeAttachment struct {
	BizID           uint32 `gorm:"column:biz_id" json:"biz_id"`
	AppID           uint32 `gorm:"column:app_id" json:"app_id"`
	Uid             uint32 `gorm:"column:uid" json:"uid"`
	TemplateSpaceID uint32 `gorm:"column:template_space_id" json:"template_space_id"`
}
