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

// Package daemon xxx
package daemon

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func (d *Daemon) reportClusterGroupNodeNum(error chan<- error) {
	condGroup := operator.NewLeafCondition(operator.Eq, operator.M{
		"status": common.StatusRunning,
	})
	groupList, err := d.model.ListNodeGroup(d.ctx, condGroup, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("reportClusterGroupNodeNum ListNodeGroup failed: %v", err)
		error <- err
		return
	}

	concurency := utils.NewRoutinePool(10)
	defer concurency.Close()

	for i := range groupList {
		concurency.Add(1)
		go func(group *cmproto.NodeGroup) {
			defer concurency.Done()

			if group.LaunchTemplate == nil || group.LaunchTemplate.InstanceType == "" {
				return
			}
			if group.AutoScaling == nil {
				return
			}

			bizId := ""
			pInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(group.ProjectID, true)
			if errLocal == nil {
				bizId = pInfo.GetBusinessID()
			}

			metrics.ReportClusterGroupAvailableNodeNum(group.ClusterID, group.NodeGroupID,
				group.LaunchTemplate.InstanceType, bizId, float64(group.AutoScaling.DesiredSize))
			metrics.ReportClusterGroupMaxNodeNum(group.ClusterID, group.NodeGroupID,
				group.LaunchTemplate.InstanceType, bizId, float64(group.AutoScaling.MaxSize))
		}(groupList[i])
	}

	concurency.Wait()
}
