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
	cli "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/install"
	cmoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

const (
	defaultReplicas = 1
)

func getInstaller(projectID string) install.Installer {
	op := cmoptions.GetGlobalCMOptions()
	client := cli.NewBCSAppClient(
		op.BCSAppConfig.Server,
		op.BCSAppConfig.AppCode,
		op.BCSAppConfig.AppSecret,
		op.BCSAppConfig.BkUserName,
		op.BCSAppConfig.Debug,
	)
	debug := false
	if !op.BCSAppConfig.Enable || op.BCSAppConfig.Debug {
		debug = true
	}
	return install.NewBKAPIInstaller(
		projectID,
		op.AutoScaler.ChartName,
		op.AutoScaler.ReleaseName,
		op.AutoScaler.ReleaseNamespace,
		op.AutoScaler.IsPublicRepo,
		client,
		debug,
	)
}

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

	// ensure
	if err := ensureAutoScalerWithInstaller(getInstaller(asOption.ProjectID), clusterID, nodegroupList,
		asOption); err != nil {
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

func ensureAutoScalerWithInstaller(installer install.Installer, clusterID string, nodeGroups []cmproto.NodeGroup,
	as *cmproto.ClusterAutoScalingOption) error {
	installed, err := installer.IsInstalled(clusterID)
	if err != nil {
		return err
	}

	if as == nil {
		return fmt.Errorf("cluster %s ClusterAutoScalingOption is nil", clusterID)
	}

	// 开了自动伸缩，但是没有安装，则安装
	// 开了自动伸缩，且安装了，则更新
	// 没有开自动伸缩，但是安装了，则卸载
	// 没有开自动伸缩，且没有安装，则不做处理

	// 开启了自动伸缩
	if as.EnableAutoscale {
		values, err := transAutoScalingOptionToValues(nodeGroups, *as, defaultReplicas)
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}
		if installed {
			return installer.Upgrade(clusterID, values)
		}
		return installer.Install(clusterID, values)
	}

	// 如果已经安装且关闭了自动伸缩，则卸载
	if installed {
		// 副本数设置为 0，则停止应用
		values, err := transAutoScalingOptionToValues(nodeGroups, *as, 0)
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}
		return installer.Upgrade(clusterID, values)
	}
	return nil
}
