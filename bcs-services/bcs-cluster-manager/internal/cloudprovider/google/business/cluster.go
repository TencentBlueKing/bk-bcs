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

// Package business xxx
package business

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// GetCloudClusterVer get cloud cluster version
func GetCloudClusterVer(ctx context.Context, opt *cloudprovider.GetClusterOption) (string, error) {
	client, err := api.NewContainerServiceClient(&opt.CommonOption)
	if err != nil {
		return "", fmt.Errorf("create google client failed, err %s", err.Error())
	}
	cloudCluster, err := client.GetCluster(ctx, opt.Cluster.SystemID)
	if err != nil {
		return "", fmt.Errorf("list google cluster failed, err %s", err.Error())
	}

	clusterVer := strings.Split(cloudCluster.CurrentMasterVersion, ".")
	if len(clusterVer) > 1 {
		return clusterVer[0] + "." + clusterVer[1], nil
	}

	return "", nil
}

// GetClusterTerminationDate get cluster termination date
func GetClusterTerminationDate(ctx context.Context, cloudID, clusterVer string) (string, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"cloudid":  cloudID,
		"version":  clusterVer,
		"moduleid": "upgrade",
	})

	flags, err := cloudprovider.GetStorageModel().ListCloudModuleFlag(ctx, cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("GetNodesNumWhenApplyInstanceTask failed: %v", err)
		return "", err
	}

	if len(flags) > 0 && len(flags[0].FlagValueList) > 0 {
		return flags[0].FlagValueList[0], nil
	}

	return "", nil
}
