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

package sqlstore

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

const (
	// PlainUserExpiredTime expired after 24 hours
	PlainUserExpiredTime = 24 * time.Hour
	// AdminSaasUserExpiredTime this means never expired
	AdminSaasUserExpiredTime = 10 * 365 * 24 * time.Hour
)

// GetUserByCondition Query user by condition
func GetUserByCondition(cond *models.BcsUser) *models.BcsUser {
	start := time.Now()
	user := models.BcsUser{}
	GCoreDB.Where(cond).First(&user)
	if user.ID != 0 {
		return &user
	}
	metrics.ReportMysqlSlowQueryMetrics("GetUserByCondition", metrics.Query, metrics.SucStatus, start)
	return nil
}

// CreateUser create new user
func CreateUser(user *models.BcsUser) error {
	start := time.Now()
	err := GCoreDB.Create(user).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("CreateUser", metrics.Create, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("CreateUser", metrics.Create, metrics.SucStatus, start)
	return nil
}

// UpdateUser update user information
func UpdateUser(user, updatedUser *models.BcsUser) error {
	start := time.Now()
	err := GCoreDB.Model(user).Updates(*updatedUser).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("UpdateUser", metrics.Update, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("UpdateUser", metrics.Update, metrics.SucStatus, start)
	return nil
}
