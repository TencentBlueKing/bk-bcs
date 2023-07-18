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

package common

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// resourcePool device label task: stepName and stepMethod
	resourcePoolLabelStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.ResourcePoolLabelAction,
		StepName:   "资源池设备设置标签",
	}
)

// BuildResourcePoolLabelTaskStep build resourcePool device task step
func BuildResourcePoolLabelTaskStep(task *proto.Task, clusterID string) {
	labelStep := cloudprovider.InitTaskStep(resourcePoolLabelStep)

	labelStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID

	task.Steps[resourcePoolLabelStep.StepMethod] = labelStep
	task.StepSequence = append(task.StepSequence, resourcePoolLabelStep.StepMethod)
}

// SetResourcePoolDeviceLabels set resourcePool device labels
func SetResourcePoolDeviceLabels(taskID, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("SetResourcePoolDeviceLabels[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("SetResourcePoolDeviceLabels[%s] task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	deviceList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams, cloudprovider.DeviceIDsKey.String(),
		",")

	if len(deviceList) == 0 {
		blog.Infof("SetResourcePoolDeviceLabels[%s] deviceList empty", taskID)
		_ = state.UpdateStepSucc(start, stepName)
		return nil
	}

	ctx := utils.WithTraceIDForContext(context.Background(), taskID)

	for i := range deviceList {
		device, err := tresource.GetResourceManagerClient().GetDeviceInfoByDeviceID(ctx, deviceList[i])
		if err != nil {
			blog.Errorf("SetResourcePoolDeviceLabels[%s] GetDeviceInfoByDeviceID[%s] failed: %v",
				taskID, deviceList[i], err)
			continue
		}

		if device.Labels == nil || len(device.Labels) == 0 {
			blog.Errorf("SetResourcePoolDeviceLabels[%s] GetDeviceInfoByDeviceID[%s] labels empty",
				taskID, deviceList[i])
			continue
		}

		specialLabels := func() map[string]string {
			labels := make(map[string]string, 0)
			for k, v := range device.Labels {
				if strings.Contains(k, utils.DeviceLabelFlag) {
					labels[k] = v
				}
			}
			return labels
		}()
		if len(specialLabels) == 0 {
			blog.Errorf("SetResourcePoolDeviceLabels[%s] specialLabels[%s] empty",
				taskID, deviceList[i])
			continue
		}

		err = UpdateClusterNodesLabels(ctx, NodeLabelsData{
			ClusterID: clusterID,
			NodeIPs:   []string{device.InnerIP},
			Labels:    specialLabels,
		})
		if err != nil {
			blog.Errorf("SetResourcePoolDeviceLabels[%s] UpdateClusterNodesLabels[%s] failed: %v",
				taskID, device.InnerIP, err)
			continue
		}
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("task %s %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}
