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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

// GetRole get bcsRole by roleName
func GetRole(RoleName string) *models.BcsRole {
	role := models.BcsRole{}
	GCoreDB.Where(&models.BcsRole{Name: RoleName}).First(&role)
	if role.Name != "" {
		return &role
	}
	return nil
}

// CreateRole create role
func CreateRole(role *models.BcsRole) error {
	err := GCoreDB.Create(role).Error
	return err
}

// GetUrrByCondition Query BcsUserResourceRole by condition
func GetUrrByCondition(cond *models.BcsUserResourceRole) *models.BcsUserResourceRole {
	urr := models.BcsUserResourceRole{}
	GCoreDB.Where(cond).First(&urr)
	if urr.ID != 0 {
		return &urr
	}
	return nil
}

// CreateUserResourceRole create user-resource-role
func CreateUserResourceRole(urr *models.BcsUserResourceRole) error {
	err := GCoreDB.Create(urr).Error
	return err
}

// DeleteUserResourceRole delete user-resource-role
func DeleteUserResourceRole(urr *models.BcsUserResourceRole) error {
	err := GCoreDB.Delete(urr).Error
	return err
}
