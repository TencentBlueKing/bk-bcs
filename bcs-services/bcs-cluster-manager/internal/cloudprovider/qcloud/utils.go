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

package qcloud

import (
	"fmt"
	"strconv"
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	cloudName = "qcloud"
)

const (
	defaultRegion = "ap-nanjing"
)

// qcloud taskName
const (
	// createClusterTaskTemplate bk-sops add task template
	createClusterTaskTemplate = "tke-create cluster: %s"
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "tke-delete cluster: %s"
	// createVirtualClusterTaskTemplate bk-sops add task template
	createVirtualClusterTaskTemplate = "tke-create virtual cluster: %s"
	// deleteVirtualClusterTaskTemplate bk-sops add task template
	deleteVirtualClusterTaskTemplate = "tke-delete virtual cluster: %s"
	// tkeAddNodeTaskTemplate bk-sops add task template
	tkeAddNodeTaskTemplate = "tke-add node: %s"
	// tkeCleanNodeTaskTemplate bk-sops add task template
	tkeCleanNodeTaskTemplate = "tke-remove node: %s"
	// tkeAddExternalNodeTaskTemplate bk-sops add task template
	tkeAddExternalNodeTaskTemplate = "tke-add external node: %s"
	// tkeCleanExternalNodeTaskTemplate bk-sops add task template
	tkeCleanExternalNodeTaskTemplate = "tke-remove external node: %s"
	// importClusterTaskTemplate bk-sops add task template
	importClusterTaskTemplate = "tke-import cluster: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "tke-create node group: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "tke-delete node group: %s/%s"
	// updateNodeGroupTaskTemplate bk-sops add task template
	updateNodeGroupTaskTemplate = "tke-update node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "tke-update node group desired node: %s/%s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "tke-remove node group nodes: %s/%s"
	// moveNodesToNodeGroupTaskTemplate bk-sops add task template
	moveNodesToNodeGroupTaskTemplate = "tke-move nodes to node group: %s/%s" // nolint
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "tke-switch node group auto scaling: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "tke-update auto scaling option: %s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "tke-switch auto scaling option status: %s"
)

// step task name&method
var (
	// import cluster task
	importClusterNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ImportClusterNodesTask", cloudName),
		StepName:   "导入集群节点",
	}
	registerClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName),
		StepName:   "注册集群kubeConfig认证",
	}
	installGSEAgentStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-InstallGSEAgentTask", cloudName),
		StepName:   "节点安装agent插件",
	}

	// create cluster task
	createClusterShieldAlarmStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateClusterShieldAlarmTask", cloudName),
		StepName:   "屏蔽机器告警",
	}
	createTKEClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateTKEClusterTask", cloudName),
		StepName:   "创建集群",
	}
	checkTKEClusterStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckTKEClusterStatusTask", cloudName),
		StepName:   "检测集群状态",
	}
	checkCreateClusterNodeStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCreateClusterNodeStatusTask", cloudName),
		StepName:   "检测集群节点状态",
	}
	registerManageClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RegisterManageClusterKubeConfigTask", cloudName),
		StepName:   "注册集群连接信息",
	}
	enableTkeClusterVpcCniStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-EnableTkeClusterVpcCniTask", cloudName),
		StepName:   "开启VPC-CNI网络模式",
	}
	updateCreateClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateCreateClusterDBInfoTask", cloudName),
		StepName:   "更新任务状态",
	}

	// delete cluster task
	deleteTKEClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteTKEClusterTask", cloudName),
		StepName:   "删除集群",
	}
	cleanClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName),
		StepName:   "清理集群数据",
	}

	// add node to cluster
	modifyInstancesVpcStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ModifyInstancesVpcTask", cloudName),
		StepName:   "节点转vpc",
	}
	checkInstanceStateStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckInstanceStateTask", cloudName),
		StepName:   "节点转vpc状态检测",
	}
	addNodesShieldAlarmStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AddNodesShieldAlarmTask", cloudName),
		StepName:   "屏蔽机器告警",
	}
	addNodesToClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AddNodesToClusterTask", cloudName),
		StepName:   "集群上架节点",
	}
	checkAddNodesStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckAddNodesStatusTask", cloudName),
		StepName:   "检测集群节点状态",
	}
	updateAddNodeDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateAddNodeDBInfoTask", cloudName),
		StepName:   "更新节点数据",
	}

	// remove node from cluster
	removeNodesFromClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RemoveNodesFromClusterTask", cloudName),
		StepName:   "删除节点",
	}
	updateRemoveNodeDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateRemoveNodeDBInfoTask", cloudName),
		StepName:   "清理节点数据",
	}

	// create nodeGroup task
	createCloudNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName),
		StepName:   "创建云节点组",
	}
	checkCloudNodeGroupStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName),
		StepName:   "检测云节点组状态",
	}
	updateCreateNodeGroupDBInfoStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-UpdateCreateNodeGroupDBInfoTask", cloudName),
		StepName:   "更新节点组数据",
	}

	// delete nodeGroup task
	deleteNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName),
		StepName:   "删除云节点组",
	}
	uninstallAutoScalerStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-UninstallAutoScalerTask", cloudName),
		StepName:   "卸载节点组自动扩缩容配置",
	}
	updateDeleteNodeGroupDBInfoStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName),
		StepName:   "清理节点组数据",
	}

	// clean node in nodeGroup task
	cleanNodeGroupNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName),
		StepName:   "下架节点组节点",
	}
	checkClusterCleanNodsStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckClusterCleanNodsTask", cloudName),
		StepName:   "检测下架节点状态",
	}
	returnIDCNodeToResourcePoolStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ReturnIDCNodeToResourcePoolTask", cloudName),
		StepName:   "下架第三方节点",
	}

	checkCleanNodeGroupNodesStatusStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-CheckCleanNodeGroupNodesStatusTask", cloudName),
		StepName:   "检查节点组状态",
	}
	updateCleanNodeGroupNodesDBInfoStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName),
		StepName:   "更新节点组数据",
	}

	// update desired nodes task
	applyInstanceMachinesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask),
		StepName:   "申请节点任务",
	}
	applyExternalNodeMachinesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyExternalNodeMachinesTask),
		StepName:   "申请IDC节点任务",
	}
	checkClusterNodesStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName),
		StepName:   "检测节点状态",
	}
	updateDesiredNodesDBInfoStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-UpdateDesiredNodesDBInfoTask", cloudName),
		StepName:   "清理节点数据",
	}

	// auto scale task
	deleteAutoScalerStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-DeleteAutoScalerTask", cloudName),
		StepName:   "删除CA组件",
	}
	updateNodeGroupAutoScalingDBStep = cloudprovider.StepInfo{ // nolint
		StepMethod: fmt.Sprintf("%s-UpdateNodeGroupAutoScalingDBTask", cloudName),
		StepName:   "更新CA组件状态",
	}

	// move nodes to nodeGroup task

	// add external nodes to cluster
	getExternalNodeScriptStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-GetExternalNodeScriptTask", cloudName),
		StepName:   "获取添加第三方节点脚本",
	}

	// remove external nodes from cluster
	removeExternalNodesFromClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RemoveExternalNodesFromClusterTask", cloudName),
		StepName:   "删除第三方节点",
	}
)

// CreateClusterTaskOption 创建集群构建step子任务
type CreateClusterTaskOption struct {
	Cluster      *proto.Cluster
	Nodes        []string
	NodeTemplate *proto.NodeTemplate
}

// BuildShieldAlertStep 屏蔽告警任务
func (cn *CreateClusterTaskOption) BuildShieldAlertStep(task *proto.Task) {
	shieldStep := cloudprovider.InitTaskStep(createClusterShieldAlarmStep)
	shieldStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	shieldStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.Nodes, ",")

	task.Steps[createClusterShieldAlarmStep.StepMethod] = shieldStep
	task.StepSequence = append(task.StepSequence, createClusterShieldAlarmStep.StepMethod)
}

// BuildCreateClusterStep 创建集群任务
func (cn *CreateClusterTaskOption) BuildCreateClusterStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createTKEClusterStep)
	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	createStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.Nodes, ",")
	if cn.NodeTemplate != nil {
		createStep.Params[cloudprovider.NodeTemplateIDKey.String()] = cn.NodeTemplate.NodeTemplateID
	}

	task.Steps[createTKEClusterStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createTKEClusterStep.StepMethod)
}

// BuildCheckClusterStatusStep 检测集群状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkTKEClusterStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkTKEClusterStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkTKEClusterStatusStep.StepMethod)
}

// BuildCheckClusterNodesStatusStep 检测创建集群节点状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterNodesStatusStep(task *proto.Task) {
	if len(cn.Nodes) == 0 {
		return
	}

	createStep := cloudprovider.InitTaskStep(checkCreateClusterNodeStatusStep)
	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkCreateClusterNodeStatusStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, checkCreateClusterNodeStatusStep.StepMethod)
}

// BuildRegisterClsKubeConfigStep 托管集群注册连接信息
func (cn *CreateClusterTaskOption) BuildRegisterClsKubeConfigStep(task *proto.Task) {
	if cloudprovider.IsInDependentCluster(cn.Cluster) {
		return
	}

	registerStep := cloudprovider.InitTaskStep(registerManageClusterKubeConfigStep)
	registerStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	registerStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	registerStep.Params[cloudprovider.IsExtranetKey.String()] = icommon.False

	task.Steps[registerManageClusterKubeConfigStep.StepMethod] = registerStep
	task.StepSequence = append(task.StepSequence, registerManageClusterKubeConfigStep.StepMethod)
}

// BuildNodeAnnotationsStep set node annotations
func (cn *CreateClusterTaskOption) BuildNodeAnnotationsStep(task *proto.Task) {
	if cn.NodeTemplate == nil || len(cn.NodeTemplate.Annotations) == 0 {
		return
	}
	common.BuildNodeAnnotationsTaskStep(task, cn.Cluster.ClusterID, cn.Nodes, cn.NodeTemplate.Annotations)
}

// BuildNodeLabelsStep set common node labels
func (cn *CreateClusterTaskOption) BuildNodeLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, cn.Cluster.ClusterID, cn.Nodes, nil)
}

// BuildEnableVpcCniStep 开启vpc-cni网络特性
func (cn *CreateClusterTaskOption) BuildEnableVpcCniStep(task *proto.Task) {
	enableVpcCniStep := cloudprovider.InitTaskStep(enableTkeClusterVpcCniStep)
	enableVpcCniStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	enableVpcCniStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[enableTkeClusterVpcCniStep.StepMethod] = enableVpcCniStep
	task.StepSequence = append(task.StepSequence, enableTkeClusterVpcCniStep.StepMethod)
}

// BuildUpdateTaskStatusStep 更新任务状态
func (cn *CreateClusterTaskOption) BuildUpdateTaskStatusStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateCreateClusterDBInfoStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	updateStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.Nodes, ",")

	task.Steps[updateCreateClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateCreateClusterDBInfoStep.StepMethod)
}

// CreateVirtualClusterTask 创建虚拟集群构建step子任务
type CreateVirtualClusterTask struct {
	Cluster     *proto.Cluster
	HostCluster *proto.Cluster
	Namespace   *proto.NamespaceInfo
}

// BuildCreateNamespaceStep host集群创建vcluster集群命名空间
func (cn *CreateVirtualClusterTask) BuildCreateNamespaceStep(task *proto.Task) {
	common.BuildCreateNamespaceTaskStep(task, cn.HostCluster.ClusterID, common.NamespaceDetail{
		Namespace:   cn.Namespace.Name,
		Labels:      cn.Namespace.Labels,
		Annotations: cn.Namespace.Annotations,
	})
}

// BuildCreateResourceQuotaStep host集群创建命名空间资源配额
func (cn *CreateVirtualClusterTask) BuildCreateResourceQuotaStep(task *proto.Task) {
	common.BuildCreateResourceQuotaTaskStep(task, cn.HostCluster.ClusterID, common.ResourceQuotaDetail{
		Name:        cn.Namespace.Name,
		CpuRequests: cn.Namespace.Quota.CpuRequests,
		CpuLimits:   cn.Namespace.Quota.CpuLimits,
		MemRequests: cn.Namespace.Quota.MemoryRequests,
		MemLimits:   cn.Namespace.Quota.MemoryLimits,
	})
}

// BuildInstallVclusterStep host集群安装vcluster集群
func (cn *CreateVirtualClusterTask) BuildInstallVclusterStep(task *proto.Task) {
	common.BuildInstallVclusterTaskStep(task, cn.Cluster.ClusterID, cn.HostCluster.ClusterID)
}

// BuildCheckAgentStatusStep 检测vcluster集群是否正常
func (cn *CreateVirtualClusterTask) BuildCheckAgentStatusStep(task *proto.Task) {
	common.BuildCheckKubeAgentStatusTaskStep(task, cn.Cluster.ClusterID)
}

// BuildInstallWatchStep vcluster集群安装watch组件
func (cn *CreateVirtualClusterTask) BuildInstallWatchStep(task *proto.Task) {
	common.BuildWatchComponentTaskStep(task, cn.Cluster, "")
}

// BuildUpdateTaskStatusStep 更新任务状态
func (cn *CreateVirtualClusterTask) BuildUpdateTaskStatusStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateCreateClusterDBInfoStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID

	task.Steps[updateCreateClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateCreateClusterDBInfoStep.StepMethod)
}

// DeleteVirtualClusterTaskOption 删除虚拟集群
type DeleteVirtualClusterTaskOption struct {
	Cluster     *proto.Cluster
	Cloud       *proto.Cloud
	HostCluster *proto.Cluster
	Namespace   *proto.NamespaceInfo
}

// BuildUninstallVClusterStep 删除vcluster集群
func (dc *DeleteVirtualClusterTaskOption) BuildUninstallVClusterStep(task *proto.Task) {
	common.BuildUnInstallVclusterTaskStep(task, dc.Cluster.ClusterID, dc.HostCluster.ClusterID)
}

// BuildDeleteNamespaceStep 删除命名空间
func (dc *DeleteVirtualClusterTaskOption) BuildDeleteNamespaceStep(task *proto.Task) {
	common.BuildDeleteNamespaceTaskStep(task, dc.HostCluster.ClusterID, dc.Namespace.Name)
}

// BuildCleanClusterDBInfoStep 清理集群数据
func (dc *DeleteVirtualClusterTaskOption) BuildCleanClusterDBInfoStep(task *proto.Task) {
	cleanStep := cloudprovider.InitTaskStep(cleanClusterDBInfoStep)
	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID

	task.Steps[cleanClusterDBInfoStep.StepMethod] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoStep.StepMethod)
}

// ImportClusterTaskOption 纳管集群
type ImportClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildImportClusterNodesStep 纳管集群节点
func (ic *ImportClusterTaskOption) BuildImportClusterNodesStep(task *proto.Task) {
	importNodesStep := cloudprovider.InitTaskStep(importClusterNodesStep)
	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[importClusterNodesStep.StepMethod] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesStep.StepMethod)
}

// BuildRegisterKubeConfigStep 注册集群kubeConfig
func (ic *ImportClusterTaskOption) BuildRegisterKubeConfigStep(task *proto.Task) {
	registerKubeConfigStep := cloudprovider.InitTaskStep(registerClusterKubeConfigStep)
	registerKubeConfigStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	registerKubeConfigStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[registerClusterKubeConfigStep.StepMethod] = registerKubeConfigStep
	task.StepSequence = append(task.StepSequence, registerClusterKubeConfigStep.StepMethod)
}

// BuildRegisterClusterKubeConfigStep 托管集群注册连接信息
func (ic *ImportClusterTaskOption) BuildRegisterClusterKubeConfigStep(task *proto.Task) {
	if cloudprovider.IsInDependentCluster(ic.Cluster) {
		return
	}

	registerStep := cloudprovider.InitTaskStep(registerManageClusterKubeConfigStep)
	registerStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	registerStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider
	registerStep.Params[cloudprovider.IsExtranetKey.String()] = icommon.False

	task.Steps[registerManageClusterKubeConfigStep.StepMethod] = registerStep
	task.StepSequence = append(task.StepSequence, registerManageClusterKubeConfigStep.StepMethod)
}

// DeleteClusterTaskOption 删除集群
type DeleteClusterTaskOption struct {
	// Cluster 集群
	Cluster *proto.Cluster
	// DeleteMode delete mode
	DeleteMode string
	// LastClusterStatus last cluster status
	LastClusterStatus string
}

// BuildDeleteTKEClusterStep 删除集群
func (dc *DeleteClusterTaskOption) BuildDeleteTKEClusterStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteTKEClusterStep)
	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider
	deleteStep.Params[cloudprovider.DeleteModeKey.String()] = dc.DeleteMode
	deleteStep.Params[cloudprovider.LastClusterStatus.String()] = dc.LastClusterStatus

	task.Steps[deleteTKEClusterStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteTKEClusterStep.StepMethod)
}

// BuildCleanClusterDBInfoStep 清理集群数据
func (dc *DeleteClusterTaskOption) BuildCleanClusterDBInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(cleanClusterDBInfoStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider

	task.Steps[cleanClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoStep.StepMethod)
}

// AddExternalNodesToClusterTaskOption 上架第三方节点
type AddExternalNodesToClusterTaskOption struct {
	Group   *proto.NodeGroup
	Cluster *proto.Cluster
	NodeIPs []string
}

// BuildGetExternalNodeScriptStep 获取上架第三方节点脚本
func (ac *AddExternalNodesToClusterTaskOption) BuildGetExternalNodeScriptStep(task *proto.Task) {
	getNodeScriptStep := cloudprovider.InitTaskStep(getExternalNodeScriptStep)
	getNodeScriptStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	getNodeScriptStep.Params[cloudprovider.NodeGroupIDKey.String()] = ac.Group.NodeGroupID
	getNodeScriptStep.Params[cloudprovider.CloudIDKey.String()] = ac.Group.Provider
	getNodeScriptStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(ac.NodeIPs, ",")

	task.Steps[getExternalNodeScriptStep.StepMethod] = getNodeScriptStep
	task.StepSequence = append(task.StepSequence, getExternalNodeScriptStep.StepMethod)
}

// BuildNodeLabelsStep 设置节点标签
func (ac *AddExternalNodesToClusterTaskOption) BuildNodeLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, ac.Cluster.ClusterID, ac.NodeIPs, nil)
}

// BuildUnCordonNodesStep 设置节点可调度状态
func (ac *AddExternalNodesToClusterTaskOption) BuildUnCordonNodesStep(task *proto.Task) {
	unCordonStep := cloudprovider.InitTaskStep(common.UnCordonNodesActionStep)

	unCordonStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	unCordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(ac.NodeIPs, ",")

	task.Steps[common.UnCordonNodesActionStep.StepMethod] = unCordonStep
	task.StepSequence = append(task.StepSequence, common.UnCordonNodesActionStep.StepMethod)
}

// RemoveExternalNodesFromClusterTaskOption 下架第三方节点
type RemoveExternalNodesFromClusterTaskOption struct {
	Cluster *proto.Cluster
	Group   *proto.NodeGroup
	NodeIPs []string
}

// BuildCordonNodesStep 设置节点不可调度状态
func (re *RemoveExternalNodesFromClusterTaskOption) BuildCordonNodesStep(task *proto.Task) {
	cordonStep := cloudprovider.InitTaskStep(common.CordonNodesActionStep)

	cordonStep.Params[cloudprovider.ClusterIDKey.String()] = re.Cluster.ClusterID
	cordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(re.NodeIPs, ",")

	task.Steps[common.CordonNodesActionStep.StepMethod] = cordonStep
	task.StepSequence = append(task.StepSequence, common.CordonNodesActionStep.StepMethod)
}

// BuildRemoveExternalNodesStep 下架第三方节点
func (re *RemoveExternalNodesFromClusterTaskOption) BuildRemoveExternalNodesStep(task *proto.Task) {
	removeNodesStep := cloudprovider.InitTaskStep(removeExternalNodesFromClusterStep)
	removeNodesStep.Params[cloudprovider.ClusterIDKey.String()] = re.Cluster.ClusterID
	removeNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = re.Group.NodeGroupID
	removeNodesStep.Params[cloudprovider.CloudIDKey.String()] = re.Group.Provider
	removeNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(re.NodeIPs, ",")

	task.Steps[removeExternalNodesFromClusterStep.StepMethod] = removeNodesStep
	task.StepSequence = append(task.StepSequence, removeExternalNodesFromClusterStep.StepMethod)
}

// AddNodesToClusterTaskOption 上架节点
type AddNodesToClusterTaskOption struct {
	Cluster        *proto.Cluster
	Cloud          *proto.Cloud
	NodeTemplate   *proto.NodeTemplate
	NodeIPs        []string
	NodeIDs        []string
	DiffVpcNodeIds []string

	PassWd       string
	Operator     string
	NodeSchedule bool
}

// BuildModifyInstancesVpcStep 节点转移vpc任务
func (ac *AddNodesToClusterTaskOption) BuildModifyInstancesVpcStep(task *proto.Task) {
	if len(ac.DiffVpcNodeIds) == 0 {
		return
	}

	modifyStep := cloudprovider.InitTaskStep(modifyInstancesVpcStep)
	modifyStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	modifyStep.Params[cloudprovider.CloudIDKey.String()] = ac.Cloud.CloudID
	modifyStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(ac.DiffVpcNodeIds, ",")

	task.Steps[modifyInstancesVpcStep.StepMethod] = modifyStep
	task.StepSequence = append(task.StepSequence, modifyInstancesVpcStep.StepMethod)
}

// BuildCheckInstanceStateStep 节点转移vpc任务检测
func (ac *AddNodesToClusterTaskOption) BuildCheckInstanceStateStep(task *proto.Task) {
	if len(ac.DiffVpcNodeIds) == 0 {
		return
	}

	checkStep := cloudprovider.InitTaskStep(checkInstanceStateStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = ac.Cloud.CloudID
	checkStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(ac.DiffVpcNodeIds, ",")

	task.Steps[checkInstanceStateStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkInstanceStateStep.StepMethod)
}

// BuildShieldAlertStep 屏蔽上架节点告警
func (ac *AddNodesToClusterTaskOption) BuildShieldAlertStep(task *proto.Task) {
	shieldStep := cloudprovider.InitTaskStep(addNodesShieldAlarmStep)
	shieldStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID

	task.Steps[addNodesShieldAlarmStep.StepMethod] = shieldStep
	task.StepSequence = append(task.StepSequence, addNodesShieldAlarmStep.StepMethod)
}

// BuildAddNodesToClusterStep 上架集群节点
func (ac *AddNodesToClusterTaskOption) BuildAddNodesToClusterStep(task *proto.Task) {
	addStep := cloudprovider.InitTaskStep(addNodesToClusterStep)
	addStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	addStep.Params[cloudprovider.CloudIDKey.String()] = ac.Cloud.CloudID
	templateID := ""
	if ac.NodeTemplate != nil {
		templateID = ac.NodeTemplate.GetNodeTemplateID()
	}
	addStep.Params[cloudprovider.NodeTemplateIDKey.String()] = templateID
	addStep.Params[cloudprovider.PasswordKey.String()] = ac.PassWd
	addStep.Params[cloudprovider.OperatorKey.String()] = ac.Operator
	addStep.Params[cloudprovider.NodeSchedule.String()] = strconv.FormatBool(ac.NodeSchedule)

	task.Steps[addNodesToClusterStep.StepMethod] = addStep
	task.StepSequence = append(task.StepSequence, addNodesToClusterStep.StepMethod)
}

// BuildCheckAddNodesStatusStep 检测节点状态
func (ac *AddNodesToClusterTaskOption) BuildCheckAddNodesStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkAddNodesStatusStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = ac.Cloud.CloudID

	task.Steps[checkAddNodesStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkAddNodesStatusStep.StepMethod)
}

// BuildUpdateAddNodeDBInfoStep 更新节点数据
func (ac *AddNodesToClusterTaskOption) BuildUpdateAddNodeDBInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateAddNodeDBInfoStep)

	updateStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = ac.Cloud.CloudID

	task.Steps[updateAddNodeDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateAddNodeDBInfoStep.StepMethod)
}

// BuildNodeAnnotationsStep set node annotations
func (ac *AddNodesToClusterTaskOption) BuildNodeAnnotationsStep(task *proto.Task) {
	if ac.NodeTemplate == nil || len(ac.NodeTemplate.Annotations) == 0 {
		return
	}
	common.BuildNodeAnnotationsTaskStep(task, ac.Cluster.ClusterID, ac.NodeIPs, ac.NodeTemplate.Annotations)
}

// BuildNodeLabelsStep set common node labels
func (ac *AddNodesToClusterTaskOption) BuildNodeLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, ac.Cluster.ClusterID, ac.NodeIPs, nil)
}

// BuildUnCordonNodesStep 设置节点可调度状态
func (ac *AddNodesToClusterTaskOption) BuildUnCordonNodesStep(task *proto.Task) {
	unCordonStep := cloudprovider.InitTaskStep(common.UnCordonNodesActionStep)

	unCordonStep.Params[cloudprovider.ClusterIDKey.String()] = ac.Cluster.ClusterID
	unCordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(ac.NodeIPs, ",")

	task.Steps[common.UnCordonNodesActionStep.StepMethod] = unCordonStep
	task.StepSequence = append(task.StepSequence, common.UnCordonNodesActionStep.StepMethod)
}

// RemoveNodesFromClusterTaskOption 下架节点
type RemoveNodesFromClusterTaskOption struct {
	Cluster    *proto.Cluster
	Cloud      *proto.Cloud
	DeleteMode string
	NodeIPs    []string
	NodeIDs    []string
}

// BuildCordonNodesStep 设置节点不可调度状态
func (rn *RemoveNodesFromClusterTaskOption) BuildCordonNodesStep(task *proto.Task) {
	cordonStep := cloudprovider.InitTaskStep(common.CordonNodesActionStep)

	cordonStep.Params[cloudprovider.ClusterIDKey.String()] = rn.Cluster.ClusterID
	cordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(rn.NodeIPs, ",")

	task.Steps[common.CordonNodesActionStep.StepMethod] = cordonStep
	task.StepSequence = append(task.StepSequence, common.CordonNodesActionStep.StepMethod)
}

// BuildRemoveNodesFromClusterStep 集群下架节点
func (rn *RemoveNodesFromClusterTaskOption) BuildRemoveNodesFromClusterStep(task *proto.Task) {
	removeStep := cloudprovider.InitTaskStep(removeNodesFromClusterStep)

	removeStep.Params[cloudprovider.ClusterIDKey.String()] = rn.Cluster.ClusterID
	removeStep.Params[cloudprovider.CloudIDKey.String()] = rn.Cloud.CloudID
	removeStep.Params[cloudprovider.DeleteModeKey.String()] = rn.DeleteMode
	removeStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(rn.NodeIPs, ",")
	removeStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(rn.NodeIDs, ",")

	task.Steps[removeNodesFromClusterStep.StepMethod] = removeStep
	task.StepSequence = append(task.StepSequence, removeNodesFromClusterStep.StepMethod)
}

// BuildCheckClusterCleanNodsStep 检测集群清理节点池节点
func (rn *RemoveNodesFromClusterTaskOption) BuildCheckClusterCleanNodsStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkClusterCleanNodsStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = rn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = rn.Cluster.Provider
	checkStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(rn.NodeIPs, ",")
	checkStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(rn.NodeIDs, ",")

	task.Steps[checkClusterCleanNodsStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkClusterCleanNodsStep.StepMethod)
}

// BuildUpdateRemoveNodeDBInfoStep 清理节点数据
func (rn *RemoveNodesFromClusterTaskOption) BuildUpdateRemoveNodeDBInfoStep(task *proto.Task) {
	updateDBStep := cloudprovider.InitTaskStep(updateRemoveNodeDBInfoStep)
	updateDBStep.Params[cloudprovider.ClusterIDKey.String()] = rn.Cluster.ClusterID
	updateDBStep.Params[cloudprovider.CloudIDKey.String()] = rn.Cluster.Provider
	updateDBStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(rn.NodeIPs, ",")
	updateDBStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(rn.NodeIDs, ",")

	task.Steps[updateRemoveNodeDBInfoStep.StepMethod] = updateDBStep
	task.StepSequence = append(task.StepSequence, updateRemoveNodeDBInfoStep.StepMethod)
}

// CreateNodeGroupTaskOption 创建节点组
type CreateNodeGroupTaskOption struct {
	Group *proto.NodeGroup
}

// BuildCreateCloudNodeGroupStep 通过云接口创建节点组
func (cn *CreateNodeGroupTaskOption) BuildCreateCloudNodeGroupStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createCloudNodeGroupStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[createCloudNodeGroupStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createCloudNodeGroupStep.StepMethod)
}

// BuildCheckCloudNodeGroupStatusStep 检测节点组状态
func (cn *CreateNodeGroupTaskOption) BuildCheckCloudNodeGroupStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkCloudNodeGroupStatusStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[checkCloudNodeGroupStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkCloudNodeGroupStatusStep.StepMethod)
}

// CleanNodeInGroupTaskOption 节点组缩容节点(兼容云节点和第三方节点)
type CleanNodeInGroupTaskOption struct {
	Group     *proto.NodeGroup
	NodeIPs   []string
	NodeIds   []string
	DeviceIDs []string
	Operator  string
}

// BuildCleanNodeGroupNodesStep 清理节点池节点
func (cn *CleanNodeInGroupTaskOption) BuildCleanNodeGroupNodesStep(task *proto.Task) {
	cleanStep := cloudprovider.InitTaskStep(cleanNodeGroupNodesStep)

	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	cleanStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	cleanStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	cleanStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	cleanStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIds, ",")

	task.Steps[cleanNodeGroupNodesStep.StepMethod] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanNodeGroupNodesStep.StepMethod)
}

// BuildCheckClusterCleanNodsStep 检测集群清理节点池节点
func (cn *CleanNodeInGroupTaskOption) BuildCheckClusterCleanNodsStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkClusterCleanNodsStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	checkStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	checkStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIds, ",")

	task.Steps[checkClusterCleanNodsStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkClusterCleanNodsStep.StepMethod)
}

// BuildCordonNodesStep 设置节点不可调度状态
func (cn *CleanNodeInGroupTaskOption) BuildCordonNodesStep(task *proto.Task) {
	cordonStep := cloudprovider.InitTaskStep(common.CordonNodesActionStep)

	cordonStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	cordonStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")

	task.Steps[common.CordonNodesActionStep.StepMethod] = cordonStep
	task.StepSequence = append(task.StepSequence, common.CordonNodesActionStep.StepMethod)
}

// BuildRemoveExternalNodesStep 下架第三方节点
func (cn *CleanNodeInGroupTaskOption) BuildRemoveExternalNodesStep(task *proto.Task) {
	removeNodesStep := cloudprovider.InitTaskStep(removeExternalNodesFromClusterStep)
	removeNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	removeNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	removeNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider
	removeNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")

	task.Steps[removeExternalNodesFromClusterStep.StepMethod] = removeNodesStep
	task.StepSequence = append(task.StepSequence, removeExternalNodesFromClusterStep.StepMethod)
}

// BuildReturnIDCNodeToResPoolStep 归还第三方节点
func (cn *CleanNodeInGroupTaskOption) BuildReturnIDCNodeToResPoolStep(task *proto.Task) {
	returnNodesStep := cloudprovider.InitTaskStep(returnIDCNodeToResourcePoolStep)

	returnNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	returnNodesStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	returnNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	returnNodesStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	returnNodesStep.Params[cloudprovider.DeviceIDsKey.String()] = strings.Join(cn.DeviceIDs, ",")
	returnNodesStep.Params[cloudprovider.OperatorKey.String()] = cn.Operator

	task.Steps[returnIDCNodeToResourcePoolStep.StepMethod] = returnNodesStep
	task.StepSequence = append(task.StepSequence, returnIDCNodeToResourcePoolStep.StepMethod)
}

// DeleteNodeGroupTaskOption 删除节点组
type DeleteNodeGroupTaskOption struct {
	Group                  *proto.NodeGroup
	CleanInstanceInCluster bool
}

// BuildDeleteNodeGroupStep 删除云节点组
func (dn *DeleteNodeGroupTaskOption) BuildDeleteNodeGroupStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteNodeGroupStep)

	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dn.Group.ClusterID
	deleteStep.Params[cloudprovider.NodeGroupIDKey.String()] = dn.Group.NodeGroupID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dn.Group.Provider
	deleteStep.Params[cloudprovider.KeepInstanceKey.String()] = icommon.True
	if dn.CleanInstanceInCluster {
		deleteStep.Params[cloudprovider.KeepInstanceKey.String()] = icommon.False
	}

	task.Steps[deleteNodeGroupStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodeGroupStep.StepMethod)
}

// UpdateDesiredNodesTaskOption 扩容节点组节点
type UpdateDesiredNodesTaskOption struct {
	Group    *proto.NodeGroup
	Desired  uint32
	Operator string
}

// BuildApplyInstanceMachinesStep 申请节点实例
func (ud *UpdateDesiredNodesTaskOption) BuildApplyInstanceMachinesStep(task *proto.Task) {
	applyInstanceStep := cloudprovider.InitTaskStep(applyInstanceMachinesStep)

	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider
	applyInstanceStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(ud.Desired))
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[applyInstanceMachinesStep.StepMethod] = applyInstanceStep
	task.StepSequence = append(task.StepSequence, applyInstanceMachinesStep.StepMethod)
}

// BuildApplyExternalNodeMachinesStep 申请第三方节点实例
func (ud *UpdateDesiredNodesTaskOption) BuildApplyExternalNodeMachinesStep(task *proto.Task) {
	applyExternalNodeStep := cloudprovider.InitTaskStep(applyExternalNodeMachinesStep)

	applyExternalNodeStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	applyExternalNodeStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	applyExternalNodeStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider
	applyExternalNodeStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(ud.Desired))
	applyExternalNodeStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[applyExternalNodeMachinesStep.StepMethod] = applyExternalNodeStep
	task.StepSequence = append(task.StepSequence, applyExternalNodeMachinesStep.StepMethod)
}

// BuildGetExternalNodeScriptStep 获取上架第三方节点脚本
func (ud *UpdateDesiredNodesTaskOption) BuildGetExternalNodeScriptStep(task *proto.Task) {
	getNodeScriptStep := cloudprovider.InitTaskStep(getExternalNodeScriptStep)
	getNodeScriptStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	getNodeScriptStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	getNodeScriptStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider

	task.Steps[getExternalNodeScriptStep.StepMethod] = getNodeScriptStep
	task.StepSequence = append(task.StepSequence, getExternalNodeScriptStep.StepMethod)
}

// BuildNodeLabelsStep 设置节点标签
func (ud *UpdateDesiredNodesTaskOption) BuildNodeLabelsStep(task *proto.Task) {
	common.BuildNodeLabelsTaskStep(task, ud.Group.ClusterID, nil, map[string]string{
		utils.NodeGroupLabelKey: ud.Group.GetNodeGroupID(),
	})
}

// BuildUnCordonNodesStep 设置节点可调度状态
func (ud *UpdateDesiredNodesTaskOption) BuildUnCordonNodesStep(task *proto.Task) {
	unCordonStep := cloudprovider.InitTaskStep(common.UnCordonNodesActionStep)

	unCordonStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID

	task.Steps[common.UnCordonNodesActionStep.StepMethod] = unCordonStep
	task.StepSequence = append(task.StepSequence, common.UnCordonNodesActionStep.StepMethod)
}

// BuildCheckClusterNodeStatusStep 检测节点实例状态
func (ud *UpdateDesiredNodesTaskOption) BuildCheckClusterNodeStatusStep(task *proto.Task) {
	checkClusterNodeStatusStep := cloudprovider.InitTaskStep(checkClusterNodesStatusStep)

	checkClusterNodeStatusStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	checkClusterNodeStatusStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	checkClusterNodeStatusStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider

	task.Steps[checkClusterNodesStatusStep.StepMethod] = checkClusterNodeStatusStep
	task.StepSequence = append(task.StepSequence, checkClusterNodesStatusStep.StepMethod)
}
