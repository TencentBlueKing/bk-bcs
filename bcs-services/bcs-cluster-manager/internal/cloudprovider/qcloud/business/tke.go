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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	qcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// 集群相关接口

// GetTkeCluster returns cluster by clusterID
func GetTkeCluster(opt *cloudprovider.CommonOption, clusterID string) (*tke.Cluster, error) {
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		return nil, err
	}

	cluster, err := tkeCli.GetTKECluster(clusterID)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// GetClusterVpcCniStatus cluster vpc-cni status
func GetClusterVpcCniStatus(cls *tke.Cluster) bool {
	if cls != nil && cls.Property != nil {
		if strings.Contains(*cls.Property, api.TKERouteEni) || strings.Contains(*cls.Property, api.TKEDirectEni) {
			return true
		}
	}

	return false
}

// GetClusterVpcCniSubnets cluster vpc-cni subnets
func GetClusterVpcCniSubnets(cls *tke.Cluster) []string {
	subnetIDs := make([]string, 0)

	if cls != nil && cls.ClusterNetworkSettings != nil && len(cls.ClusterNetworkSettings.Subnets) > 0 {
		for _, subnet := range cls.ClusterNetworkSettings.Subnets {
			subnetIDs = append(subnetIDs, *subnet)
		}
	}

	blog.Infof("getClusterVpcCniSubnets %v", subnetIDs)
	return subnetIDs
}

// GetClusterSubnetsZoneUsage get cluster subnets zone usage
func GetClusterSubnetsZoneUsage(cmOption *cloudprovider.CommonOption, subnetIDs []string, extraIP bool) (
	map[string]*ZoneSubnetRatio, float64, error) {
	subnets, err := GetDrSubnetInfo(cmOption, subnetIDs)
	if err != nil {
		return nil, 0, err
	}

	zoneSubnetNum := make(map[string]*ZoneSubnetRatio, 0)
	for i := range subnets {
		if zoneSubnetNum[subnets[i].Zone] == nil {
			zoneSubnetNum[subnets[i].Zone] = &ZoneSubnetRatio{}
		}

		if extraIP {
			zoneSubnetNum[subnets[i].Zone].TotalIps += subnets[i].TotalIps + 3
		} else {
			zoneSubnetNum[subnets[i].Zone].TotalIps += subnets[i].TotalIps
		}
		zoneSubnetNum[subnets[i].Zone].AvailableIps += subnets[i].AvailableIps
	}

	var (
		totalIps     uint64
		availableIps uint64
	)

	for zone := range zoneSubnetNum {
		zoneSubnetNum[zone].Ratio = 100 * (float64(zoneSubnetNum[zone].TotalIps-zoneSubnetNum[zone].AvailableIps) /
			float64(zoneSubnetNum[zone].TotalIps))

		totalIps += zoneSubnetNum[zone].TotalIps
		availableIps += zoneSubnetNum[zone].AvailableIps
	}

	totalRatio := 100 * (float64(totalIps-availableIps) / float64(totalIps))

	return zoneSubnetNum, totalRatio, nil
}

// ZoneSubnetRatio zone subnet ratio
type ZoneSubnetRatio struct {
	TotalIps     uint64
	AvailableIps uint64
	Ratio        float64
}

// GetClusterCurrentVpcCniSubnets get tke cluster subnets
func GetClusterCurrentVpcCniSubnets(cls *proto.Cluster, extraIP bool) (map[string]*ZoneSubnetRatio,
	float64, []string, error) {
	cmOption, err := cloudprovider.GetCloudCmOptionByCluster(cls)
	if err != nil {
		return nil, 0, nil, err
	}

	tkeCls, err := GetTkeCluster(cmOption, cls.GetSystemID())
	if err != nil {
		return nil, 0, nil, err
	}

	if !GetClusterVpcCniStatus(tkeCls) {
		return nil, 0, nil, fmt.Errorf("cluster not enable vpc-cni mode")
	}

	// 获取集群子网
	subnetIds := GetClusterVpcCniSubnets(tkeCls)
	if len(subnetIds) == 0 {
		return nil, 0, nil, fmt.Errorf("clsuter[%s] subnets empty", cls.GetClusterID())
	}

	zoneSubnetNum, totalRatio, err := GetClusterSubnetsZoneUsage(cmOption, subnetIds, extraIP)
	if err != nil {
		return nil, 0, nil, err
	}

	return zoneSubnetNum, totalRatio, subnetIds, nil
}

// AddSubnetsToCluster add subnet to tke cluster
func AddSubnetsToCluster(cls *proto.Cluster, subnets []string, opt *cloudprovider.CommonOption) error {
	client, err := api.NewTkeClient(opt)
	if err != nil {
		return err
	}

	err = client.AddVpcCniSubnets(&api.AddVpcCniSubnetsInput{
		ClusterID: cls.GetSystemID(),
		VpcID:     cls.GetVpcID(),
		SubnetIDs: subnets,
	})
	if err != nil {
		blog.Errorf("AddSubnetsToCluster[%s] failed:%v", cls.GetClusterID(), err)
		return err
	}
	if cls.GetNetworkSettings().GetEniSubnetIDs() == nil {
		cls.GetNetworkSettings().EniSubnetIDs = make([]string, 0)
	}

	cls.GetNetworkSettings().EniSubnetIDs = append(cls.GetNetworkSettings().EniSubnetIDs, subnets...)
	return cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cls)
}

// 集群下架节点

// 第三方节点下架操作

// RemoveExternalNodesFromCluster remove external nodes from cluster, 移除第三方节点
func RemoveExternalNodesFromCluster(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIPs []string) error {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

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

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("success [%v]", nodeIPs))

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
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

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

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("origin [%v] success [%v]", nodeIDs, success))

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
func GetClusterExternalNodeScript(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	internal bool) (string, error) {
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
			Internal:   internal,
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
func GenerateNTAddExistedInstanceReq(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs, nodeIPs []string, passwd, operator string, options *NodeAdvancedOptions) *api.AddExistedInstanceReq {
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

	if options != nil && options.Advance != nil {
		if options.Advance.GetNodeOs() != "" {
			req.ImageId = options.Advance.GetNodeOs()
		}
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

					// only handle cvm first /dev/vdb disk
					if disk.DiskCount >= 1 {
						overrideInstanceAdvanced.DataDisks = []api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
					}

					req.InstanceAdvancedSettingsOverrides = append(req.InstanceAdvancedSettingsOverrides,
						overrideInstanceAdvanced)
				}
			}
		}
	}

	return req
}

// genGpuAdvSettingOverride 生成GPU节点的节点定制化高级配置
func genGpuAdvSettingOverride(req *api.AddExistedInstanceReq, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs, nodeIPs []string, operator string, options *NodeAdvancedOptions, gpuNodeTemplate *proto.NodeTemplate) {
	// 未使用节点模板 或者 节点模板未配置磁盘格式化
	if info.NodeTemplate == nil || len(info.NodeTemplate.DataDisks) == 0 {
		// 使用默认配置, 主要解决CVM多盘挂载问题
		req.InstanceAdvancedSettingsOverrides = make([]*api.InstanceAdvancedSettings, 0)
		instanceDisk, err := GetNodeInstanceDataDiskInfo(nodeIDs, info.CmOption)
		if err != nil {
			blog.Errorf("GenerateGPUAddExistedInstanceReqs GetNodeInstanceDataDiskInfo failed: %v", err)
		} else {
			// 控制instance 和 InstanceAdvancedSettings 顺序
			for i := range nodeIDs {
				if disk, ok := instanceDisk[nodeIDs[i]]; ok {
					blog.Infof("GenerateGPUAddExistedInstanceReqs[%s] generate overrideInstanceAdvanced[%v]",
						disk.InstanceID, disk.DiskCount)

					overrideInstanceAdvanced := GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
						Cluster:  info.Cluster,
						IPList:   strings.Join(nodeIPs, ","),
						Operator: operator,
						Render:   true,
					}, options)

					// only handle cvm first /dev/vdb disk
					if disk.DiskCount >= 1 {
						overrideInstanceAdvanced.DataDisks = []api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
					}
					overrideInstanceAdvanced.GPUArgs = generateGPUArgs(gpuNodeTemplate.GpuArgs)

					req.InstanceAdvancedSettingsOverrides = append(req.InstanceAdvancedSettingsOverrides,
						overrideInstanceAdvanced)
				}
			}
		}
	}
}

// gpuNodeTemplatesMapByNodeIDs get gpuNodeTemplatesMap by nodeIDs, map key: nodeId, value: gpuNodeTemplate
func getGPUNodeTemplatesMapByNodeIDs(
	nodeIDs []string, cmOption *cloudprovider.CommonOption) (map[string]*proto.NodeTemplate, error) {
	gpuNodeTemplates := make(map[string]*proto.NodeTemplate)
	nodes, err := TransInstanceIDsToNodes(nodeIDs, &cloudprovider.ListNodesOption{
		Common: cmOption,
	})
	if err != nil {
		blog.Errorf("getGPUNodeTemplatesMapByNodeIDs TransInstanceIDsToNodes ids[%s] err:%s", nodeIDs, err.Error())
		return nil, err
	}

	for i := range nodes {
		gpuNodeTemplate, err := cloudprovider.GetGpuNodeTemplate(nodes[i].InstanceType)
		if err != nil {
			blog.Errorf("getGPUNodeTemplatesMapByNodeIDs GetGPUNodeTemplate failed: %v", err)
			return nil, err
		}
		gpuNodeTemplates[nodes[i].NodeID] = gpuNodeTemplate
	}

	return gpuNodeTemplates, nil
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
			advanceInfo.ExtraArgs.Kubelet = utils.FilterEmptyString(paras)
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

// GenerateGPUAddExistedInstanceReqs generate gpu add existed instance request
func GenerateGPUAddExistedInstanceReqs(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs, nodeIPs []string, idToIP map[string]string, passwd, operator string,
	options *NodeAdvancedOptions) []*api.AddExistedInstanceReq {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	reqs := make([]*api.AddExistedInstanceReq, 0)

	// GPU节点分类, 将相同模版的节点进行聚合
	imagesToGpuNodesInfoMap, err := getImagesToGpuNodesInfoMap(nodeIDs, info)
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] getNodesImagesToIdsMap failed: %v", taskID, err)
		return nil
	}

	for imageId, gpuNodeInfo := range imagesToGpuNodesInfoMap {
		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("gpu imageId:[%s], nodeIds:[%s], nodeTemplate gpuArgs:[%+v]",
				imageId, gpuNodeInfo.nodeIds, gpuNodeInfo.gpuNodeTemplate.GetGpuArgs()))

		imageNodeIps := make([]string, 0)
		imageNodeIds := gpuNodeInfo.nodeIds
		for i := range imageNodeIds {
			if ip, ok := idToIP[imageNodeIds[i]]; ok {
				imageNodeIps = append(imageNodeIps, ip)
			}
		}

		req := &api.AddExistedInstanceReq{
			ClusterID:   info.Cluster.SystemID,
			InstanceIDs: imageNodeIds,
			AdvancedSetting: GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
				Cluster:  info.Cluster,
				IPList:   strings.Join(imageNodeIps, ","),
				Operator: operator,
				Render:   true,
			}, options),
			LoginSetting:        &api.LoginSettings{Password: passwd},
			SkipValidateOptions: skipValidateOption(info.Cluster),
		}

		req.AdvancedSetting.GPUArgs = generateGPUArgs(gpuNodeInfo.gpuNodeTemplate.GetGpuArgs())
		req.ImageId = imageId

		genGpuAdvSettingOverride(req, info, imageNodeIds, imageNodeIps, operator, options, gpuNodeInfo.gpuNodeTemplate)

		for _, o := range req.InstanceAdvancedSettingsOverrides {
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("InstanceAdvancedSettingsOverrides gpuArgs:[%+v]", o.GPUArgs))
		}

		reqs = append(reqs, req)
	}

	return reqs
}

// NodeGroup生成上架节点请求, 解决多盘问题主要取决于用户是否配置 多盘挂载, 类比于qcloud产品

// GenerateNGAddExistedInstanceReq xxx
func GenerateNGAddExistedInstanceReq(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs, nodeIPs []string, passwd, operator string, options *NodeAdvancedOptions) *api.AddExistedInstanceReq {
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

	if info.NodeGroup.GetNodeTemplate().GetImage().GetImageID() != "" {
		req.ImageId = info.NodeGroup.GetNodeTemplate().GetImage().GetImageID()
	}

	generateGpuInfoInNGReq(ctx, req, info, nodeIDs)

	return req
}

func generateGpuInfoInNGReq(ctx context.Context, req *api.AddExistedInstanceReq,
	info *cloudprovider.CloudDependBasicInfo, nodeIDs []string) {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)
	// check gpu node
	nodes, err := TransInstanceIDsToNodes(nodeIDs, &cloudprovider.ListNodesOption{
		Common: info.CmOption,
	})
	if err != nil {
		blog.Errorf("GenerateNGAddExistedInstanceReq generateGpuInfoInNGReq TransInstanceIDsToNodes ids[%s] err:%s",
			nodeIDs, err.Error())
		return
	}
	for _, node := range nodes {
		if !node.GetIsGpuNode() {
			return
		}
	}

	instanceType := info.NodeGroup.GetLaunchTemplate().GetInstanceType()
	gpuNodeTemplate, err := cloudprovider.GetGpuNodeTemplate(instanceType)
	if err != nil || gpuNodeTemplate == nil {
		blog.Errorf("GenerateNGAddExistedInstanceReq instances[%v] GetGpuNodeTemplate failed, err:[%s]",
			instanceType, err)
		cloudprovider.GetStorageModel().CreateTaskStepLogWarn(context.Background(), taskID, stepName,
			fmt.Sprintf("GenerateNGAddExistedInstanceReq instances[%v] not found GPUNodeTemplate, err:[%s], "+
				"falling back to default settings", instanceType, err))

		// if not find gpu node template, use default gpu node template
		defaultGpuNodeTemplate, localErr := cloudprovider.GetGpuNodeTemplate(common.DefaultGpuNodeTemplateName)
		if localErr != nil {
			blog.Errorf("GenerateNGAddExistedInstanceReq instances[%v] GetGpuNodeTemplate failed, err:[%s]",
				instanceType, localErr)
			return
		}
		gpuNodeTemplate = defaultGpuNodeTemplate
	}

	req.ImageId = setImageIdFromNodeTemplate(gpuNodeTemplate, info.CmOption)
	req.AdvancedSetting.GPUArgs = generateGPUArgs(gpuNodeTemplate.GetGpuArgs())

	gpuArgsJSON, _ := json.Marshal(req.AdvancedSetting.GPUArgs)
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("GenerateNGAddExistedInstanceReq instances[%v] found GPUNodeTemplate, "+
			"gpuArgs:[%s], imageId:[%s]", instanceType, gpuArgsJSON, req.ImageId))

}

func setImageIdFromNodeTemplate(nodeTemplate *proto.NodeTemplate, cmOption *cloudprovider.CommonOption) string {
	if imageId := nodeTemplate.GetImage().GetImageID(); imageId != "" {
		return imageId
	}

	imageName := nodeTemplate.GetImage().GetImageName()
	if imageName == "" {
		return ""
	}
	imageId, err := GetCVMImageIDByImageName(imageName, cmOption)
	if err != nil {
		blog.Errorf("GenerateNTAddExistedInstanceReq setImageIdFromNodeTemplate GetCVMImageIDByImageName failed: %v", err)
	}
	return imageId
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
			kubeletParams = append(kubeletParams, utils.FilterEmptyString(paras)...)
		}
		if kubelet, ok := cls.NodeSettings.ExtraArgs[common.Kubelet]; ok {
			paras := strings.Split(kubelet, ";")
			kubeletParams = append(kubeletParams, utils.FilterEmptyString(paras)...)
		}
		advanceInfo.ExtraArgs.Kubelet = kubeletParams
	}

	if len(nodeTemplate.PreStartUserScript) > 0 && (options == nil || options.SetPreStartUserScript) {
		script, err := template.GetNodeTemplateScript(vars, nodeTemplate.PreStartUserScript)
		if err != nil {
			blog.Errorf("generateInstanceAdvanceInfoFromNp getNodeTemplateScript failed: %v", err)
		}
		advanceInfo.PreStartUserScript = script
	}

	if options != nil && options.Advance != nil && options.Advance.GetIsGPUNode() && nodeTemplate.GpuArgs != nil {
		gpuArgs := nodeTemplate.GpuArgs
		advanceInfo.GPUArgs = generateGPUArgs(gpuArgs)
	}

	return advanceInfo
}

func generateGPUArgs(gpuArgs *proto.GPUArgs) *api.GPUArgs {
	if gpuArgs == nil {
		return nil
	}

	result := &api.GPUArgs{}

	result.MIGEnable = gpuArgs.MigEnable

	if gpuArgs.Driver != nil {
		result.Driver = &api.DriverVersion{
			Version: gpuArgs.Driver.Version,
			Name:    gpuArgs.Driver.Name,
		}
	}

	if gpuArgs.Cuda != nil {
		result.CUDA = &api.DriverVersion{
			Version: gpuArgs.Cuda.Version,
			Name:    gpuArgs.Cuda.Name,
		}
	}

	if gpuArgs.Cudnn != nil {
		result.CUDNN = &api.CUDNN{
			Version: gpuArgs.Cudnn.Version,
			Name:    gpuArgs.Cudnn.Name,
			DocName: gpuArgs.Cudnn.DocName,
			DevName: gpuArgs.Cudnn.DevName,
		}
	}

	if gpuArgs.CustomDriver != nil {
		result.CustomDriver = &api.CustomDriver{
			Address: gpuArgs.CustomDriver.Address,
		}
	}

	return result
}

// AddExistedInstanceResult add existed result
type AddExistedInstanceResult struct {
	SuccessNodeInfos []InstanceInfo
	FailedNodeInfos  []InstanceInfo
	FailedReasons    []string
}

// AddNodesToCluster add nodes to cluster and return nodes result
// nolint:funlen
func AddNodesToCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, options *NodeAdvancedOptions,
	nodeIDs []string, passwd string, isNodeGroup bool, idToIP map[string]string,
	operator string) (*AddExistedInstanceResult, error) {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] NewTkeClient failed: %v", taskID, err)
		return nil, err
	}

	var (
		resp            *api.AddExistedInstanceRsp
		addInstanceReqs []*api.AddExistedInstanceReq
		result          = &AddExistedInstanceResult{}
		nodeIPs         = make([]string, 0)
	)

	for i := range nodeIDs {
		if ip, ok := idToIP[nodeIDs[i]]; ok {
			nodeIPs = append(nodeIPs, ip)
		}
	}

	// nodeGroup or nodeTemplate
	if isNodeGroup {
		addInstanceReq := GenerateNGAddExistedInstanceReq(ctx, info, nodeIDs, nodeIPs, passwd, operator, options)
		addInstanceReqs = append(addInstanceReqs, addInstanceReq)
	} else {
		if options != nil && options.Advance != nil && options.Advance.GetIsGPUNode() {
			addInstanceReqs = GenerateGPUAddExistedInstanceReqs(ctx, info, nodeIDs, nodeIDs, idToIP, passwd, operator, options)
		} else {
			addInstanceReq := GenerateNTAddExistedInstanceReq(ctx, info, nodeIDs, nodeIDs, passwd, operator, options)
			addInstanceReqs = append(addInstanceReqs, addInstanceReq)
		}
	}

	if len(addInstanceReqs) == 0 {
		return nil, fmt.Errorf("AddNodesToCluster[%s] addInstanceReqs is empty", taskID)
	}

	for _, addReq := range addInstanceReqs {

		blog.Infof("AddNodesToCluster[%s] AddExistedInstancesToCluster request[%+v]", taskID, addReq)

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("generate addReq image[%v] to instances[%v]", addReq.ImageId, nodeIDs))

		err = retry.Do(func() error {
			resp, err = tkeCli.AddExistedInstancesToCluster(addReq)
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
			for i := range resp.SuccessInstanceIDs {
				nodeID := resp.SuccessInstanceIDs[i]
				instanceInfo := InstanceInfo{
					NodeId: nodeID,
					NodeIp: idToIP[nodeID],
				}
				result.SuccessNodeInfos = append(result.SuccessNodeInfos, instanceInfo)
			}
		}
		if len(resp.FailedInstanceIDs) > 0 || len(resp.TimeoutInstanceIDs) > 0 {
			for i := range resp.FailedInstanceIDs {
				nodeID := resp.FailedInstanceIDs[i]
				instanceInfo := InstanceInfo{
					NodeId: nodeID,
					NodeIp: idToIP[nodeID],
				}
				result.FailedNodeInfos = append(result.FailedNodeInfos, instanceInfo)
			}
			for i := range resp.TimeoutInstanceIDs {
				nodeID := resp.TimeoutInstanceIDs[i]
				instanceInfo := InstanceInfo{
					NodeId: nodeID,
					NodeIp: idToIP[nodeID],
				}
				result.FailedNodeInfos = append(result.FailedNodeInfos, instanceInfo)
			}
			if len(resp.FailedInstanceIDs) > 0 {
				result.FailedReasons = resp.FailedReasons
			}
		}
	}

	blog.Infof("AddNodesToCluster[%s] AddExistedInstancesToCluster success [%v],"+
		"failed [%v], reasons[%v]", taskID, result.SuccessNodeInfos, result.FailedNodeInfos, result.FailedReasons)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("success nodes: [%v]", result.SuccessNodeInfos))

	if len(result.FailedNodeInfos) > 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("failed nodes: [%v], reason [%v]", result.FailedNodeInfos, result.FailedReasons))
	}

	return result, nil
}

// gpuNodesInfo xx
type gpuNodesInfo struct {
	nodeIds         []string
	gpuNodeTemplate *proto.NodeTemplate
}

// getImagesToGpuNodesInfoMap get gpu nodes info map
func getImagesToGpuNodesInfoMap(
	nodeIds []string, info *cloudprovider.CloudDependBasicInfo) (map[string]*gpuNodesInfo, error) {
	imageToGpuNodesInfo := make(map[string]*gpuNodesInfo)
	// 获取每个GPU节点使用的模版
	gpuNodeTemplatesMap, err := getGPUNodeTemplatesMapByNodeIDs(nodeIds, info.CmOption)
	if err != nil {
		blog.Errorf("gpuNodeTemplatesMap[%s] failed: %v", nodeIds, err)
		return nil, err
	}

	for _, nodeId := range nodeIds {
		gpuNodeTemplate, find := gpuNodeTemplatesMap[nodeId]
		if find {
			var imageID string
			if gpuNodeTemplate.GetImage().GetImageID() != "" {
				imageID = gpuNodeTemplate.GetImage().GetImageID()
			} else {
				imageName := gpuNodeTemplate.GetImage().GetImageName()
				imageID, err = GetCVMImageIDByImageName(imageName, info.CmOption)
				if err != nil {
					blog.Errorf("getNodesImagesToIdsMap GetCVMImageIDByImageName failed: %v", err)
					continue
				}
			}

			// 使用相同镜像的节点聚合
			if _, exists := imageToGpuNodesInfo[imageID]; !exists {
				imageToGpuNodesInfo[imageID] = &gpuNodesInfo{
					nodeIds:         []string{nodeId},
					gpuNodeTemplate: gpuNodeTemplate,
				}
			} else {
				imageToGpuNodesInfo[imageID].nodeIds = append(imageToGpuNodesInfo[imageID].nodeIds, nodeId)
			}
		}
	}

	return imageToGpuNodesInfo, nil
}

// CheckClusterDeletedNodes check if nodeIds are deleted in cluster
func CheckClusterDeletedNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIds []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("CheckClusterDeletedNodes[%s] failed, %s", taskID, err)
		return err
	}

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err = loop.LoopDoFunc(timeCtx, func() error {
		instances, errQuery := cli.QueryTkeClusterAllInstances(ctx, info.Cluster.GetSystemID(), nil)
		if errQuery != nil {
			blog.Errorf("CheckClusterDeletedNodes[%s] QueryTkeClusterAllInstances failed: %v", taskID, errQuery)
			return nil
		}

		if len(instances) == 0 {
			return loop.EndLoop
		}

		clusterNodeIds := make([]string, 0)
		for i := range instances {
			clusterNodeIds = append(clusterNodeIds, instances[i].InstanceID)
		}

		for i := range nodeIds {
			if utils.StringInSlice(nodeIds[i], clusterNodeIds) {
				blog.Infof("CheckClusterDeletedNodes[%s] %s in cluster[%v]", taskID, nodeIds[i], clusterNodeIds)
				return nil
			}
		}

		return loop.EndLoop
	}, loop.LoopInterval(20*time.Second))
	// other error
	if err != nil {
		blog.Errorf("CheckClusterDeletedNodes[%s] failed: %v", taskID, err)
		return err
	}

	blog.Infof("CheckClusterDeletedNodes[%s] deleted nodes success[%v]", taskID, nodeIds)
	return nil
}

// CheckClusterInstanceStatus 检测集群节点状态 && 更新节点状态
func CheckClusterInstanceStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	instanceIDs []string) ([]string, []string, error) {
	var (
		addSuccessNodes = make([]InstanceInfo, 0)
		addFailureNodes = make([]InstanceInfo, 0)
	)

	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	// get task timeout from template config, if no configure, use default timeout
	defaultTime := 30 * time.Minute
	taskTimeout := cloudprovider.GetTaskTimeout(ctx, info.Cluster.ProjectID,
		info.Cluster.ClusterID, stepName, defaultTime)

	// get qcloud client
	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterInstanceStatus[%s] failed, %s", taskID, err)
		return nil, nil, err
	}

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), taskTimeout)
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

		var (
			index   int
			running = make([]InstanceInfo, 0)
			failure = make([]InstanceInfo, 0)
		)
		for _, ins := range instances {
			// instances info
			blog.Infof("provider[%s] checkClusterInstanceStatus[%s] businessID[%s] projectID[%s] cluster[%s] "+
				"instanceInfo[%s:%s] status[%s]", info.Cluster.GetProvider(), taskID, info.Cluster.GetBusinessID(),
				info.Cluster.GetProjectID(), utils.StringPtrToString(ins.InstanceId),
				utils.StringPtrToString(ins.LanIP), info.Cluster.GetSystemID(),
				utils.StringPtrToString(ins.InstanceState))
			switch *ins.InstanceState {
			case api.RunningInstanceTke.String():
				running = append(running, InstanceInfo{
					NodeId: *ins.InstanceId,
					NodeIp: *ins.LanIP,
				})
				index++
			case api.FailedInstanceTke.String():
				failure = append(failure, InstanceInfo{
					NodeId: *ins.InstanceId,
					NodeIp: *ins.LanIP,
				})
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

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("QueryTkeClusterInstances DeadlineExceeded: %v", err))

		blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, err)

		var (
			running = make([]InstanceInfo, 0)
			failure = make([]InstanceInfo, 0)
		)
		instances, errQuery := cli.QueryTkeClusterInstances(&api.DescribeClusterInstances{
			ClusterID:   info.Cluster.SystemID,
			InstanceIDs: instanceIDs,
		})
		if errQuery != nil {

			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("QueryTkeClusterInstances errQuery: %v", errQuery))

			blog.Errorf("checkClusterInstanceStatus[%s] QueryTkeClusterInstances failed: %v", taskID, errQuery)
			return nil, nil, errQuery
		}
		for _, ins := range instances {
			switch *ins.InstanceState {
			case api.RunningInstanceTke.String():
				running = append(running, InstanceInfo{
					NodeId: *ins.InstanceId,
					NodeIp: *ins.LanIP,
				})
			default:
				failure = append(failure, InstanceInfo{
					NodeId: *ins.InstanceId,
					NodeIp: *ins.LanIP,
				})
			}
		}
		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterInstanceStatus[%s] success instances[%v] failure instance[%v]",
		taskID, addSuccessNodes, addFailureNodes)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("success instance [%v]", addSuccessNodes))

	if len(addFailureNodes) > 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("failure instance [%v]", addFailureNodes))

		// set cluster node status
		for _, n := range addFailureNodes {
			err = cloudprovider.UpdateNodeStatus(false, n.NodeId, common.StatusAddNodesFailed)
			if err != nil {
				blog.Errorf("checkClusterInstanceStatus[%s] UpdateNodeStatusByInstanceID[%s] failed: %v", taskID, n, err)
			}
		}
	}

	var (
		successNodeIds []string
		failedNodeIds  []string
	)
	for i := range addSuccessNodes {
		successNodeIds = append(successNodeIds, addSuccessNodes[i].NodeId)
	}
	for i := range addFailureNodes {
		failedNodeIds = append(failedNodeIds, addFailureNodes[i].NodeId)
	}

	return successNodeIds, failedNodeIds, nil
}

// GetFailedNodesReason get add nodes failed reason
func GetFailedNodesReason(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceIDs []string) (map[string]InstanceInfo, string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("GetFailedNodesReason[%s] failed: %v", taskID, err)
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

// DeleteTkeClusterByClusterID delete cluster by clsId
func DeleteTkeClusterByClusterID(ctx context.Context, opt *cloudprovider.CommonOption,
	clsID string, deleteMode string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clsID) == 0 {
		blog.Warnf("DeleteTkeClusterByClusterId[%s] clusterID empty", taskID)
		return nil
	}

	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] init tkeClient failed: %v", taskID, err)
		return err
	}

	err = tkeCli.DeleteTKECluster(clsID, api.DeleteMode(deleteMode))
	if err != nil && !strings.Contains(err.Error(), api.ErrClusterNotFound.Error()) {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] deleteCluster failed: %v", taskID, err)
		return err
	}

	blog.Infof("DeleteTkeClusterByClusterId[%s] deleteCluster[%s] success", taskID, clsID)

	return nil
}

// EnableClusterAudit enable cluster audit
func EnableClusterAudit(ctx context.Context, cls *proto.Cluster, opt *cloudprovider.EnableClusterAuditOption) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		return err
	}

	err = tkeCli.EnableClusterAudit(cls.SystemID, opt.LogsetId, opt.TopicId, opt.TopicRegion, opt.WithoutCollection)
	if err != nil {
		return err
	}

	blog.Infof("EnableClusterAudit[%s] cluster[%s] success", taskID, cls.SystemID)

	return nil
}
