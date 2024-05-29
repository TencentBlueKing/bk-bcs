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

package business

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
)

// FilterClusterInstanceFromNodesIDs nodeIDs existInCluster or notExistInCluster
func FilterClusterInstanceFromNodesIDs(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs []string) ([]string, []string, error) {
	var (
		nodes             []model.Node
		existInCluster    = make([]string, 0)
		notExistInCluster = make([]string, 0)
	)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return nil, nil, err
	}

	err = retry.Do(func() error {
		nodes, err = client.ListClusterNodes(info.Cluster.SystemID)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))

	for _, id := range nodeIDs {
		exit := false
		for _, node := range nodes {
			if *node.Metadata.Uid == id {
				exit = true
				existInCluster = append(existInCluster, id)
			}
		}
		if !exit {
			notExistInCluster = append(notExistInCluster, id)
		}
	}

	return existInCluster, notExistInCluster, nil
}
