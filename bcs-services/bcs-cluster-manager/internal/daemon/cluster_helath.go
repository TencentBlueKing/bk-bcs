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

func (d *Daemon) reportClusterHealthStatus(error chan<- error) {
	condCluster := operator.NewLeafCondition(operator.In, operator.M{
		"status": []string{common.StatusRunning, common.StatusConnectClusterFailed},
	})
	clusterList, err := d.model.ListCluster(d.ctx, condCluster, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("reportClusterHealthStatus ListCluster failed: %v", err)
		error <- err
		return
	}

	concurency := utils.NewRoutinePool(10)
	defer concurency.Close()

	for i := range clusterList {
		concurency.Add(1)
		go func(cls cmproto.Cluster) {
			defer concurency.Done()

			newCluster, errLocal := d.model.GetCluster(d.ctx, cls.GetClusterID())
			if errLocal != nil {
				blog.Errorf("reportClusterHealthStatus GetCluster failed: %v", errLocal)
				error <- errLocal
				return
			}
			if !utils.StringInSlice(newCluster.GetStatus(),
				[]string{common.StatusRunning, common.StatusConnectClusterFailed}) {
				blog.Errorf("reportClusterHealthStatus[%s] %v", newCluster.ClusterID, newCluster.GetStatus())
				return
			}

			k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), d.model)
			kubeCli, errLocal := k8sOperator.GetClusterClient(cls.ClusterID)
			if errLocal != nil {
				error <- errLocal
				return
			}
			_, err = kubeCli.Discovery().ServerVersion()
			if err != nil {
				if options.GetEditionInfo().IsCommunicationEdition() {
					_ = d.updateClusterStatus(cls.ClusterID, common.StatusConnectClusterFailed)
				}
				metrics.ReportCloudClusterHealthStatus(cls.Provider, cls.ClusterID, 0)
				error <- err
				return
			}
			
			if options.GetEditionInfo().IsCommunicationEdition() {
				_ = d.updateClusterStatus(cls.ClusterID, common.StatusRunning)
			}
			metrics.ReportCloudClusterHealthStatus(cls.Provider, cls.ClusterID, 1)
		}(clusterList[i])
	}

	concurency.Wait()
}

// nolint
func (d *Daemon) updateClusterStatus(clusterId, status string) error {
	cluster, err := d.model.GetCluster(d.ctx, clusterId)
	if err != nil {
		return err
	}
	if cluster.Status == status {
		return nil
	}
	cluster.Status = status

	return d.model.UpdateCluster(d.ctx, cluster)
}
