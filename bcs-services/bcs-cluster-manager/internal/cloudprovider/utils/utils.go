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

package utils

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
)

// SyncClusterInfoToPassCC sync clusterInfo to pass-cc
func SyncClusterInfoToPassCC(taskID string, cluster *proto.Cluster) {
	err := passcc.GetCCClient().CreatePassCCCluster(cluster)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCCluster[%s] failed: %v",
			taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCCluster[%s] successful",
			taskID, cluster.ClusterID)
	}

	err = passcc.GetCCClient().CreatePassCCClusterSnapshoot(cluster)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCClusterSnapshoot[%s] failed: %v",
			taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCClusterSnapshoot[%s] successful",
			taskID, cluster.ClusterID)
	}
}

// SyncDeletePassCCCluster sync delete pass-cc cluster
func SyncDeletePassCCCluster(taskID string, cluster *proto.Cluster) {
	err := passcc.GetCCClient().DeletePassCCCluster(cluster.ProjectID, cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: DeletePassCCCluster[%s] failed: %v", taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("CleanClusterDBInfoTask[%s]: DeletePassCCCluster[%s] successful", taskID, cluster.ClusterID)
	}
}

func getResourceType(env string) string {
	if env == "prod" {
		return auth.ClusterProd
	}

	return auth.ClusterTest
}
