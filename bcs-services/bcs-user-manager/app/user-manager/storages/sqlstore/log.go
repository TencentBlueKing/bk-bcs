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

// ListOperationLogByClusterID get operationLogs by clusterID
func ListOperationLogByClusterID(clusterID string) []models.BcsOperationLog {
	var operationLogs []models.BcsOperationLog
	GCoreDB.Where("cluster_id = ?", clusterID).Find(&operationLogs)
	return operationLogs
}

// ListOperationLogByUserClusterID get operationLogs by user and clusterID
func ListOperationLogByUserClusterID(clusterID string, user string) []models.BcsOperationLog {
	var operationLogs []models.BcsOperationLog
	GCoreDB.Where("cluster_id = ? AND op_user = ?", clusterID, user).Find(&operationLogs)
	return operationLogs
}

// CreateOperationLog create operation log
func CreateOperationLog(log *models.BcsOperationLog) error {
	err := GCoreDB.Create(log).Error
	return err
}

// DeleteOperationLogByTime delete operationLogs between start and end time
func DeleteOperationLogByTime(start time.Time, end time.Time) error {
	err := GCoreDB.Where("created_at BETWEEN ? AND ?", start, end).Delete(&models.BcsOperationLog{}).Error

	return err
}

