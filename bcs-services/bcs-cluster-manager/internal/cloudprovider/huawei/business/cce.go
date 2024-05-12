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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
)

func DeleteClusterInstance(client *api.CceClient, clusterID string, nodes []model.Node) ([]string, error) {
	success := make([]string, 0)
	for _, node := range nodes {
		_, err := client.DeleteNode(clusterID, *node.Metadata.Uid, true)
		if err != nil {
			continue
		}

		success = append(success, *node.Metadata.Uid)
	}

	return success, nil
}
