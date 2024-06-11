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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func getAvailableIPNumByVpc(model store.ClusterManagerModel, ipType string, vpc cmproto.CloudVPC) (uint32, error) {
	cloud, err := actions.GetCloudByCloudID(model, vpc.CloudID)
	if err != nil {
		blog.Errorf("getAvailableIPNumByVpc[%s:%s] failed: %v", vpc.Region, vpc.VpcID, err)
		return 0, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud: cloud,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when getAvailableIPNumByVpc[%s:%s] failed, %s",
			cloud.CloudID, cloud.CloudProvider, vpc.Region, vpc.VpcID, err.Error(),
		)
		return 0, err
	}
	cmOption.Region = vpc.Region

	vpcMgr, err := cloudprovider.GetVPCMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider[%s] vpcManager[%s:%s] for getAvailableIPNumByVpc failed, %s",
			cloud.CloudProvider, vpc.Region, vpc.VpcID, err.Error(),
		)
		return 0, err
	}

	return vpcMgr.GetVpcIpSurplus(vpc.VpcID, ipType, nil, cmOption)
}

func (d *Daemon) reportVpcAvailableIPCount(error chan<- error) {
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
		go func(vpc cmproto.CloudVPC) {
			defer concurency.Done()

			overlayCnt, errGet := getAvailableIPNumByVpc(d.model, common.ClusterOverlayNetwork, vpc)
			if errGet != nil {
				error <- errGet
				return
			}

			metrics.ReportCloudVpcAvailableIPNum(vpc.CloudID, vpc.VpcID, float64(overlayCnt))
		}(cloudVPCs[i])
	}

	concurency.Wait()
}
