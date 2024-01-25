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
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	qcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// 集群下架节点

// 第三方节点下架操作

// RemoveExternalNodesFromCluster remove external nodes from cluster, 移除第三方节点
func RemoveExternalNodesFromCluster(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIPs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// filter exist external nodes
	clusterNodes, err := FilterClusterExternalNodesByIPs(ctx, info, nodeIPs)
	if err != nil {
		blog.Errorf("RemoveExternalNodesFromCluster[%s] FilterClusterExternalInstanceFromNodesIPs err: %v",
			taskID, err)
		return err
	}
	if len(clusterNodes.ExistInClusterIPs) == 0 {
		blog.Errorf("RemoveExternalNodesFromCluster[%s] FilterClusterExternalInstanceFromNodesIPs "+
			"successful, existInClusterNodes is zero", taskID)
		return nil
	}

	// delete exist external nodes
	err = DeleteClusterExternalNodes(ctx, info, clusterNodes.ExistInClusterNames)
	if err != nil {
		blog.Errorf("RemoveExternalNodesFromCluster[%s] DeleteClusterExternalNodes failed: %v", taskID, err)
		return err
	}
	blog.Infof("RemoveExternalNodesFromCluster[%s] success[%v]", taskID, nodeIPs)

	return nil
}

// ClusterExternalNodes cluster external nodes
type ClusterExternalNodes struct {
	ExistInClusterIPs      []string
	ExistInClusterNames    []string
	NotExistInClusterIPs   []string
	NotExistInClusterNames []string
}

// FilterClusterExternalNodesByIPs nodeIPs existInCluster or notExistInCluster，过滤集群第三方节点
func FilterClusterExternalNodesByIPs(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIPs []string) (*ClusterExternalNodes, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		return nil, err
	}

	var (
		clusterExternalInstances   []api.ExternalNodeInfo
		clusterExternalInstanceIPs []string
		clusterIPToName            = make(map[string]string, 0)
		clusterExternalNodes       = &ClusterExternalNodes{
			ExistInClusterIPs:      make([]string, 0),
			ExistInClusterNames:    make([]string, 0),
			NotExistInClusterIPs:   make([]string, 0),
			NotExistInClusterNames: make([]string, 0),
		}
	)
	// query cluster nodePool external nodes
	err = retry.Do(func() error {
		clusterExternalInstances, err = tkeCli.DescribeExternalNode(info.Cluster.SystemID, api.DescribeExternalNodeConfig{
			NodePoolId: info.NodeGroup.CloudNodeGroupID,
		})
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("FilterClusterExternalInstanceFromNodesIPs[%s]: "+
			"DescribeExternalNode for cluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		return nil, err
	}
	for i := range clusterExternalInstances {
		clusterExternalInstanceIPs = append(clusterExternalInstanceIPs, clusterExternalInstances[i].IP)
		clusterIPToName[clusterExternalInstances[i].IP] = clusterExternalInstances[i].Name
	}

	clusterExternalNodes.ExistInClusterIPs, clusterExternalNodes.NotExistInClusterIPs = utils.SplitExistString(clusterExternalInstanceIPs, nodeIPs) // nolint
	blog.Infof("FilterClusterExternalInstanceFromNodesIPs[%s]: "+
		"DescribeExternalNode existedInstance[%v] notExistedInstance[%v]",
		taskID, clusterExternalNodes.ExistInClusterIPs, clusterExternalNodes.NotExistInClusterIPs)

	for _, ip := range clusterExternalNodes.ExistInClusterIPs {
		name, ok := clusterIPToName[ip]
		if ok {
			clusterExternalNodes.ExistInClusterNames = append(clusterExternalNodes.ExistInClusterNames, name)
		}
	}

	return clusterExternalNodes, nil
}

// DeleteClusterExternalNodes delete TKE cluster external nodes，删除集群第三方节点
func DeleteClusterExternalNodes(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNames []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		return err
	}

	err = retry.Do(func() error {
		errExternal := tkeCli.DeleteExternalNode(info.Cluster.SystemID, api.DeleteExternalNodeConfig{
			Names: nodeNames,
			Force: true,
		})
		if errExternal != nil {
			return errExternal
		}

		return nil
	})
	if err != nil {
		blog.Errorf("DeleteClusterExternalNodes[%s]: DeleteExternalNode failed: %v", taskID, err)
		return err
	}
	blog.Infof("DeleteClusterExternalNodes[%s]: DeleteExternalNode successful[%v]",
		taskID, nodeNames)

	return nil
}

// 普通节点下架操作

// RemoveNodesFromCluster remove nodes from cluster
func RemoveNodesFromCluster(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIDs []string, force bool) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// filter exist instanceIDs
	existInClusterNodes, _, err := FilterClusterInstanceFromNodesIDs(ctx, info, nodeIDs)
	if err != nil {
		blog.Errorf("removeNodesFromCluster[%s] filterClusterNotExistInstance err: %v", taskID, err)
		return nil, err
	}
	if len(existInClusterNodes) == 0 {
		blog.Errorf("removeNodesFromCluster[%s] filterClusterNotExistInstance "+
			"successful, existInClusterNodes is zero", taskID)
		// once again delete nodes
		_, err = DeleteClusterInstance(ctx, info, nodeIDs, force)
		if err != nil {
			blog.Errorf("removeNodesFromCluster[%s] onceAgain failed: %v", taskID, err)
		}

		return nil, nil
	}

	// delete exist instanceIDs
	success, err := DeleteClusterInstance(ctx, info, existInClusterNodes, force)
	if err != nil {
		blog.Errorf("removeNodesFromCluster[%s] deleteClusterInstance failed: %v", taskID, err)
		return nil, err
	}
	blog.Infof("removeNodesFromCluster[%s] success, origin[%v] success[%v]", taskID, nodeIDs, success)

	return success, nil
}

// FilterClusterInstanceFromNodesIDs nodeIDs existInCluster or notExistInCluster
func FilterClusterInstanceFromNodesIDs(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIDs []string) ([]string, []string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	ctx = utils.WithTraceIDForContext(ctx, taskID)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		return nil, nil, err
	}

	var (
		clusterInstances   []*api.InstanceInfo
		clusterInstanceIDs []string
	)
	// query cluster All instances
	err = retry.Do(func() error {
		clusterInstances, err = tkeCli.QueryTkeClusterAllInstances(ctx, info.Cluster.SystemID, nil)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("filterClusterNotExistInstance[%s]: QueryTkeClusterAllInstances for cluster[%s] failed, %v",
			taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}
	for i := range clusterInstances {
		clusterInstanceIDs = append(clusterInstanceIDs, clusterInstances[i].InstanceID)
	}

	blog.Infof("filterClusterNotExistInstance[%s]: QueryTkeClusterAllInstances for cluster[%s] instanceID[%+v], "+
		"nodeID[%+v]", taskID, info.Cluster.ClusterID, clusterInstanceIDs, nodeIDs)

	existInCluster, notExistInCluster := utils.SplitExistString(clusterInstanceIDs, nodeIDs)
	blog.Infof("filterClusterNotExistInstance[%s]: QueryTkeClusterAllInstances existedInstance[%v] notExistedInstance[%v]",
		taskID, existInCluster, notExistInCluster)

	return existInCluster, notExistInCluster, nil
}

// DeleteClusterInstance delete TKE cluster Instances
func DeleteClusterInstance(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, deleteNodeIDs []string, force bool) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		return nil, err
	}

	var (
		successIDs, failedIDs, notFoundIDs []string
	)
	err = retry.Do(func() error {
		result, errDelete := tkeCli.DeleteTkeClusterInstance(&api.DeleteInstancesRequest{
			ClusterID:   info.Cluster.SystemID,
			Instances:   deleteNodeIDs,
			ForceDelete: force,
		})
		if errDelete != nil {
			return errDelete
		}
		successIDs = result.Success
		failedIDs = result.Failure
		notFoundIDs = result.NotFound

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("deleteClusterInstance[%s]: DeleteTkeClusterInstance failed: %v", taskID, err)
		return nil, err
	}
	blog.Infof("deleteClusterInstance[%s]: DeleteTkeClusterInstance result, success[%v] failed[%v] notFound[%v]",
		taskID, successIDs, failedIDs, notFoundIDs)

	return successIDs, nil
}

// GetClusterExternalNodeScript get cluster externalNode script，获取第三方节点上架脚本
func GetClusterExternalNodeScript(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		return "", err
	}

	var (
		nodeScriptResp *api.DescribeExternalNodeScriptResponseParams
		script         string
	)
	// query cluster nodePool external nodes
	err = retry.Do(func() error {
		nodeScriptResp, err = tkeCli.DescribeExternalNodeScript(info.Cluster.SystemID, api.DescribeExternalNodeScriptConfig{
			NodePoolId: info.NodeGroup.CloudNodeGroupID,
		})
		if err != nil {
			blog.Errorf("GetClusterExternalNodeScript[%s] DescribeExternalNodeScript failed: %v", taskID, err)
			return err
		}

		script = base64.StdEncoding.EncodeToString([]byte(*nodeScriptResp.Command))
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("GetClusterExternalNodeScript[%s]: DescribeExternalNodeScript for cluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		return "", err
	}

	blog.Infof("GetClusterExternalNodeScript[%s] successful: requestID[%s] resp: cmd[%s] token[%s] link[%s]",
		taskID, nodeScriptResp.RequestId, *nodeScriptResp.Command, *nodeScriptResp.Token, *nodeScriptResp.Link)

	return script, nil
}

// 集群上架节点

// 上架节点并检查节点状态

// GenerateNTAddExistedInstanceReq 生成上架节点请求. 节点模板抽象理论上需要用户保证节点配置高度一致, 若用户配置了多盘挂载,
// 则使用用户配置选项若没有配置节点模板 或者 节点模板没有配置多盘选项, 则需要自动进行多盘挂载
func GenerateNTAddExistedInstanceReq(info *cloudprovider.CloudDependBasicInfo, nodeIDs, nodeIPs []string,
	passwd, operator string, options *NodeAdvancedOptions) *api.AddExistedInstanceReq {
	req := &api.AddExistedInstanceReq{
		ClusterID:   info.Cluster.SystemID,
		InstanceIDs: nodeIDs,
		AdvancedSetting: GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   strings.Join(nodeIPs, ","),
			Operator: operator,
			Render:   true,
		}, options),
		LoginSetting:        &api.LoginSettings{Password: passwd},
		SkipValidateOptions: skipValidateOption(info.Cluster),
	}

	// 未使用节点模板 或者 节点模板未配置磁盘格式化
	if info.NodeTemplate == nil || len(info.NodeTemplate.DataDisks) == 0 {
		// 使用默认配置, 主要解决CVM多盘挂载问题
		req.InstanceAdvancedSettingsOverrides = make([]*api.InstanceAdvancedSettings, 0)

		instanceDisk, err := GetNodeInstanceDataDiskInfo(nodeIDs, info.CmOption)
		if err != nil {
			blog.Errorf("GenerateNTAddExistedInstanceReq GetNodeInstanceDataDiskInfo failed: %v", err)
		} else {
			// 控制instance 和 InstanceAdvancedSettings 顺序
			for i := range nodeIDs {
				if disk, ok := instanceDisk[nodeIDs[i]]; ok {
					blog.Infof("GenerateNTAddExistedInstanceReq[%s] generate overrideInstanceAdvanced[%v]",
						disk.InstanceID, disk.DiskCount)

					overrideInstanceAdvanced := GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
						Cluster:  info.Cluster,
						IPList:   strings.Join(nodeIPs, ","),
						Operator: operator,
						Render:   true,
					}, options)
					// has many data disk
					if disk.DiskCount > 1 {
						overrideInstanceAdvanced.DataDisks = []api.DataDetailDisk{api.GetDefaultDataDisk("")}
					}

					req.InstanceAdvancedSettingsOverrides = append(req.InstanceAdvancedSettingsOverrides,
						overrideInstanceAdvanced)
				}
			}
		}
	}

	return req
}

// GenerateClsAdvancedInsSettingFromNT xxx
func GenerateClsAdvancedInsSettingFromNT(info *cloudprovider.CloudDependBasicInfo,
	vars template.RenderVars, options *NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	if info.NodeTemplate == nil {
		return generateInstanceAdvanceInfo(info.Cluster, options)
	}

	return generateInstanceAdvanceInfoFromNp(info.Cluster, info.NodeTemplate, ClusterCommonLabels(info.Cluster),
		"", vars, options)
}

// generateInstanceAdvanceInfo instance advanced info
func generateInstanceAdvanceInfo(cluster *proto.Cluster, options *NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	if cluster.NodeSettings.MountTarget == "" {
		cluster.NodeSettings.MountTarget = common.MountTarget
	}
	if cluster.NodeSettings.DockerGraphPath == "" {
		cluster.NodeSettings.DockerGraphPath = common.DockerGraphPath
	}

	// advanced instance setting
	advanceInfo := &api.InstanceAdvancedSettings{
		MountTarget:     cluster.NodeSettings.MountTarget,
		DockerGraphPath: cluster.NodeSettings.DockerGraphPath,
		Unschedulable: func() *int64 {
			if options != nil && options.NodeScheduler {
				return qcommon.Int64Ptr(0)
			}

			return qcommon.Int64Ptr(int64(cluster.NodeSettings.UnSchedulable))
		}(),
	}

	// node common labels
	if len(ClusterCommonLabels(cluster)) > 0 {
		for key, value := range ClusterCommonLabels(cluster) {
			advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
				Name:  key,
				Value: value,
			})
		}
	}

	// cluster node common labels
	if len(cluster.NodeSettings.Labels) > 0 {
		for key, value := range cluster.NodeSettings.Labels {
			advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
				Name:  key,
				Value: value,
			})
		}
	}

	// Kubelet start params
	if len(cluster.NodeSettings.ExtraArgs) > 0 {
		advanceInfo.ExtraArgs = &api.InstanceExtraArgs{}

		if kubelet, ok := cluster.NodeSettings.ExtraArgs[common.Kubelet]; ok {
			paras := strings.Split(kubelet, ";")
			advanceInfo.ExtraArgs.Kubelet = paras
		}
	}

	return advanceInfo
}

// skipValidateOption 是否忽略容器IP不足校验
func skipValidateOption(cls *proto.Cluster) []string {
	skipNetworkValidate := []string{api.VpcCniCIDRCheck}

	if cls.ExtraInfo != nil {
		v, ok := cls.ExtraInfo[api.GlobalRouteCIDRCheck]
		if ok && v == common.True {
			skipNetworkValidate = append(skipNetworkValidate, api.GlobalRouteCIDRCheck)
		}
	}

	return skipNetworkValidate
}

// NodeGroup生成上架节点请求, 解决多盘问题主要取决于用户是否配置 多盘挂载, 类比于qcloud产品

// GenerateNGAddExistedInstanceReq xxx
func GenerateNGAddExistedInstanceReq(info *cloudprovider.CloudDependBasicInfo, nodeIDs, nodeIPs []string,
	passwd, operator string, options *NodeAdvancedOptions) *api.AddExistedInstanceReq {
	req := &api.AddExistedInstanceReq{
		ClusterID:   info.Cluster.SystemID,
		InstanceIDs: nodeIDs,
		AdvancedSetting: generateClsAdvancedInsSettingFromNP(info, template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   strings.Join(nodeIPs, ","),
			Operator: operator,
			Render:   true,
		}, options),
		LoginSetting:        &api.LoginSettings{Password: passwd},
		SkipValidateOptions: skipValidateOption(info.Cluster),
	}

	return req
}

func generateClsAdvancedInsSettingFromNP(info *cloudprovider.CloudDependBasicInfo,
	vars template.RenderVars, options *NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	if info.NodeGroup.NodeTemplate == nil {
		return generateInstanceAdvanceInfoFromNg(info.Cluster, info.NodeGroup, options)
	}

	// inject common labels
	commonLabels := ClusterCommonLabels(info.Cluster)
	nodeGroupLabels := info.NodeGroup.GetLabels()
	newLabels := utils.MergeMap(commonLabels, nodeGroupLabels)

	return generateInstanceAdvanceInfoFromNp(info.Cluster, info.NodeGroup.NodeTemplate, newLabels,
		info.NodeGroup.NodeGroupID, vars, options)
}

func getDefaultNodePath(cluster *proto.Cluster) (string, string) {
	mountPath := cluster.GetNodeSettings().GetMountTarget()
	if len(mountPath) == 0 {
		mountPath = common.MountTarget
	}
	dockerPath := cluster.GetNodeSettings().GetDockerGraphPath()
	if len(dockerPath) == 0 {
		dockerPath = common.DockerGraphPath
	}

	return mountPath, dockerPath
}

// getNodeCommonLabels common labels
func getNodeCommonLabels(cls *proto.Cluster, group *proto.NodeGroup) []*api.KeyValue {
	labels := make([]*api.KeyValue, 0)

	if group != nil {
		labels = append(labels, &api.KeyValue{
			Name:  utils.NodeGroupLabelKey,
			Value: group.NodeGroupID,
		})
		for k, v := range group.Labels {
			labels = append(labels, &api.KeyValue{
				Name:  k,
				Value: v,
			})
		}
	}
	if cls != nil {
		for k, v := range ClusterCommonLabels(cls) {
			labels = append(labels, &api.KeyValue{
				Name:  k,
				Value: v,
			})
		}
	}

	return labels
}

func generateInstanceAdvanceInfoFromNg(cluster *proto.Cluster, group *proto.NodeGroup,
	options *NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	mountTarget, dockerGraphPath := getDefaultNodePath(cluster)

	advanceInfo := &api.InstanceAdvancedSettings{
		MountTarget:     mountTarget,
		DockerGraphPath: dockerGraphPath,
		// 默认值是0 表示参与调度, 节点池的由用户决定是否立即加入调度; 若需要各个流程执行完毕可调度,则需设置为 不可调度
		Unschedulable: func() *int64 {
			if options != nil && options.NodeScheduler {
				return qcommon.Int64Ptr(0)
			}
			return &common.Unschedulable
		}(),

		Labels: getNodeCommonLabels(cluster, group),
	}

	return advanceInfo
}

// nodeGroupID for nodePool label
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
// nolint
func generateInstanceAdvanceInfoFromNp(cls *proto.Cluster, nodeTemplate *proto.NodeTemplate, labels map[string]string,
	nodeGroupID string, vars template.RenderVars, options *NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	var (
		mountTarget     = nodeTemplate.GetMountTarget()
		dockerGraphPath = nodeTemplate.GetDockerGraphPath()
	)
	if mountTarget == "" {
		mountTarget = common.MountTarget
	}
	if dockerGraphPath == "" {
		dockerGraphPath = common.DockerGraphPath
	}

	advanceInfo := &api.InstanceAdvancedSettings{
		MountTarget:     mountTarget,
		DockerGraphPath: dockerGraphPath,
		// 默认值是0 表示参与调度, 节点池的由用户决定是否立即加入调度; 若需要各个流程执行完毕可调度,则需设置为 不可调度
		// NOCC:CCN_threshold(工具误报:)
		Unschedulable: func() *int64 {
			if options != nil && options.NodeScheduler {
				return qcommon.Int64Ptr(0)
			}
			return &common.Unschedulable
		}(),
	}
	// 后置脚本 base64 编码的用户脚本, 此脚本会在 k8s 组件运行后执行, 需要用户保证脚本的可重入及重试逻辑
	/*
		if len(nodeTemplate.UserScript) > 0 {
			script, err := getNodeTemplateScript(vars, nodeTemplate.UserScript)
			if err != nil {
				blog.Errorf("generateInstanceAdvanceInfoFromNp getNodeTemplateScript failed: %v", err)
			}
			advanceInfo.UserScript = script
		}
	*/

	if len(nodeGroupID) > 0 {
		advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
			Name:  utils.NodeGroupLabelKey,
			Value: nodeGroupID,
		})
	}
	if len(nodeTemplate.Labels) > 0 || len(labels) > 0 {
		for key, value := range nodeTemplate.Labels {
			advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
				Name:  key,
				Value: value,
			})
		}
		// 兼容旧版本nodeGroup, 若存在labels则进行设置
		for key, value := range labels {
			advanceInfo.Labels = append(advanceInfo.Labels, &api.KeyValue{
				Name:  key,
				Value: value,
			})
		}
	}
	if len(nodeTemplate.Taints) > 0 {
		for _, t := range nodeTemplate.Taints {
			advanceInfo.TaintList = append(advanceInfo.TaintList, &api.Taint{
				Key:    qcommon.StringPtr(t.Key),
				Value:  qcommon.StringPtr(t.Value),
				Effect: qcommon.StringPtr(t.Effect),
			})
		}
	}
	if len(nodeTemplate.DataDisks) > 0 {
		for i, disk := range nodeTemplate.DataDisks {
			diskSize, _ := strconv.Atoi(disk.DiskSize)
			if disk.DiskPartition == "" && i < len(api.DefaultDiskPartition) {
				disk.DiskPartition = api.DefaultDiskPartition[i]
			}

			advanceInfo.DataDisks = append(advanceInfo.DataDisks, api.DataDetailDisk{
				DiskType:           disk.DiskType,
				DiskSize:           int64(diskSize),
				AutoFormatAndMount: disk.AutoFormatAndMount,
				FileSystem:         disk.FileSystem,
				MountTarget:        disk.MountTarget,
				DiskPartition:      disk.DiskPartition,
			})
		}
	}
	if len(nodeTemplate.ExtraArgs) > 0 || len(cls.NodeSettings.ExtraArgs) > 0 {
		if advanceInfo.ExtraArgs == nil {
			advanceInfo.ExtraArgs = &api.InstanceExtraArgs{}
		}
		kubeletParams := make([]string, 0)
		if kubePara, ok := nodeTemplate.ExtraArgs[common.Kubelet]; ok {
			paras := strings.Split(kubePara, ";")
			kubeletParams = append(kubeletParams, paras...)
		}
		if kubelet, ok := cls.NodeSettings.ExtraArgs[common.Kubelet]; ok {
			paras := strings.Split(kubelet, ";")
			kubeletParams = append(kubeletParams, paras...)
		}
		advanceInfo.ExtraArgs.Kubelet = kubeletParams
	}

	if len(nodeTemplate.PreStartUserScript) > 0 {
		script, err := template.GetNodeTemplateScript(vars, nodeTemplate.PreStartUserScript)
		if err != nil {
			blog.Errorf("generateInstanceAdvanceInfoFromNp getNodeTemplateScript failed: %v", err)
		}
		advanceInfo.PreStartUserScript = script
	}

	return advanceInfo
}

// AddExistedInstanceResult add existed result
type AddExistedInstanceResult struct {
	SuccessNodes []string
	FailedNodes  []string
}

// AddNodesToCluster add nodes to cluster and return nodes result
func AddNodesToCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, options *NodeAdvancedOptions,
	nodeIDs []string, passwd string, isNodeGroup bool, idToIP map[string]string,
	operator string) (*AddExistedInstanceResult, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] NewTkeClient failed: %v", taskID, err)
		return nil, err
	}

	var (
		resp           *api.AddExistedInstanceRsp
		result         = &AddExistedInstanceResult{}
		addInstanceReq *api.AddExistedInstanceReq
		nodeIPs        = make([]string, 0)
	)
	for i := range nodeIDs {
		if ip, ok := idToIP[nodeIDs[i]]; ok {
			nodeIPs = append(nodeIPs, ip)
		}
	}

	// nodeGroup or nodeTemplate
	if isNodeGroup {
		addInstanceReq = GenerateNGAddExistedInstanceReq(info, nodeIDs, nodeIPs, passwd, operator, options)
	} else {
		addInstanceReq = GenerateNTAddExistedInstanceReq(info, nodeIDs, nodeIPs, passwd, operator, options)
	}

	blog.Infof("AddNodesToCluster[%s] AddExistedInstancesToCluster request[%+v]", taskID, *addInstanceReq)

	err = retry.Do(func() error {
		resp, err = tkeCli.AddExistedInstancesToCluster(addInstanceReq)
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] AddExistedInstancesToCluster failed: %v", taskID, err)
		return nil, err
	}

	if len(resp.SuccessInstanceIDs) > 0 {
		result.SuccessNodes = resp.SuccessInstanceIDs
	}
	if len(resp.FailedInstanceIDs) > 0 || len(resp.TimeoutInstanceIDs) > 0 {
		result.FailedNodes = append(result.FailedNodes, resp.TimeoutInstanceIDs...)
		result.FailedNodes = append(result.FailedNodes, resp.FailedInstanceIDs...)
	}

	blog.Infof("AddNodesToCluster[%s] AddExistedInstancesToCluster success[%v] failed[%v]"+
		"reasons[%v]", taskID, result.SuccessNodes, result.FailedNodes, resp.FailedReasons)

	return result, nil
}

// CheckClusterInstanceStatus 检测集群节点状态 && 更新节点状态
func CheckClusterInstanceStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	instanceIDs []string) ([]string, []string, error) {
	var (
		addSuccessNodes = make([]string, 0)
		addFailureNodes = make([]string, 0)
	)

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterInstanceStatus[%s] failed, %s", taskID, err)
		return nil, nil, err
	}

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 12*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err = loop.LoopDoFunc(timeCtx, func() error {
		instances, errQuery := cli.QueryTkeClusterInstances(&api.DescribeClusterInstances{
			ClusterID:   info.Cluster.SystemID,
			InstanceIDs: instanceIDs,
		})
		if errQuery != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, errQuery)
			return nil
		}

		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, ins := range instances {
			blog.Infof("checkClusterInstanceStatus[%s] instance[%s] status[%s]", taskID, *ins.InstanceId, *ins.InstanceState)
			switch *ins.InstanceState {
			case api.RunningInstanceTke.String():
				running = append(running, *ins.InstanceId)
				index++
			case api.FailedInstanceTke.String():
				failure = append(failure, *ins.InstanceId)
				index++
			default:
			}
		}

		if index == len(instanceIDs) {
			addSuccessNodes = running
			addFailureNodes = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(20*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, err)
		return nil, nil, err
	}
	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, err)

		running, failure := make([]string, 0), make([]string, 0)
		instances, errQuery := cli.QueryTkeClusterInstances(&api.DescribeClusterInstances{
			ClusterID:   info.Cluster.SystemID,
			InstanceIDs: instanceIDs,
		})
		if errQuery != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, errQuery)
			return nil, nil, errQuery
		}
		for _, ins := range instances {
			blog.Infof("checkClusterInstanceStatus[%s] instance[%s] status[%s]", taskID, *ins.InstanceId, *ins.InstanceState)
			switch *ins.InstanceState {
			case api.RunningInstanceTke.String():
				running = append(running, *ins.InstanceId)
			default:
				failure = append(failure, *ins.InstanceId)
			}
		}
		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterInstanceStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	// set cluster node status
	for _, n := range addFailureNodes {
		err = cloudprovider.UpdateNodeStatus(false, n, common.StatusAddNodesFailed)
		if err != nil {
			blog.Errorf("checkClusterInstanceStatus[%s] UpdateNodeStatusByInstanceID[%s] failed: %v", taskID, n, err)
		}
	}

	return addSuccessNodes, addFailureNodes, nil
}

// GetFailedNodesReason get add nodes failed reason
func GetFailedNodesReason(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceIDs []string) (map[string]InstanceInfo, string, error) {
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("GetFailedNodesReason[%s] failed: %v", taskId, err)
		return nil, "", err
	}

	var (
		insMapInfo   = make(map[string]InstanceInfo, 0)
		allInsReason = make([]string, 0)
	)
	for i := range instanceIDs {
		reason, _ := tkeCli.DescribeInstanceCreateProgress(info.Cluster.GetSystemID(), instanceIDs[i])
		insMapInfo[instanceIDs[i]] = InstanceInfo{
			NodeId:       instanceIDs[i],
			FailedReason: reason,
		}
	}
	for _, ins := range insMapInfo {
		allInsReason = append(allInsReason, ins.GetNodeFailedReason())
	}

	return insMapInfo, strings.Join(allInsReason, ";"), nil
}

// DeleteTkeClusterByClusterId delete cluster by clsId
func DeleteTkeClusterByClusterId(ctx context.Context, opt *cloudprovider.CommonOption,
	clsId string, deleteMode string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clsId) == 0 {
		blog.Warnf("DeleteTkeClusterByClusterId[%s] clusterID empty", taskID)
		return nil
	}

	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] init tkeClient failed: %v", taskID, err)
		return err
	}

	err = tkeCli.DeleteTKECluster(clsId, api.DeleteMode(deleteMode))
	if err != nil && !strings.Contains(err.Error(), api.ErrClusterNotFound.Error()) {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] deleteCluster failed: %v", taskID, err)
		return err
	}

	blog.Infof("DeleteTkeClusterByClusterId[%s] deleteCluster[%s] success", taskID, clsId)

	return nil
}

