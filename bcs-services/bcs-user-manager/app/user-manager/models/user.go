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

import (
	"time"
)

// BcsUser user table
type BcsUser struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Name      string     `json:"name" gorm:"not null"`
	UserType  uint       `json:"user_type"`
	UserToken string     `json:"user_token" gorm:"unique;size:64"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`                       // 用户创建时间
	UpdatedAt time.Time  `json:"updated_at"`                       // user-token刷新时间
	ExpiresAt time.Time  `json:"expires_at"`                       // user-token过期时间
	DeletedAt *time.Time `json:"deleted_at" gorm:"type:timestamp"` // user-token删除时间
}

// HasExpired mean that is this token has been expired
func (t *BcsUser) HasExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
