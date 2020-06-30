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
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
)

func CreateBCSClusterInfo(external *m.BCSClusterInfo) error {
	err := GCoreDB.Create(external).Error
	return err
}

// Query BCSClusterInfo search a BCSClusterInfo object using given conditions
func QueryBCSClusterInfo(info *m.BCSClusterInfo) *m.BCSClusterInfo {
	result := m.BCSClusterInfo{}
	GCoreDB.Where(info).First(&result)
	if result.ID != 0 {
		return &result
	}
	return nil

}

// GetClusterByBCSInfo query for the cluster by given clusterId
func GetClusterByBCSInfo(sourceProjectID string, sourceClusterID string) *m.Cluster {
	var externalClusterInfo *m.BCSClusterInfo
	if sourceProjectID != "" {
		externalClusterInfo = QueryBCSClusterInfo(&m.BCSClusterInfo{
			SourceProjectId: sourceProjectID,
			SourceClusterId: sourceClusterID,
		})
	} else {
		externalClusterInfo = QueryBCSClusterInfo(&m.BCSClusterInfo{
			SourceClusterId: sourceClusterID,
		})
	}

	if externalClusterInfo == nil {
		return nil
	}

	cluster := m.Cluster{}
	GCoreDB.Where(&m.Cluster{ID: externalClusterInfo.ClusterId}).First(&cluster)
	if cluster.ID != "" {
		return &cluster
	}
	return nil
}
