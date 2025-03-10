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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	remain = "remain"
	total  = "total"
	ratio  = "ratio"
)

func getIpUsageByVpc(model store.ClusterManagerModel, ipType string, vpc *cmproto.CloudVPC) (uint32, uint32, error) {
	cloud, err := actions.GetCloudByCloudID(model, vpc.CloudID)
	if err != nil {
		blog.Errorf("getIpUsageByVpc[%s:%s] failed: %v", vpc.Region, vpc.VpcID, err)
		return 0, 0, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud: cloud,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when getIpUsageByVpc[%s:%s] failed, %s",
			cloud.CloudID, cloud.CloudProvider, vpc.Region, vpc.VpcID, err.Error(),
		)
		return 0, 0, err
	}
	cmOption.Region = vpc.Region

	vpcMgr, err := cloudprovider.GetVPCMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider[%s] vpcManager[%s:%s] for getIpUsageByVpc failed, %s",
			cloud.CloudProvider, vpc.Region, vpc.VpcID, err.Error(),
		)
		return 0, 0, err
	}

	return vpcMgr.GetVpcIpUsage(vpc.VpcID, ipType, nil, cmOption)
}

func getIpUsageByCluster(model store.ClusterManagerModel, ipType string, cluster *cmproto.Cluster) (uint32,
	uint32, error) {
	cloud, err := actions.GetCloudByCloudID(model, cluster.GetProvider())
	if err != nil {
		blog.Errorf("getIpUsageByCluster[%s:%s] failed: %v", cluster.Region, cluster.GetClusterID(), err)
		return 0, 0, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud: cloud,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when getIpUsageByCluster[%s:%s] failed, %s",
			cloud.CloudID, cloud.CloudProvider, cluster.Region, cluster.GetClusterID(), err.Error(),
		)
		return 0, 0, err
	}
	cmOption.Region = cluster.Region

	vpcMgr, err := cloudprovider.GetVPCMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider[%s] vpcManager[%s:%s] for getIpUsageByCluster failed, %s",
			cloud.CloudProvider, cluster.Region, cluster.GetClusterID(), err.Error(),
		)
		return 0, 0, err
	}

	return vpcMgr.GetClusterIpUsage(cluster.GetClusterID(), ipType, cmOption)
}

func (d *Daemon) reportVpcIpResourceUsage(error chan<- error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{"available": "true"})
	cloudVPCs, err := d.model.ListCloudVPC(d.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		error <- err
		return
	}

	concurency := utils.NewRoutinePool(5)
	defer concurency.Close()

	for i := range cloudVPCs {
		concurency.Add(1)
		go func(vpc *cmproto.CloudVPC) {
			defer concurency.Done()

			overlayTotal, overlaySurplus, errGet := getIpUsageByVpc(d.model, common.ClusterOverlayNetwork, vpc)
			if errGet != nil {
				error <- errGet
				return
			}
			underlayTotal, underlaySurplus, errGet := getIpUsageByVpc(d.model, common.ClusterUnderlayNetwork, vpc)
			if errGet != nil {
				error <- errGet
				return
			}

			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterOverlayNetwork,
				remain, float64(overlaySurplus))
			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterOverlayNetwork,
				total, float64(overlayTotal))
			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterOverlayNetwork,
				ratio, float64(overlayTotal-overlaySurplus)/float64(overlayTotal))

			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterUnderlayNetwork,
				remain, float64(underlaySurplus))
			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterUnderlayNetwork,
				total, float64(underlayTotal))
			metrics.ReportCloudVpcResourceUsage(vpc.CloudID, vpc.Region, vpc.VpcID, common.ClusterUnderlayNetwork,
				ratio, float64(underlayTotal-underlaySurplus)/float64(underlayTotal))

		}(cloudVPCs[i])
	}

	concurency.Wait()
}

func (d *Daemon) reportClusterVpcUsage(error chan<- error) {
	statusCond := operator.NewLeafCondition(operator.In, operator.M{
		"status": []string{common.StatusRunning, common.StatusConnectClusterFailed},
	})
	providerCond := operator.NewLeafCondition(operator.Eq, operator.M{"provider": tencentCloud})
	cond := operator.NewBranchCondition(operator.And, statusCond, providerCond)

	clusterList, err := d.model.ListCluster(d.ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("reportClusterVpcUsage ListCluster failed: %v", err)
		error <- err
		return
	}

	for i := range clusterList {
		if clusterList[i].ClusterType == common.ClusterTypeVirtual {
			continue
		}
		if clusterList[i].SystemID == "" {
			continue
		}

		overlayTotal, overlaySurplus, errGet := getIpUsageByCluster(d.model,
			common.ClusterOverlayNetwork, clusterList[i])
		if errGet != nil {
			error <- errGet
			continue
		}

		metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
			clusterList[i].GetClusterID(), common.ClusterOverlayNetwork, remain, float64(overlaySurplus))
		metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
			clusterList[i].GetClusterID(), common.ClusterOverlayNetwork, total, float64(overlayTotal))
		metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
			clusterList[i].GetClusterID(), common.ClusterOverlayNetwork,
			ratio, float64(overlayTotal-overlaySurplus)/float64(overlayTotal))

		if clusterList[i].GetNetworkSettings().GetEnableVPCCni() {
			underlayTotal, underlaySurplus, errGet := getIpUsageByCluster(d.model,
				common.ClusterUnderlayNetwork, clusterList[i])
			if errGet != nil {
				error <- errGet
				blog.Errorf("reportClusterVpcUsage[%s:%s] getIpUsageByCluster failed: %v",
					clusterList[i].Region, clusterList[i].GetClusterID(), errGet)
				continue
			}

			metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
				clusterList[i].GetClusterID(), common.ClusterUnderlayNetwork, remain, float64(underlaySurplus))
			metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
				clusterList[i].GetClusterID(), common.ClusterUnderlayNetwork, total, float64(underlayTotal))
			metrics.ReportClusterVpcResourceUsage(clusterList[i].GetProvider(), clusterList[i].GetBusinessID(),
				clusterList[i].GetClusterID(), common.ClusterUnderlayNetwork,
				ratio, float64(underlayTotal-underlaySurplus)/float64(underlayTotal))

			zoneSubs, _, _, errGet := business.GetClusterCurrentVpcCniSubnets(clusterList[i], false)
			if errGet != nil {
				error <- errGet
				blog.Errorf("reportClusterVpcUsage[%s:%s] GetClusterCurrentVpcCniSubnets failed: %v",
					clusterList[i].Region, clusterList[i].GetClusterID(), errGet)
				continue
			}
			for zone, sub := range zoneSubs {
				metrics.ReportClusterVpcCniSubnetResourceUsage(clusterList[i].GetProvider(), clusterList[i].BusinessID,
					clusterList[i].GetClusterID(), remain, zone, float64(sub.AvailableIps))
				metrics.ReportClusterVpcCniSubnetResourceUsage(clusterList[i].GetProvider(), clusterList[i].BusinessID,
					clusterList[i].GetClusterID(), total, zone, float64(sub.TotalIps))
				metrics.ReportClusterVpcCniSubnetResourceUsage(clusterList[i].GetProvider(), clusterList[i].BusinessID,
					clusterList[i].GetClusterID(), ratio, zone, sub.Ratio)
			}
		}
	}
}
