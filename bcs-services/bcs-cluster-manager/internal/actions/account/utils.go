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

package account

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	spb "google.golang.org/protobuf/types/known/structpb"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
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
func getAllAccountsByCloudID(
	ctx context.Context, model store.ClusterManagerModel, cloudID string) ([]*proto.CloudAccount, error) {
	condM := make(operator.M)
	condM[account.CloudKey] = cloudID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	return model.ListCloudAccount(ctx, cond, &options.ListOption{})
}

func getRelativeClustersByAccountID(
	ctx context.Context, model store.ClusterManagerModel, account CloudAccount) ([]string, error) {
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

// GetProjectAccountsV3Perm get iam v3 perm
func GetProjectAccountsV3Perm(user actions.PermInfo, accountList []string) (map[string]*spb.Struct, error) {
	var (
		v3Perm map[string]map[string]interface{}
		err    error
	)

	v3Perm, err = getUserAccountPermList(user, accountList)
	if err != nil {
		blog.Errorf("GetProjectAccountsV3Perm failed: %v", err.Error())
		return nil, err
	}

	// trans result for adapt front
	v3ResultPerm := make(map[string]*spb.Struct)
	for accountID := range v3Perm {
		actionPerm, err := spb.NewStruct(v3Perm[accountID])
		if err != nil {
			return nil, err
		}

		v3ResultPerm[accountID] = actionPerm
	}

	return v3ResultPerm, nil
}

func getUserAccountPermList(user actions.PermInfo, accountList []string) (map[string]map[string]interface{}, error) {
	permissions := make(map[string]map[string]interface{})

	accountPerm, err := auth.GetCloudAccountIamClient(user.TenantID)
	if err != nil {
		return nil, err
	}

	actionIDs := []string{cloudaccount.AccountUse.String(), cloudaccount.AccountManage.String()}

	perms, err := accountPerm.GetMultiAccountMultiActionPerm(user.UserID, user.ProjectID, accountList, actionIDs)
	if err != nil {
		return nil, err
	}

	for accountID, perm := range perms {
		if permissions[accountID] == nil {
			permissions[accountID] = make(map[string]interface{})
		}
		for action, res := range perm {
			permissions[accountID][action] = res
		}
	}

	return permissions, nil

}
