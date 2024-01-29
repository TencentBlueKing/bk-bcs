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

package models

import (
	"strings"
	"time"
)

// BcsUser user table
type BcsUser struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Name      string     `json:"name" gorm:"not null"`
	UserType  uint       `json:"user_type"`
	UserToken string     `json:"user_token" gorm:"unique;size:128"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamp null;default:null"` // 用户创建时间
	UpdatedAt time.Time  `json:"updated_at" gorm:"type:timestamp null;default:null"` // user-token刷新时间
	ExpiresAt time.Time  `json:"expires_at" gorm:"type:timestamp null;default:null"` // user-token过期时间
	DeletedAt *time.Time `json:"deleted_at" gorm:"type:timestamp null;default:null"` // user-token删除时间
}

const (
	// AdminUser definition
	AdminUser = iota + 1
	// SaasUser definition
	SaasUser
	// PlainUser definition
	PlainUser
	// ClientUser define jwt client user
	ClientUser
)

// HasExpired mean that is this token has been expired
func (t *BcsUser) HasExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsClient check if this user is client
func (t *BcsUser) IsClient() bool {
	return t.UserType != PlainUser
}

// IsAdmin check if this user is admin
func (t *BcsUser) IsAdmin() bool {
	return t.UserType == AdminUser || t.UserType == SaasUser
}

// Userinfo response user info
type Userinfo struct {
	UserName  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
}

// BcsClientUser client user table 平台账号
type BcsClientUser struct {
	ProjectCode   string    `json:"project_code" gorm:"not null"`
	Name          string    `json:"name" gorm:"not null"`
	UserType      uint      `json:"user_type"`
	UserToken     string    `json:"user_token" gorm:"unique;size:128"`
	CreatedBy     string    `json:"created_by"`
	Manager       string    `json:"manager"`                                            // 管理员，逗号分隔
	AuthorityUser string    `json:"authority_user"`                                     // 授权用户，逗号分隔
	ExpiresAt     time.Time `json:"expires_at" gorm:"type:timestamp null;default:null"` // user-token过期时间
	CreatedAt     time.Time `json:"created_at" gorm:"type:timestamp null;default:null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"type:timestamp null;default:null"`
}

// BcsClient client user table 平台账号
type BcsClient struct {
	ID            uint       `json:"id" gorm:"primary_key"`
	ProjectCode   string     `json:"project_code" gorm:"not null"`
	Name          string     `json:"name" gorm:"not null"`
	Manager       *string    `json:"manager"`        // 管理员，逗号分隔
	AuthorityUser *string    `json:"authority_user"` // 授权用户，逗号分隔
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at" gorm:"type:timestamp null;default:null"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"type:timestamp null;default:null"`
	DeletedAt     *time.Time `json:"deleted_at" gorm:"type:timestamp null;default:null"`
}

// HasExpired mean that is this token has been expired
func (t *BcsClientUser) HasExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// ManagerList get manager list
func (t *BcsClientUser) ManagerList() []string {
	managers := []string{}
	if t.Manager != "" {
		managers = strings.Split(t.Manager, ",")
	}
	return managers
}

// IsManager check if this user is manager
func (t *BcsClientUser) IsManager(user string) bool {
	for _, v := range t.ManagerList() {
		if v == user {
			return true
		}
	}
	return false
}

// AuthorityUserList get authority user list
func (t *BcsClientUser) AuthorityUserList() []string {
	authorityUsers := []string{}
	if t.AuthorityUser != "" {
		authorityUsers = strings.Split(t.AuthorityUser, ",")
	}
	return authorityUsers
}

// IsAuthorityUser check if this user is authority user
func (t *BcsClientUser) IsAuthorityUser(user string) bool {
	for _, v := range t.AuthorityUserList() {
		if v == user {
			return true
		}
	}
	return false
}

// ManagerList get manager list
func (t *BcsClient) ManagerList() []string {
	managers := []string{}
	if t.Manager != nil && *t.Manager != "" {
		managers = strings.Split(*t.Manager, ",")
	}
	return managers
}

// IsManager check if this user is manager
func (t *BcsClient) IsManager(user string) bool {
	for _, v := range t.ManagerList() {
		if v == user {
			return true
		}
	}
	return false
}

// AuthorityUserList get authority user list
func (t *BcsClient) AuthorityUserList() []string {
	authorityUsers := []string{}
	if t.AuthorityUser != nil && *t.AuthorityUser != "" {
		authorityUsers = strings.Split(*t.AuthorityUser, ",")
	}
	return authorityUsers
}

// IsAuthorityUser check if this user is authority user
func (t *BcsClient) IsAuthorityUser(user string) bool {
	for _, v := range t.AuthorityUserList() {
		if v == user {
			return true
		}
	}
	return false
}
