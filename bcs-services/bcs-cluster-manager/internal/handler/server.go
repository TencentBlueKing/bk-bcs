/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ClusterManager server for cluster manager
type ClusterManager struct {
	kubeOp *clusterops.K8SOperator
	model  store.ClusterManagerModel
	locker lock.DistributedLock
}

// NewClusterManager create clustermanager Handler
func NewClusterManager(model store.ClusterManagerModel, kubeOp *clusterops.K8SOperator,
	locker lock.DistributedLock) *ClusterManager {
	return &ClusterManager{
		model:  model,
		kubeOp: kubeOp,
		locker: locker,
	}
}
