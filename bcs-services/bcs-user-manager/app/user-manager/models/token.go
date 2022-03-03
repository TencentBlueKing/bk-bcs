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

import "time"

const (
	CreatedBySystem = "system"
	CreatedBySync   = "sync"
)

// BcsTempToken is the temprary token, which is used to create by other client,
// and it can't be refreshed.
type BcsTempToken struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Username  string     `json:"username" gorm:"not null"`
	Token     string     `json:"token" gorm:"unique;size:64"`
	UserType  uint       `json:"user_type"` // normal user or admin
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// HasExpired mean that is temp token has been expired
func (t *BcsTempToken) HasExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
