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

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/autoscaler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/vcluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/component/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	ioptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

	installVclusterComponentStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.InstallVclusterAction,
		StepName:   "安装/更新vCluster组件",
	}
	uninstallVclusterComponentStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.DeleteVclusterAction,
		StepName:   "删除vCluster组件",
	}
)

// BuildWatchComponentTaskStep build common watch step
func BuildWatchComponentTaskStep(task *proto.Task, cls *proto.Cluster, namespace string) {
	watchStep := cloudprovider.InitTaskStep(installWatchComponentStep)

	watchStep.Params[cloudprovider.ProjectIDKey.String()] = cls.ProjectID
	watchStep.Params[cloudprovider.ClusterIDKey.String()] = cls.ClusterID
	watchStep.Params[cloudprovider.NamespaceKey.String()] = namespace

	task.Steps[installWatchComponentStep.StepMethod] = watchStep
	task.StepSequence = append(task.StepSequence, installWatchComponentStep.StepMethod)
}

// EnsureWatchComponentTask deploy bcs-k8s-watch task, if not exist, create it, if exist, update it
func EnsureWatchComponentTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start ensure watch component")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	// get auto scaling option
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	projectID := step.Params[cloudprovider.ProjectIDKey.String()]
	namespaceID := step.Params[cloudprovider.NamespaceKey.String()]

	// InstallWatchComponentByHelm install watch component but not handle error, need user to handle release
	err = InstallWatchComponentByHelm(ctx, projectID, clusterID, namespaceID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("ensure watch component failed [%s]", err))
		blog.Errorf("EnsureWatchComponentTask[%s] failed: %v", taskID, err)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"ensure watch component successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureWatchComponentTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// InstallWatchComponentByHelm deploy watch by helm
func InstallWatchComponentByHelm(ctx context.Context, projectID,
	clusterID, namespace string) error {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	bcsWatch := &watch.BcsWatch{
		ClusterID: clusterID,
	}
	values, err := bcsWatch.GetValues()
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] get bcsWatch[%s] failed: %v", taskID, clusterID, err)
		return err
	}
	blog.Infof("InstallWatchComponentByHelm[%s] get bcsWatchValues[%s] successful", taskID, values)

	// check cluster namespace and create namespace when not exist
	if namespace == "" {
		namespace = ioptions.GetGlobalCMOptions().ComponentDeploy.Watch.ReleaseNamespace
	}
	err = CreateClusterNamespace(ctx, clusterID, NamespaceDetail{
		Namespace: namespace,
	})
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] CreateClusterNamespace failed: %v", taskID, err)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"create cluster namespace successful")

	installer, err := watch.GetWatchInstaller(projectID, namespace)
	if err != nil {
		blog.Errorf("InstallWatchComponentByHelm[%s] GetWatchInstaller failed: %v", taskID, err)
		return err
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
func DeleteWatchComponentByHelm(ctx context.Context, projectID,
	clusterID, namespace string) error {
	traceID := cloudprovider.GetTaskIDFromContext(ctx)

	install, err := watch.GetWatchInstaller(projectID, namespace)
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

	err = loop.LoopDoFunc(timeContext, func() error {
		var exist bool
		exist, err = install.IsInstalled(clusterID)
		if err != nil {
			blog.Errorf("DeleteWatchComponentByHelm[%s] failed[%s:%s]: %v", traceID, projectID, clusterID, err)
			return nil
		}

		blog.Infof("DeleteWatchComponentByHelm[%s] watchRelease[%s] status[%v]", traceID, clusterID, exist)
		if !exist {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("DeleteWatchComponentByHelm[%s] watchRelease[%s] failed: %v", traceID, clusterID, err)
		return err
	}

	blog.Infof("DeleteWatchComponentByHelm[%s] successful[%s:%s]", traceID, projectID, clusterID)
	return nil
}

// install vcluster component

// BuildInstallVclusterTaskStep build common vcluster component
func BuildInstallVclusterTaskStep(task *proto.Task, clusterID, hostClusterID string) {
	installStep := cloudprovider.InitTaskStep(installVclusterComponentStep)

	installStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	installStep.Params[cloudprovider.HostClusterIDKey.String()] = hostClusterID

	task.Steps[installVclusterComponentStep.StepMethod] = installStep
	task.StepSequence = append(task.StepSequence, installVclusterComponentStep.StepMethod)
}

func getVClusterAndHostCluster(clusterID, hostClusterID string) (*proto.Cluster, *proto.Cluster, error) {
	var (
		vCluster, hostCluster *proto.Cluster
		err                   error
	)
	vCluster, err = cloudprovider.GetClusterByID(clusterID)
	if err != nil {
		return nil, nil, err
	}
	hostCluster, err = cloudprovider.GetClusterByID(hostClusterID)
	if err != nil {
		return nil, nil, err
	}

	return vCluster, hostCluster, nil
}

// InstallVclusterTask ensure auto scaler task, if not exist, create it, if exist, update it
func InstallVclusterTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start install vcluster")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get vcluster option
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	hostClusterID := step.Params[cloudprovider.HostClusterIDKey.String()]

	cluster, hostCluster, err := getVClusterAndHostCluster(clusterID, hostClusterID)
	if err != nil {
		blog.Errorf("InstallVclusterTask[%s]: get GetClusterByID for %s/%s failed", taskID, clusterID, hostClusterID)
		retErr := fmt.Errorf("GetClusterByID information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	vclusterData, err := buildVClusterInfoByVCluster(cluster, hostCluster)
	if err != nil {
		blog.Errorf("InstallVclusterTask[%s] buildVClusterInfoByVCluster for %s failed: %v", taskID, clusterID, err)
		retErr := fmt.Errorf("buildVClusterInfoByVCluster failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := ensureVclusterWithInstaller(ctx, vclusterData); err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("install vcluster failed [%s]", err))
		blog.Errorf("InstallVclusterTask[%s] for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("InstallVclusterTask failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"install vcluster successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("InstallVclusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func buildVClusterInfoByVCluster(cls *proto.Cluster, hostCls *proto.Cluster) (*VclusterInfo, error) {
	vclusterData := &VclusterInfo{
		CloudID:      cls.Provider,
		ClusterID:    cls.ClusterID,
		ClusterEnv:   cls.GetEnvironment(),
		SrcClusterID: cls.SystemID,
	}

	// hostCluster projectID
	vclusterData.ProjectID = hostCls.ProjectID

	// vcluster network type
	env := utils.GetValueFromMap(cls.ExtraInfo, common.VClusterNetworkKey)
	if !utils.StringInSlice(env, []string{utils.DEVNET.String(), utils.IDC.String()}) {
		return nil, fmt.Errorf("vcluster[%s] env error", cls.ClusterID)
	}
	vclusterData.Env = utils.EnvType(env)

	// vcluster namespace info in hostCluster
	namespaceStr := utils.GetValueFromMap(cls.ExtraInfo, common.VClusterNamespaceInfo)
	if len(namespaceStr) == 0 {
		return nil, fmt.Errorf("vcluster[%s] namespace empty", cls.ClusterID)
	}
	nsInfo := &proto.NamespaceInfo{}
	err := json.Unmarshal([]byte(namespaceStr), nsInfo)
	if err != nil {
		return nil, err
	}
	vclusterData.Namespace = nsInfo.Name

	// hostCluster serviceCIDR
	vclusterData.ServiceCIDR = func() string {
		if hostCls != nil && hostCls.NetworkSettings != nil {
			return hostCls.NetworkSettings.ServiceIPv4CIDR
		}

		return ""
	}()

	// need to generate tke etcd cluster if hostNetwork is idc
	if env == utils.IDC.String() {
		servers := utils.GetValueFromMap(cls.ExtraInfo, cloudprovider.VclusterEtcdServersKey.String())
		if len(servers) == 0 {
			return nil, fmt.Errorf("vcluster[%s] etcdServers empty", cls.ClusterID)
		}
		vclusterData.EtcdServers = servers

		ca := utils.GetValueFromMap(cls.ExtraInfo, cloudprovider.VclusterEtcdCAKey.String())
		if len(ca) == 0 {
			return nil, fmt.Errorf("vcluster[%s] etcdCA empty", cls.ClusterID)
		}
		vclusterData.EtcdCA = ca

		cert := utils.GetValueFromMap(cls.ExtraInfo, cloudprovider.VclusterEtcdClientCertKey.String())
		if len(cert) == 0 {
			return nil, fmt.Errorf("vcluster[%s] etcdClientCert empty", cls.ClusterID)
		}
		vclusterData.EtcdClientCert = cert

		key := utils.GetValueFromMap(cls.ExtraInfo, cloudprovider.VclusterEtcdClientKeyKey.String())
		if len(key) == 0 {
			return nil, fmt.Errorf("vcluster[%s] etcdClientKey empty", cls.ClusterID)
		}
		vclusterData.EtcdClientKey = key
	}

	return vclusterData, nil
}

// VclusterInfo xxx
type VclusterInfo struct {
	Env        utils.EnvType
	CloudID    string
	ClusterID  string
	ClusterEnv string

	// host cluster info
	ProjectID    string
	SrcClusterID string
	Namespace    string

	EtcdServers    string
	EtcdCA         string
	EtcdClientCert string
	EtcdClientKey  string

	ServiceCIDR string
}

func ensureVclusterWithInstaller(ctx context.Context, info *VclusterInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	installer, err := vcluster.GetVclusterInstaller(info.ProjectID, info.ClusterID, info.Namespace)
	if err != nil {
		blog.Errorf("ensureVclusterWithInstaller[%s] GetVclusterInstaller failed: %v", taskID, err)
		return err
	}
	installed, err := installer.IsInstalled(info.SrcClusterID)
	if err != nil {
		blog.Errorf("ensureVclusterWithInstaller[%s] IsInstalled failed: %v", taskID, err)
		return err
	}

	// 安装vcluster, 没有安装择安装; 安装，则更新
	vc := vcluster.Vcluster{
		Env:            info.Env,
		EtcdServers:    info.EtcdServers,
		EtcdCA:         info.EtcdCA,
		EtcdClientCert: info.EtcdClientCert,
		EtcdClientKey:  info.EtcdClientKey,
		ServiceCIDR:    info.ServiceCIDR,
		ClusterID:      info.ClusterID,
		ClusterEnv:     info.ClusterEnv,
	}
	values, err := vc.GetValues()
	if err != nil {
		return fmt.Errorf("transAutoScalingOptionToValues failed, err: %s", err)
	}

	// install or upgrade
	if installed {
		if errUpgrade := installer.Upgrade(info.SrcClusterID, values); errUpgrade != nil {
			return fmt.Errorf("upgrade app failed, err %s", errUpgrade)
		}
	} else {
		if errInstall := installer.Install(info.SrcClusterID, values); errInstall != nil {
			return fmt.Errorf("install app failed, err %s", errInstall)
		}
	}

	// check status
	ok, err := installer.CheckAppStatus(info.SrcClusterID, time.Minute*10, false)
	if err != nil {
		return fmt.Errorf("check app status failed, err %s", err)
	}
	if !ok {
		return fmt.Errorf("app install failed, err %s", err)
	}
	return nil
}

// BuildUnInstallVclusterTaskStep build common vcluster component
func BuildUnInstallVclusterTaskStep(task *proto.Task, clusterID, hostClusterID string) {
	installStep := cloudprovider.InitTaskStep(uninstallVclusterComponentStep)

	installStep.Params[cloudprovider.ClusterIDKey.String()] = clusterID
	installStep.Params[cloudprovider.HostClusterIDKey.String()] = hostClusterID

	task.Steps[uninstallVclusterComponentStep.StepMethod] = installStep
	task.StepSequence = append(task.StepSequence, uninstallVclusterComponentStep.StepMethod)
}

// UnInstallVclusterTask delete vcluster
func UnInstallVclusterTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start delete vcluster")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get vcluster option
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	hostClusterID := step.Params[cloudprovider.HostClusterIDKey.String()]

	cluster, hostCluster, err := getVClusterAndHostCluster(clusterID, hostClusterID)
	if err != nil {
		blog.Errorf("UnInstallVclusterTask[%s]: get GetClusterByID for %s/%s failed", taskID, clusterID, hostClusterID)
		retErr := fmt.Errorf("GetClusterByID information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	vclusterData, err := buildVClusterInfoByVCluster(cluster, hostCluster)
	if err != nil {
		blog.Errorf("UnInstallVclusterTask[%s] buildVClusterInfoByVCluster for %s failed: %v", taskID, clusterID, err)
		retErr := fmt.Errorf("UnInstallVclusterTask failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := DeleteVclusterComponentByHelm(ctx, vclusterData); err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete vcluster failed [%s]", err))
		blog.Errorf("UnInstallVclusterTask[%s] for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("UnInstallVclusterTask failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete vcluster successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UnInstallVclusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// DeleteVclusterComponentByHelm unInstall vcluster
func DeleteVclusterComponentByHelm(ctx context.Context, info *VclusterInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	install, err := vcluster.GetVclusterInstaller(info.ProjectID, info.ClusterID, info.Namespace)
	if err != nil {
		blog.Errorf("DeleteVclusterComponentByHelm[%s] GetVclusterInstaller failed: %v", taskID, err)
		return err
	}
	err = install.Uninstall(info.SrcClusterID)
	if err != nil {
		blog.Errorf("DeleteVclusterComponentByHelm[%s] Uninstall failed: %v", taskID, err)
		return err
	}

	// wait check delete component status
	timeContext, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	err = loop.LoopDoFunc(timeContext, func() error {
		exist, errInstall := install.IsInstalled(info.SrcClusterID)
		if errInstall != nil {
			blog.Errorf("DeleteVclusterComponentByHelm[%s] failed[%s:%s]: %v", taskID, info.ProjectID,
				info.SrcClusterID, errInstall)
			return nil
		}

		blog.Infof("DeleteVclusterComponentByHelm[%s] watchRelease[%s] status[%v]", taskID, info.SrcClusterID, exist)
		if !exist {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("DeleteVclusterComponentByHelm[%s] watchRelease[%s] failed: %v", taskID, info.SrcClusterID, err)
		return err
	}

	blog.Infof("DeleteVclusterComponentByHelm[%s] successful[%s:%s]", taskID, info.ProjectID, info.SrcClusterID)
	return nil
}

// install CA component

// BuildEnsureAutoScalerTaskStep build common autoScaler component
func BuildEnsureAutoScalerTaskStep(task *proto.Task, clusterID, cloudID string) {
	ensureStep := cloudprovider.InitTaskStep(ensureAutoScalerStep)

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
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start ensure auto scaler")
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
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	// get cluster nodegroup list
	nodegroupList, err := getClusterNodeGroups(clusterID)
	if err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s]: ListNodeGroup for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("ListNodeGroup failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if err := ensureAutoScalerWithInstaller(ctx, nodegroupList, asOption); err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("ensure auto scaler failed [%s]", err))
		blog.Errorf("EnsureAutoScalerTask[%s] for %s failed: %v", taskID, clusterID, err)
		retErr := fmt.Errorf("EnsureAutoScalerTask failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"ensure auto scaler successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func ensureAutoScalerWithInstaller(ctx context.Context, nodeGroups []proto.NodeGroup,
	as *proto.ClusterAutoScalingOption) error {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	installer, err := autoscaler.GetAutoScalerInstaller(as.ProjectID)
	if err != nil {
		blog.Errorf("ensureAutoScalerWithInstaller GetAutoScalerInstaller failed: %v", err)
		return err
	}

	// check cluster namespace and create namespace when not exist
	err = CreateClusterNamespace(ctx, as.ClusterID, NamespaceDetail{
		Namespace: ioptions.GetGlobalCMOptions().ComponentDeploy.AutoScaler.ReleaseNamespace,
	})
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

			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(),
				taskID, stepName, "upgrade app successful")
		} else {
			if err = installer.Install(as.ClusterID, values); err != nil {
				return fmt.Errorf("install app failed, err %s", err)
			}

			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(),
				taskID, stepName, "install app successful")
		}

		// check status
		var ok bool
		ok, err = installer.CheckAppStatus(as.ClusterID, time.Minute*10, false)
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
		ok, errCheck := installer.CheckAppStatus(as.ClusterID, time.Minute*10, false)
		if errCheck != nil {
			return fmt.Errorf("check app status failed, err %s", err)
		}
		if !ok {
			return fmt.Errorf("app install failed, err %s", err)
		}

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(),
			taskID, stepName, "uninstall app successful")
	}

	return nil
}
