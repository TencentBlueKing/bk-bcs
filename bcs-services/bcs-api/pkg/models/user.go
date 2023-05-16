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
 *
 */

package models

import "time"

// BackendCredentials xxx
type BackendCredentials map[string]interface{}

// User is the internal user model for bke, when bke wants to be "connected" with other user systems like
// "blueking auth system", the external user credentials should always been transformed into internal bke user(and tokens).
type User struct {
	ID   uint   `gorm:"primary_key"`
	Name string `gorm:"unique;not null"`
	// A "super user" is a user who can create other users, this super user is often initialized by config file
	IsSuperUser bool
	CreatedAt   time.Time

	// These field startswith Backend* will be set by auth filters
	BackendType        string             `gorm:"-"`
	BackendCredentials BackendCredentials `gorm:"-"`
}

const (
	// UserTokenTypeSession xxx
	UserTokenTypeSession = iota + 1
	// UserTokenTypeKubeConfigForPaas xxx
	UserTokenTypeKubeConfigForPaas
	// UserTokenTypeKubeConfigPlain xxx
	UserTokenTypeKubeConfigPlain
)

// UserToken is the token which can be used by tools like kubectl to connect to Kubernetes clusers.
type UserToken struct {
	ID        uint
	UserId    uint
	Type      uint
	Value     string `gorm:"unique;size:64"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

// HasExpired mean that is this token has been expired
func (t *UserToken) HasExpired() bool {
	if time.Now().After(t.ExpiresAt) {
		return true
	}
	return false
}

const (
	// ExternalUserSourceTypeBCS xxx
	ExternalUserSourceTypeBCS = iota + 1
)

// ExternalUserRecord stores the replationship between [bke internal user] and [user from external provider]
type ExternalUserRecord struct {
	ID     uint
	UserId uint
	// There should be no duplicated external source/user pair in this table
	// SourceType is a string => type+userID is unique
	SourceType uint `gorm:"unique_index:idx_source_type_user_id"`
	// user_type + user_id => bke user
	SourceUserType string `gorm:"unique_index:idx_source_type_user_id"`
	SourceUserId   string `gorm:"unique_index:idx_source_type_user_id"`
	CreatedAt      time.Time
}
