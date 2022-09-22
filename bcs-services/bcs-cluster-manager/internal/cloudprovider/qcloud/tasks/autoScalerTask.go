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

package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/autoscaler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

const (
	defaultReplicas = 1
)

// EnsureAutoScalerTask ensure auto scaler task, if not exist, create it, if exist, update it
func EnsureAutoScalerTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get auto scaling option
	clusterID := step.Params["ClusterID"]
	asOption, err := cloudprovider.GetStorageModel().GetAutoScalingOption(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s]: get autoscalingoption for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get autoscalingoption information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get cluster nodegroup list
	cond := &operator.Condition{
		Op:    operator.Eq,
		Value: operator.M{"clusterid": clusterID, "enableautoscale": true},
	}
	nodegroupList, err := cloudprovider.GetStorageModel().ListNodeGroup(context.Background(), cond, &options.ListOption{
		All: true})
	if err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s]: ListNodeGroup for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("ListNodeGroup failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := ensureAutoScalerWithInstaller(nodegroupList, asOption); err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s] for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("EnsureAutoScalerTask failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func ensureAutoScalerWithInstaller(nodeGroups []cmproto.NodeGroup, as *cmproto.ClusterAutoScalingOption) error {
	installer, err := autoscaler.GetAutoScalerInstaller(as.ProjectID)
	if err != nil {
		blog.Errorf("ensureAutoScalerWithInstaller GetAutoScalerInstaller failed: %v", err)
		return err
	}

	installed, err := installer.IsInstalled(as.ClusterID)
	if err != nil {
		blog.Errorf("ensureAutoScalerWithInstaller IsInstalled failed: %v", err)
		return err
	}

	// 开了自动伸缩，但是没有安装，则安装
	// 开了自动伸缩，且安装了，则更新
	// 没有开自动伸缩，但是安装了，则卸载
	// 没有开自动伸缩，且没有安装，则不做处理

	scaler := autoscaler.AutoScaler{
		NodeGroups:        nodeGroups,
		AutoScalingOption: as,
	}
	// 开启了自动伸缩
	if as.EnableAutoscale {
		scaler.Replicas = defaultReplicas

		values, err := scaler.GetValues()
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}

		// install or upgrade
		if installed {
			if err := installer.Upgrade(as.ClusterID, values); err != nil {
				return fmt.Errorf("upgrade app failed, err %s", err)
			}
		} else {
			if err := installer.Install(as.ClusterID, values); err != nil {
				return fmt.Errorf("install app failed, err %s", err)
			}
		}

		// check status
		ok, err := installer.CheckAppStatus(as.ClusterID, time.Minute*10)
		if err != nil {
			return fmt.Errorf("check app status failed, err %s", err)
		}
		if !ok {
			return fmt.Errorf("app install failed, err %s", err)
		}
		return nil
	}

	// 如果已经安装且关闭了自动伸缩，则卸载
	if installed {
		// 副本数设置为 0，则停止应用
		scaler.Replicas = 0

		values, err := scaler.GetValues()
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}

		if err := installer.Upgrade(as.ClusterID, values); err != nil {
			return fmt.Errorf("upgrade app failed, err %s", err)
		}
		// check status
		ok, err := installer.CheckAppStatus(as.ClusterID, time.Minute*10)
		if err != nil {
			return fmt.Errorf("check app status failed, err %s", err)
		}
		if !ok {
			return fmt.Errorf("app install failed, err %s", err)
		}
	}

	return nil
}
