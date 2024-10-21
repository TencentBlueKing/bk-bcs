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

package types

// UserPrivilege 用户权限
type UserPrivilege struct {
	AppID           uint32
	TemplateSpaceID uint32
	User            string
	Uid             uint32
}

// UserGroupPrivilege 用户组权限
type UserGroupPrivilege struct {
	AppID           uint32
	TemplateSpaceID uint32
	UserGroup       string
	Gid             uint32
}

// FileGroupPrivilege 文件权限
type FileGroupPrivilege struct {
	Name            string
	Path            string
	TemplateSpaceID uint32
	Uid             uint32
	User            string
	Gid             uint32
	UserGroup       string
}
