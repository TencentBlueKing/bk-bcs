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

package common

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// CreateClusterNamespace for cluster create namespace
func CreateClusterNamespace(ctx context.Context, clusterID, ns string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clusterID) == 0 || len(ns) == 0 {
		blog.Errorf("CreateClusterNamespace[%s] resource empty")
		return fmt.Errorf("cluster or ns resource empty")
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	err := k8sOperator.CreateNamespace(ctx, clusterID, ns)
	if err != nil {
		blog.Errorf("CreateClusterNamespace[%s] resource[%s:%s] failed: %v", taskID, clusterID, ns, err)
		return err
	}

	blog.Infof("CreateClusterNamespace[%s] success[%s:%s]", taskID, clusterID, ns)

	return nil
}
