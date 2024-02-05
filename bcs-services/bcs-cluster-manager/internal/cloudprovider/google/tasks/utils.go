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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	container "google.golang.org/api/container/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// updateNodeGroupCloudNodeGroupID set nodegroup cloudNodeGroupID
func updateNodeGroupCloudNodeGroupID(nodeGroupID string, newGroup *cmproto.NodeGroup) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	group.Region = newGroup.Region
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

func checkOperationStatus(computeCli *api.ComputeServiceClient, url, taskID string, d time.Duration) error {
	return loop.LoopDoFunc(context.Background(), func() error {
		o, err := api.GetOperation(computeCli, url)
		if err != nil {
			blog.Warnf("Error[%s] while getting operation %s on %s: %v", taskID, o.Name, o.TargetLink, err)
			return nil
		}
		blog.Infof("Operation[%s] [%s] %s status: %s", taskID, url, o.Name, o.Status)
		if o.Status == "DONE" {
			if o.Error != nil {
				errBytes, err := o.Error.MarshalJSON()
				if err != nil {
					errBytes = []byte(fmt.Sprintf("operation failed, but error couldn't be recovered: %v", err))
				}
				return fmt.Errorf("error while getting operation %s on %s: %s", o.Name, o.TargetLink, errBytes)
			}
			return loop.EndLoop
		}
		blog.Infof("taskID[%s] operation %s still running", taskID, o.SelfLink)

		return nil
	}, loop.LoopInterval(d))
}

func checkGKEOperationStatus(containerCli *api.ContainerServiceClient, operation *container.Operation,
	taskID string, d time.Duration) error {
	return loop.LoopDoFunc(context.Background(), func() error {

		o, err := containerCli.GetGKEOperation(context.Background(), operation.Name)
		if err != nil {
			return err
		}
		if o.Status == "DONE" {
			if o.Error != nil {
				return fmt.Errorf("%d, %v", o.Error.Code, o.Error.Details)
			}
			return loop.EndLoop
		}
		blog.Infof("taskID[%s] operation %s still running", taskID, o.SelfLink)
		return nil
	}, loop.LoopInterval(d))
}
