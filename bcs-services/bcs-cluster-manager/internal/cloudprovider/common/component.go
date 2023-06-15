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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/autoscaler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	ioptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

var (
	installWatchComponentStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.WatchTask,
		StepName:   "安装集群watch组件",
	}

	ensureAutoScalerStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.EnsureAutoScalerAction,
		StepName:   "安装/更新CA组件",
	}
)

// BuildWatchComponentTaskStep build common watch step
func BuildWatchComponentTaskStep(task *proto.Task, cls *proto.Cluster) {
	watchStep := cloudprovider.InitTaskStep(installWatchComponentStep)

	watchStep.Params[cloudprovider.ProjectIDKey.String()] = cls.ProjectID
	watchStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID

	task.Steps[installWatchComponentStep.StepMethod] = watchStep
	task.StepSequence = append(task.StepSequence, installWatchComponentStep.StepMethod)
}

// EnsureWatchComponentTask deploy bcs-k8s-watch task, if not exist, create it, if exist, update it
func EnsureWatchComponentTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// get auto scaling option
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	projectID := step.Params[cloudprovider.ProjectIDKey.String()]

	// InstallWatchComponentByHelm install watch component but not handle error, need user to handle release
	err = InstallWatchComponentByHelm(ctx, projectID, clusterID)
	if err != nil {
		blog.Errorf("EnsureWatchComponentTask[%s] failed: %v", taskID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureWatchComponentTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// InstallWatchComponentByHelm deploy watch service by helm
func InstallWatchComponentByHelm(ctx context.Context, projectID, clusterID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	bcsWatch := &watch.BcsWatch{
		ClusterID: clusterID,
	}
	values, err := bcsWatch.GetValues()
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] get bcsWatch[%s] failed: %v", taskID, clusterID, err)
		return err
	}
	blog.Infof("InstallWatchComponentByHelm[%s] get bcsWatchValues[%s] successful", taskID, values)

	installer, err := watch.GetWatchInstaller(projectID)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] GetWatchInstaller failed: %v", taskID, err)
		return err
	}

	// check cluster namespace and create namespace when not exist
	err = CreateClusterNamespace(ctx, clusterID, ioptions.GetGlobalCMOptions().ComponentDeploy.Watch.ReleaseNamespace)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] CreateClusterNamespace failed: %v", taskID, err)
	}

	err = installer.Install(clusterID, values)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] Install failed: %v", taskID, err)
		return err
	}

	blog.Infof("InstallWatchComponentByHelm[%s] successful[%s:%s]", taskID, projectID, clusterID)
	return nil
}

// DeleteWatchComponentByHelm unInstall watch
func DeleteWatchComponentByHelm(ctx context.Context, projectID, clusterID string) error {
	traceID := cloudprovider.GetTaskIDFromContext(ctx)

	install, err := watch.GetWatchInstaller(projectID)
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] GetWatchInstaller failed: %v", traceID, err)
		return err
	}
	err = install.Uninstall(clusterID)
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] Uninstall failed: %v", traceID, err)
		return err
	}
	// wait check delete component status
	timeContext, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	err = cloudprovider.LoopDoFunc(timeContext, func() error {
		var exist bool
		exist, err = install.IsInstalled(clusterID)
		if err != nil {
			blog.Errorf("DeleteWatchComponentByHelm[%s] failed[%s:%s]: %v", traceID, projectID, clusterID, err)
			return nil
		}

		blog.Infof("DeleteWatchComponentByHelm[%s] watchRelease[%s] status[%v]", traceID, clusterID, exist)
		if !exist {
			return cloudprovider.EndLoop
		}

		return nil
	}, cloudprovider.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] watchRelease[%s] failed: %v", traceID, clusterID, err)
		return err
	}

	blog.Infof("DeleteWatchComponentByHelm[%s] successful[%s:%s]", traceID, projectID, clusterID)
	return nil
}

// install CA component

// BuildEnsureAutoScalerTaskStep build common autoScaler component
func BuildEnsureAutoScalerTaskStep(task *proto.Task, stepName, clusterID, cloudID string) {
	ensureStep := cloudprovider.InitTaskStep(ensureAutoScalerStep, cloudprovider.WithStepTaskName(stepName))

	ensureStep.Params[cloudprovider.CloudIDKey.String()] = cloudID
	ensureStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID

	task.Steps[ensureAutoScalerStep.StepMethod] = ensureStep
	task.StepSequence = append(task.StepSequence, ensureAutoScalerStep.StepMethod)
}

const (
	defaultReplicas = 1
)

func getClusterNodeGroups(clusterID string) ([]proto.NodeGroup, error) {
	// get cluster nodegroup list
	cond := &operator.Condition{
		Op: operator.Eq,
		Value: operator.M{
			"clusterid":       clusterID,
			"enableautoscale": true,
		},
	}
	nodegroupList, err := cloudprovider.GetStorageModel().ListNodeGroup(context.Background(), cond, &options.ListOption{
		All: true})
	if err != nil {
		return nil, fmt.Errorf("getClusterNodeGroups ListNodeGroup failed: %v", err)
	}

	// filter status deleting node group
	filterGroups := make([]proto.NodeGroup, 0)
	for _, group := range nodegroupList {
		if group.Status == common.StatusDeleteNodeGroupDeleting {
			continue
		}
		filterGroups = append(filterGroups, group)
	}

	return filterGroups, nil
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
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	asOption, err := cloudprovider.GetStorageModel().GetAutoScalingOption(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s]: get autoscalingoption for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get autoscalingoption information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// get cluster nodegroup list
	nodegroupList, err := getClusterNodeGroups(clusterID)
	if err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s]: ListNodeGroup for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("ListNodeGroup failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := ensureAutoScalerWithInstaller(ctx, nodegroupList, asOption); err != nil {
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

func ensureAutoScalerWithInstaller(ctx context.Context, nodeGroups []proto.NodeGroup, as *proto.ClusterAutoScalingOption) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	installer, err := autoscaler.GetAutoScalerInstaller(as.ProjectID)
	if err != nil {
		blog.Errorf("ensureAutoScalerWithInstaller GetAutoScalerInstaller failed: %v", err)
		return err
	}

	// check cluster namespace and create namespace when not exist
	err = CreateClusterNamespace(ctx, as.ClusterID, ioptions.GetGlobalCMOptions().ComponentDeploy.AutoScaler.ReleaseNamespace)
	if err != nil {
		blog.Errorf("ensureAutoScalerWithInstaller[%s] CreateClusterNamespace failed: %v", taskID, err)
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

		var values string
		// 注意: 配置打开弹性伸缩的节点池
		values, err = scaler.GetValues()
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}
		// install or upgrade
		if installed {
			if err = installer.Upgrade(as.ClusterID, values); err != nil {
				return fmt.Errorf("upgrade app failed, err %s", err)
			}
		} else {
			if err = installer.Install(as.ClusterID, values); err != nil {
				return fmt.Errorf("install app failed, err %s", err)
			}
		}

		// check status
		var ok bool
		ok, err = installer.CheckAppStatus(as.ClusterID, time.Minute*10)
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

		var values string
		values, err = scaler.GetValues()
		if err != nil {
			return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
		}

		if err = installer.Upgrade(as.ClusterID, values); err != nil {
			return fmt.Errorf("upgrade app failed, err %s", err)
		}
		// check status
		ok, errCheck := installer.CheckAppStatus(as.ClusterID, time.Minute*10)
		if errCheck != nil {
			return fmt.Errorf("check app status failed, err %s", err)
		}
		if !ok {
			return fmt.Errorf("app install failed, err %s", err)
		}
	}

	return nil
}
