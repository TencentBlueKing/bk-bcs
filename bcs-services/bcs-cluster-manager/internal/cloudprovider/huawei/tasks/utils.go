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

package tasks

import (
	"context"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// updateNodeGroupCloudNodeGroupID set nodegroup cloudNodeGroupID
func updateNodeGroupCloudNodeGroupID(nodeGroupID string, newGroup *cmproto.NodeGroup) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = newGroup.AutoScaling.VpcID
	}
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}
