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

// GetCluster get clusterInfo by clusterID
func GetCluster(clusterId string) *models.BcsCluster {
	cluster := models.BcsCluster{}
	GCoreDB.Where(&models.BcsCluster{ID: clusterId}).First(&cluster)
	if cluster.ID != "" {
		return &cluster
	}
	return nil
}

// CreateCluster create cluster
func CreateCluster(cluster *models.BcsCluster) error {
	err := GCoreDB.Create(cluster).Error
	return err
}
