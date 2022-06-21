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

package account

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// CloudAccount cloudID and accountID
type CloudAccount struct {
	CloudID   string
	AccountID string
}

// getAllAccountsByCloudID list all accounts by cloud
func getAllAccountsByCloudID(ctx context.Context, model store.ClusterManagerModel, cloudID string) ([]proto.CloudAccount, error) {
	condM := make(operator.M)
	condM[account.CloudKey] = cloudID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	return model.ListCloudAccount(ctx, cond, &options.ListOption{})
}

func getRelativeClustersByAccountID(ctx context.Context, model store.ClusterManagerModel, account CloudAccount) ([]string, error) {
	condM := make(operator.M)
	condM["provider"] = account.CloudID
	condM["status"] = common.StatusRunning
	condM["cloudaccountid"] = account.AccountID

	clusterIDs := make([]string, 0)
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusters, err := model.ListCluster(ctx, cond, &options.ListOption{})
	if err != nil {
		return nil, err
	}

	for i := range clusters {
		clusterIDs = append(clusterIDs, clusters[i].ClusterID)
	}

	return clusterIDs, nil
}
