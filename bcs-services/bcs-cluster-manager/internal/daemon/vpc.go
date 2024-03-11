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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrmanager"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func (d *Daemon) reportVpcAvailableIPCount(error chan<- error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{"available": "true"})
	cloudVPCs, err := d.model.ListCloudVPC(d.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		error <- err
		return
	}

	cidrCli, conClose, err := cidrmanager.GetCidrClient().GetCidrManagerClient()
	if err != nil || cidrCli == nil {
		errMsg := fmt.Errorf("GetCidrManagerClient failed: %v", err)
		error <- errMsg
		return
	}
	defer func() {
		if conClose != nil {
			conClose()
		}
	}()

	concurency := utils.NewRoutinePool(5)
	defer concurency.Close()

	for i := range cloudVPCs {
		concurency.Add(1)
		go func(vpc cmproto.CloudVPC) {
			defer concurency.Done()
			resp, errGet := cidrCli.GetVPCIPSurplus(d.ctx, &cidrmanager.GetVPCIPSurplusRequest{
				Region:   vpc.Region,
				CidrType: utils.GlobalRouter.String(),
				VpcID:    vpc.VpcID,
			})
			if errGet != nil {
				error <- errGet
				return
			}
			if resp.Code != 0 {
				error <- fmt.Errorf("GetVPCIPSurplus failed: %v", resp.Message)
				return
			}
			metrics.ReportCloudVpcAvailableIPNum(vpc.CloudID, vpc.VpcID, float64(resp.Data.IPSurplus))
		}(cloudVPCs[i])
	}

	concurency.Wait()
}
