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

package nodegroup

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// filterNodeGroupOption options
type filterNodeGroupOption struct {
	Name      string
	ClusterID string
	ProjectID string
	Region    string
	ListOpt   *storeopt.ListOption
}

// listNodeGroupByConds list node groups
func listNodeGroupByConds(model store.ClusterManagerModel, options filterNodeGroupOption) (
	[]*cmproto.NodeGroup, int64, error) {
	var (
		groupList = make([]*cmproto.NodeGroup, 0)
		err       error
		condM     = make(operator.M)
	)

	if len(options.Name) != 0 {
		condM["name"] = options.Name
	}
	if len(options.ClusterID) != 0 {
		condM["clusterid"] = options.ClusterID
	}
	if len(options.ProjectID) != 0 {
		condM["projectid"] = options.ProjectID
	}
	if len(options.Region) != 0 {
		condM["region"] = options.Region
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
	branchCond := operator.NewBranchCondition(operator.And, cond, condStatus)

	count, err := model.CountNodeGroup(context.Background(), branchCond)
	if err != nil {
		return nil, 0, err
	}

	if options.ListOpt == nil {
		options.ListOpt = &storeopt.ListOption{}
	}
	groups, err := model.ListNodeGroup(context.Background(), branchCond, options.ListOpt)
	if err != nil {
		return nil, 0, err
	}
	for i := range groups {
		groupList = append(groupList, removeSensitiveInfo(groups[i]))
	}
	return groupList, count, nil
}

func virtualNodeID() string {
	return "bcs-" + utils.RandomHexString(8)
}

func checkNodeGroupResourceValidate(ctx context.Context, provider string, nodeGroup *cmproto.NodeGroup,
	operation string, scaleUpResource uint32) error {
	ngr, err := cloudprovider.GetNodeGroupMgr(provider)
	if err != nil {
		return err
	}

	return ngr.CheckResourcePoolQuota(ctx, nodeGroup, operation, scaleUpResource)
}
