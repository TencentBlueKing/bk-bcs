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

package daemon

import (
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	provider_common "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	azureCloud                       = "azureCloud"
	removeAzureClusterPlatformTaints = "removeAzureClusterPlatformTaints"
)

func (d *Daemon) removeAzureClusterPlatformTaints(error chan<- error) {
	d.lock.Lock(removeAzureClusterPlatformTaints, []lock.LockOption{lock.LockTTL(time.Second * 10)}...) // nolint
	defer d.lock.Unlock(removeAzureClusterPlatformTaints)                                               // nolint

	defer utils.RecoverPrintStack("removeAzureClusterPlatformTaints")

	conds := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.In, operator.M{
			"status": []string{common.StatusRunning, common.StatusConnectClusterFailed},
		}),
		operator.NewLeafCondition(operator.Eq, operator.M{
			"provider": azureCloud,
		}),
	)

	azureClusterList, err := d.model.ListCluster(d.ctx, conds, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("removeAzureClusterPlatformTaints ListCluster failed: %v", err)
		error <- err
		return
	}

	concurency := utils.NewRoutinePool(30)
	defer concurency.Close()

	for i := range azureClusterList {
		concurency.Add(1)
		go func(cls cmproto.Cluster) {
			defer concurency.Done()

			connect := ConnectToCluster(d.model, cls.ClusterID)
			if !connect {
				errMsg := fmt.Sprintf("removeAzureClusterPlatformTaints ConnectToCluster[%s] failed", cls.ClusterID)
				blog.Errorf(errMsg)
				error <- errors.New(errMsg)
				return
			}

			// get cluster node list
			k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), d.model)
			nodes, errLocal := k8sOperator.ListClusterNodes(d.ctx, cls.ClusterID)
			if errLocal != nil {
				blog.Errorf("removeAzureClusterPlatformTaints ListClusterNodes failed: %v", errLocal)
				error <- errLocal
				return
			}

			// 过滤存在数据库 且 状态是RUNNING 的节点
			nodeNames := make([]string, 0)
			for _, n := range nodes {
				daoNode, errGet := d.model.GetNodeByName(d.ctx, cls.ClusterID, n.GetName())
				if errGet != nil {
					blog.Errorf("removeAzureClusterPlatformTaints GetNodeByName[%s:%s] failed: %v",
						cls.ClusterID, n.GetName(), errGet)
					continue
				}

				if daoNode.Status == common.StatusRunning {
					nodeNames = append(nodeNames, n.GetName())
				}
			}

			blog.Infof("removeAzureClusterPlatformTaints nodeName[%v]", nodeNames)

			// 移除平台taint
			ctx := cloudprovider.WithTaskIDForContext(d.ctx,
				fmt.Sprintf("%s-%s", removeAzureClusterPlatformTaints, cls.GetClusterID()))
			_ = provider_common.RemoveClusterNodesTaint(ctx, cls.GetClusterID(),
				cls.GetProvider(), nodeNames, nil)

		}(azureClusterList[i])
	}
	concurency.Wait()
}
