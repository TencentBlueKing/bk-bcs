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
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	tencentCloud = "tencentCloud"

	platform = "platform"
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
		// filter cluster
		if clusterList[i].ClusterType == common.ClusterTypeVirtual {
			continue
		}
		if clusterList[i].SystemID == "" {
			continue
		}

		concurency.Add(1)
		go func(cls *cmproto.Cluster) {
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
				blog.Errorf("reportClusterHealthStatus GetClusterClient failed: %v", errLocal)
				error <- errLocal
				return
			}
			_, errLocal = kubeCli.Discovery().ServerVersion()
			if errLocal != nil {
				blog.Errorf("reportClusterHealthStatus GetClusterClient failed: %v", errLocal)
				// if options.GetEditionInfo().IsCommunicationEdition() {}
				_ = d.updateClusterStatus(cls.ClusterID, common.StatusConnectClusterFailed)

				metrics.ReportCloudClusterHealthStatus(cls.Provider, cls.ClusterID, 0)
				error <- errLocal
				return
			}

			// if options.GetEditionInfo().IsCommunicationEdition() {}
			_ = d.updateClusterStatus(cls.ClusterID, common.StatusRunning)

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

	if cluster.Status == common.StatusDeleted {
		return nil
	}

	if cluster.Status == status {
		return nil
	}
	cluster.Status = status

	return d.model.UpdateCluster(d.ctx, cluster)
}

// ConnectToCluster connect to cluster
func ConnectToCluster(model store.ClusterManagerModel, clusterId string) bool {
	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), model)
	kubeCli, errLocal := k8sOperator.GetClusterClient(clusterId)
	if errLocal != nil {
		blog.Errorf("ConnectToCluster[%s] failed: %v", clusterId, errLocal)
		return false
	}
	_, errLocal = kubeCli.Discovery().ServerVersion()
	if errLocal != nil {
		blog.Errorf("ConnectToCluster[%s] failed: %v", clusterId, errLocal)
		return false
	}

	return true
}

// caStat 按 provider 维度统计 CA 使用率/开启率
type caStat struct {
	used, total           int
	debugUsed, debugTotal int
	prodUsed, prodTotal   int

	enabled, debugEnabled, prodEnabled int
}

func (d *Daemon) reportClusterCaUsageRatio(error chan<- error) {
	statusCond := operator.NewLeafCondition(operator.In, operator.M{
		"status": []string{common.StatusRunning, common.StatusConnectClusterFailed},
	})

	clusterList, err := d.model.ListCluster(d.ctx, statusCond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("reportClusterCaUsageRatio ListCluster failed: %v", err)
		error <- err
		return
	}

	var (
		used, total           int
		debugUsed, debugTotal int
		prodUsed, prodTotal   int

		enabled, debugEnabled, prodEnabled int

		providerStats = make(map[string]*caStat)
	)

	for i := range clusterList {
		// filter cluster
		if clusterList[i].ClusterType == common.ClusterTypeVirtual {
			continue
		}
		if clusterList[i].SystemID == "" {
			continue
		}
		if !ConnectToCluster(d.model, clusterList[i].ClusterID) {
			blog.Errorf("reportClusterCaUsageRatio[%s] ConnectToCluster failed", clusterList[i].ClusterID)
			continue
		}

		provider := clusterList[i].GetProvider()
		if providerStats[provider] == nil {
			providerStats[provider] = &caStat{}
		}
		stat := providerStats[provider]

		total++
		stat.total++
		switch clusterList[i].Environment {
		case common.Debug:
			debugTotal++
			stat.debugTotal++
		case common.Prod:
			prodTotal++
			stat.prodTotal++
		default:
		}

		condGroup := operator.NewLeafCondition(operator.Eq, operator.M{
			"clusterid": clusterList[i].GetClusterID(),
		})
		groupList, errLocal := d.model.ListNodeGroup(d.ctx, condGroup, &storeopt.ListOption{All: true})
		if errLocal != nil {
			blog.Errorf("reportClusterCaUsageRatio[%s] ListNodeGroup failed: %v", clusterList[i].ClusterID, err)
			continue
		}

		// 接入节点池 & 开启弹性伸缩
		if len(groupList) > 0 {
			used++
			stat.used++
			switch clusterList[i].Environment {
			case common.Debug:
				debugUsed++
				stat.debugUsed++
			case common.Prod:
				prodUsed++
				stat.prodUsed++
			default:
			}

			asOption, errLocal := d.model.GetAutoScalingOption(context.Background(), clusterList[i].ClusterID)
			if errLocal != nil {
				blog.Errorf("reportClusterCaUsageRatio[%s] GetAutoScalingOption failed: %v",
					clusterList[i].ClusterID, err)
				continue
			}
			if asOption.GetEnableAutoscale() {
				enabled++
				stat.enabled++

				switch clusterList[i].Environment {
				case common.Debug:
					debugEnabled++
					stat.debugEnabled++
				case common.Prod:
					prodEnabled++
					stat.prodEnabled++
				default:
				}
			}
		}
	}

	metrics.ReportCaUsageRatio(platform, platform, float64(used)/float64(total))
	metrics.ReportCaUsageRatio(common.Debug, platform, float64(debugUsed)/float64(debugTotal))
	metrics.ReportCaUsageRatio(common.Prod, platform, float64(prodUsed)/float64(prodTotal))

	metrics.ReportCaEnableRatio(platform, platform, float64(enabled)/float64(used))
	metrics.ReportCaEnableRatio(common.Debug, platform, float64(debugEnabled)/float64(debugUsed))
	metrics.ReportCaEnableRatio(common.Prod, platform, float64(prodEnabled)/float64(prodUsed))

	for provider, stat := range providerStats {
		metrics.ReportCaUsageRatio(platform, provider, float64(stat.used)/float64(stat.total))
		metrics.ReportCaUsageRatio(common.Debug, provider, float64(stat.debugUsed)/float64(stat.debugTotal))
		metrics.ReportCaUsageRatio(common.Prod, provider, float64(stat.prodUsed)/float64(stat.prodTotal))

		metrics.ReportCaEnableRatio(platform, provider, float64(stat.enabled)/float64(stat.used))
		metrics.ReportCaEnableRatio(common.Debug, provider, float64(stat.debugEnabled)/float64(stat.debugUsed))
		metrics.ReportCaEnableRatio(common.Prod, provider, float64(stat.prodEnabled)/float64(stat.prodUsed))
	}
}
