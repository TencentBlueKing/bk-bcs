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

package daemon

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	qcloudProvider = "tencentCloud"
)

func (d *Daemon) reportClusterHealthStatus(error chan<- error) {
	condCluster := operator.NewLeafCondition(operator.Eq, operator.M{
		"provider": qcloudProvider,
		"status":   common.StatusRunning,
	})
	clusterList, err := d.model.ListCluster(d.ctx, condCluster, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("reportClusterHealthStatus ListCluster failed: %v", err)
		error <- err
	}

	concurency := utils.NewRoutinePool(10)
	defer concurency.Close()

	for i := range clusterList {
		concurency.Add(1)
		go func(cls cmproto.Cluster) {
			defer concurency.Done()

			k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), d.model)
			kubeCli, err := k8sOperator.GetClusterClient(cls.ClusterID)
			if err != nil {
				error <- err
				return
			}
			_, err = kubeCli.Discovery().ServerVersion()
			if err != nil {
				metrics.ReportCloudClusterHealthStatus(cls.Provider, cls.ClusterID, 0)
				error <- err
				return
			}
			metrics.ReportCloudClusterHealthStatus(cls.Provider, cls.ClusterID, 1)
		}(clusterList[i])
	}

	concurency.Wait()
}
