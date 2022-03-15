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

package sqlstore

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

const (
	// PlainUserExpiredTime expired after 24 hours
	PlainUserExpiredTime = 24 * time.Hour
	// AdminSaasUserExpiredTime this means never expired
	AdminSaasUserExpiredTime = 10 * 365 * 24 * time.Hour
)

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

// GetUserByCondition Query user by condition
func GetUserByCondition(cond *models.BcsUser) *models.BcsUser {
	user := models.BcsUser{}
	GCoreDB.Where(cond).First(&user)
	if user.ID != 0 {
		return &user
	}
	return nil
}

// CreateUser create new user
func CreateUser(user *models.BcsUser) error {
	err := GCoreDB.Create(user).Error
	return err
}

// UpdateUser update user information
func UpdateUser(user, updatedUser *models.BcsUser) error {
	err := GCoreDB.Model(user).Updates(*updatedUser).Error
	return err
}
