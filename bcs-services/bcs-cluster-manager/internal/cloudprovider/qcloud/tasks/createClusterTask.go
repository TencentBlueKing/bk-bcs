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
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"github.com/ghodss/yaml"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// as far as possible to keep every operation unit simple

// generateClusterCIDRInfo cidr info
func generateClusterCIDRInfo(ctx context.Context, cluster *proto.Cluster,
	opt *cloudprovider.CommonOption) (*api.ClusterCIDRSettings, error) {
	cidrInfo := &api.ClusterCIDRSettings{
		ClusterCIDR:          cluster.NetworkSettings.ClusterIPv4CIDR,
		MaxNodePodNum:        uint64(cluster.NetworkSettings.MaxNodePodNum),
		MaxClusterServiceNum: uint64(cluster.NetworkSettings.MaxServiceNum),
		ServiceCIDR:          cluster.NetworkSettings.ServiceIPv4CIDR,
	}

	// cluster.NetworkSettings.ClusterIPv4CIDR is empty, auto allocate cidr when create cluster
	if cluster.NetworkSettings.ClusterIPv4CIDR == "" {

		cloudprovider.GetDistributeLock().Lock(utils.BuildAllocateVpcCidrLockKey(
			cluster.Provider, cluster.Region, cluster.VpcID), []lock.LockOption{lock.LockTTL(time.Second * 5)}...)
		defer cloudprovider.GetDistributeLock().Unlock(utils.BuildAllocateVpcCidrLockKey(
			cluster.Provider, cluster.Region, cluster.VpcID))

		mask := utils.CalMaskLen(float64(cluster.NetworkSettings.CidrStep))
		subnet, err := business.AllocateGrCidrSubnet(ctx, opt, cluster.GetProvider(),
			cluster.VpcID, int(mask), nil)
		if err != nil {
			return nil, err
		}
		cidrInfo.ClusterCIDR = subnet.ID
		cluster.NetworkSettings.ClusterIPv4CIDR = subnet.ID

		// update cluster cidr info
		_ = cloudprovider.UpdateCluster(cluster)
	}

	return cidrInfo, nil
}

// generateTags tags info
func generateTags(bizID int64, operator string) map[string]string {
	cli := cmdb.GetCmdbClient()
	if cli == nil {
		return nil
	}

	// get internal cloud tags
	tags, err := cli.GetCloudTags(cmdb.BizInfo{BizID: bizID}, operator)
	if err != nil {
		blog.Errorf("TKE cluster generateTags failed: %v", err)
		return nil
	}

	return tags
}

// generateClusterBasicInfo cluster basic info
func generateClusterBasicInfo(cluster *proto.Cluster, imageID, operator string) *api.ClusterBasicSettings {
	basicInfo := &api.ClusterBasicSettings{
		ClusterOS:                 imageID,
		ClusterVersion:            cluster.ClusterBasicSettings.Version,
		ClusterName:               cluster.ClusterID,
		VpcID:                     cluster.VpcID,
		SubnetID:                  cluster.ClusterBasicSettings.SubnetID,
		ClusterLevel:              cluster.ClusterBasicSettings.ClusterLevel,
		IsAutoUpgradeClusterLevel: cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel,
	}

	basicInfo.TagSpecification = make([]*api.TagSpecification, 0)
	// build qcloud tag info
	if len(cluster.ClusterBasicSettings.ClusterTags) > 0 {
		tags := make([]*api.Tag, 0)
		for k, v := range cluster.ClusterBasicSettings.ClusterTags {
			tags = append(tags, &api.Tag{
				Key:   common.StringPtr(k),
				Value: common.StringPtr(v),
			})
		}
		basicInfo.TagSpecification = append(basicInfo.TagSpecification, &api.TagSpecification{
			ResourceType: "cluster",
			Tags:         tags,
		})
	}

	// internal cloud tags
	if options.GetEditionInfo().IsInnerEdition() {
		// according to cloud different realization to adapt
		bizID, _ := strconv.Atoi(cluster.BusinessID)
		cloudTags := generateTags(int64(bizID), operator)

		blog.Infof("generateClusterBasicInfo tags %+v", cloudTags)

		tags := make([]*api.Tag, 0)
		if len(cloudTags) > 0 {
			for k, v := range cloudTags {
				tags = append(tags, &api.Tag{
					Key:   common.StringPtr(k),
					Value: common.StringPtr(v),
				})
			}

			basicInfo.TagSpecification = append(basicInfo.TagSpecification, &api.TagSpecification{
				ResourceType: "cluster",
				Tags:         tags,
			})
		}
	}

	return basicInfo
}

// generateClusterAdvancedInfo cluster advanced info
func generateClusterAdvancedInfo(cluster *proto.Cluster) *api.ClusterAdvancedSettings {
	advancedInfo := &api.ClusterAdvancedSettings{
		IPVS:               cluster.ClusterAdvanceSettings.IPVS,
		ContainerRuntime:   cluster.ClusterAdvanceSettings.ContainerRuntime,
		RuntimeVersion:     cluster.ClusterAdvanceSettings.RuntimeVersion,
		ExtraArgs:          &api.ClusterExtraArgs{},
		NetworkType:        cluster.ClusterAdvanceSettings.NetworkType,
		DeletionProtection: cluster.ClusterAdvanceSettings.DeletionProtection,
	}

	if options.GetEditionInfo().IsInnerEdition() {
		advancedInfo.AuditEnabled = true
	} else {
		advancedInfo.AuditEnabled = cluster.ClusterAdvanceSettings.AuditEnabled
	}

	// extraArgs
	if len(cluster.ClusterAdvanceSettings.ExtraArgs) > 0 {
		if apiserver, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeAPIServer]; ok {
			paras := strings.Split(apiserver, ";")
			advancedInfo.ExtraArgs.KubeAPIServer = common.StringPtrs(paras)
		}

		if controller, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeController]; ok {
			paras := strings.Split(controller, ";")
			advancedInfo.ExtraArgs.KubeControllerManager = common.StringPtrs(paras)
		}

		if scheduler, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeScheduler]; ok {
			paras := strings.Split(scheduler, ";")
			advancedInfo.ExtraArgs.KubeScheduler = common.StringPtrs(paras)
		}

		if etcd, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.Etcd]; ok {
			paras := strings.Split(etcd, ";")
			advancedInfo.ExtraArgs.Etcd = common.StringPtrs(paras)
		}
	}

	return advancedInfo
}

// generateInstanceAdvanceInfo instance advanced info
func generateInstanceAdvanceInfo(cluster *proto.Cluster,
	options *business.NodeAdvancedOptions) *api.InstanceAdvancedSettings {
	if cluster.NodeSettings.MountTarget == "" {
		cluster.NodeSettings.MountTarget = icommon.MountTarget
	}
	if cluster.NodeSettings.DockerGraphPath == "" {
		cluster.NodeSettings.DockerGraphPath = icommon.DockerGraphPath
	}

	// advanced instance setting
	advanceInfo := &api.InstanceAdvancedSettings{
		MountTarget:     cluster.NodeSettings.MountTarget,
		DockerGraphPath: cluster.NodeSettings.DockerGraphPath,
		Unschedulable: func() *int64 {
			if options != nil && options.NodeScheduler {
				return common.Int64Ptr(0)
			}

			return common.Int64Ptr(int64(cluster.NodeSettings.UnSchedulable))
		}(),
	}

	// node common labels
	if len(business.ClusterCommonLabels(cluster)) > 0 {
		for key, value := range business.ClusterCommonLabels(cluster) {
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

		if kubelet, ok := cluster.NodeSettings.ExtraArgs[icommon.Kubelet]; ok {
			paras := strings.Split(kubelet, ";")
			advanceInfo.ExtraArgs.Kubelet = utils.FilterEmptyString(paras)
		}
	}

	return advanceInfo
}

// handleClusterMasterNodes handle cluster master nodes
func handleClusterMasterNodes(ctx context.Context, req *api.CreateClusterRequest,
	info *cloudprovider.CloudDependBasicInfo, passwd string, instanceIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// filter nodes data disks
	filterDisk, err := business.FilterNodesByDataDisk(instanceIDs, info.CmOption)
	if err != nil {
		blog.Errorf("createClusterReq[%s] FilterNodesByDataDisk[%s] failed: %+v",
			taskID, info.Cluster.ClusterID, err)
		return err
	}

	blog.Infof("createClusterReq FilterNodesByDataDisk result[%+v]", filterDisk)

	if req.ExistedInstancesForNode == nil {
		req.ExistedInstancesForNode = make([]*api.ExistedInstancesForNode, 0)
	}

	// single disk & many disk
	if len(filterDisk.SingleDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateMasterExistedInstance(api.MASTER_ETCD.String(), passwd, filterDisk.SingleDiskInstance, false, info.Cluster))
	}
	if len(filterDisk.ManyDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateMasterExistedInstance(api.MASTER_ETCD.String(), passwd, filterDisk.ManyDiskInstance, true, info.Cluster))
	}

	return nil
}

// generateMasterExistedInstance cluster master setting
func generateMasterExistedInstance(role, passwd string, instanceIDs []string, manyDisk bool,
	cls *proto.Cluster) *api.ExistedInstancesForNode {
	existedInstance := &api.ExistedInstancesForNode{
		NodeRole: role,
		ExistedInstancesPara: &api.ExistedInstancesPara{
			InstanceIDs:              instanceIDs,
			LoginSettings:            &api.LoginSettings{Password: passwd},
			InstanceAdvancedSettings: generateInstanceAdvanceInfo(cls, nil),
		},
	}

	// instance advanced setting override
	existedInstance.InstanceAdvancedSettingsOverride = generateInstanceAdvanceInfo(cls, nil)
	if manyDisk {
		existedInstance.InstanceAdvancedSettingsOverride.DataDisks =
			[]api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
	}

	return existedInstance
}

// handleClusterWorkerNodes handle cluster worker nodes
func handleClusterWorkerNodes(ctx context.Context, req *api.CreateClusterRequest,
	info *cloudprovider.CloudDependBasicInfo, passwd string, instanceIDs []string, operator string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(instanceIDs) == 0 {
		return nil
	}

	if req.ExistedInstancesForNode == nil {
		req.ExistedInstancesForNode = make([]*api.ExistedInstancesForNode, 0)
	}

	// filter nodes data disks
	filterDisk, err := business.FilterNodesByDataDisk(instanceIDs, info.CmOption)
	if err != nil {
		blog.Errorf("handleClusterWorkerNodes[%s] FilterNodesByDataDisk[%s] failed: %+v",
			taskID, info.Cluster.ClusterID, err)
		retErr := fmt.Errorf("call FilterNodesByDataDisk[%s] api err, %s", info.Cluster.ClusterID, err.Error())
		return retErr
	}

	blog.Infof("handleClusterWorkerNodes[%s] FilterNodesByDataDisk result[%+v]", taskID, filterDisk)

	// single disk
	if len(filterDisk.SingleDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateWorkerExistedInstance(info, filterDisk.SingleDiskInstance, filterDisk.SingleDiskInstanceIP, passwd,
				false, operator))
	}
	// many disk
	if len(filterDisk.ManyDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateWorkerExistedInstance(info, filterDisk.ManyDiskInstance, filterDisk.ManyDiskInstanceIP, passwd,
				true, operator))
	}

	return nil
}

// generateWorkerExistedInstance cluster worker setting
func generateWorkerExistedInstance(info *cloudprovider.CloudDependBasicInfo, nodeIDs, nodeIPs []string,
	passwd string, manyDisk bool, operator string) *api.ExistedInstancesForNode {
	existedInstance := &api.ExistedInstancesForNode{
		NodeRole: api.WORKER.String(),
		ExistedInstancesPara: &api.ExistedInstancesPara{
			InstanceIDs:   nodeIDs,
			LoginSettings: &api.LoginSettings{Password: passwd},
			InstanceAdvancedSettings: business.GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
				Cluster:  info.Cluster,
				IPList:   strings.Join(nodeIPs, ","),
				Operator: operator,
				Render:   true,
			}, &business.NodeAdvancedOptions{NodeScheduler: true}),
		},
	}

	// instance advanced setting
	existedInstance.InstanceAdvancedSettingsOverride = business.GenerateClsAdvancedInsSettingFromNT(info,
		template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   strings.Join(nodeIPs, ","),
			Operator: operator,
			Render:   true,
		}, &business.NodeAdvancedOptions{NodeScheduler: true})

	if manyDisk {
		existedInstance.InstanceAdvancedSettingsOverride.DataDisks =
			[]api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
	}

	return existedInstance
}

// disksToCVMDisks transfer cvm disk
func disksToCVMDisks(disks []*proto.DataDisk) []*cvm.DataDisk {
	if len(disks) == 0 {
		return nil
	}

	cvmDisks := make([]*cvm.DataDisk, 0)
	for i := range disks {
		size, _ := utils.StringToInt(disks[i].DiskSize)
		cvmDisks = append(cvmDisks, &cvm.DataDisk{
			DiskSize: common.Int64Ptr(int64(size)),
			DiskType: common.StringPtr(disks[i].DiskType),
		})
	}

	return cvmDisks
}

// generateRunInstance run instances
func generateRunInstance(cluster *proto.Cluster, role, passwd string) *api.RunInstancesForNode {
	runInstance := &api.RunInstancesForNode{
		NodeRole: role,
	}

	// create instance template
	for i := range cluster.Template {
		systemDiskSize, _ := utils.StringToInt(cluster.Template[i].SystemDisk.DiskSize)
		req := &cvm.RunInstancesRequest{
			Placement: &cvm.Placement{
				Zone: common.StringPtr(cluster.Template[i].Zone),
			},
			InstanceType: common.StringPtr(cluster.Template[i].InstanceType),
			ImageId:      common.StringPtr(cluster.Template[i].ImageInfo.ImageID),
			SystemDisk: &cvm.SystemDisk{
				DiskType: common.StringPtr(cluster.Template[i].SystemDisk.DiskType),
				DiskSize: common.Int64Ptr(int64(systemDiskSize)),
			},
			DataDisks: disksToCVMDisks(cluster.Template[i].DataDisks),
			VirtualPrivateCloud: &cvm.VirtualPrivateCloud{
				VpcId:    common.StringPtr(cluster.Template[i].VpcID),
				SubnetId: common.StringPtr(cluster.Template[i].SubnetID),
			},

			InstanceCount: common.Int64Ptr(int64(cluster.Template[i].ApplyNum)),
			LoginSettings: &cvm.LoginSettings{
				Password: common.StringPtr(passwd),
			},
		}

		requestStr := req.ToJsonString()
		runInstance.RunInstancesPara = append(runInstance.RunInstancesPara, common.StringPtr(requestStr))
	}

	return runInstance
}

// CreateClusterShieldAlarmTask call alarm interface to shield alarm
func CreateClusterShieldAlarmTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateClusterShieldAlarmTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateClusterShieldAlarmTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// masterIP list
	masterIPs := cloudprovider.GetClusterMasterIPList(cluster)

	allIPs := make([]string, 0)
	if len(masterIPs) > 0 {
		allIPs = append(allIPs, masterIPs...)
	}
	if len(nodes) > 0 {
		allIPs = append(allIPs, nodes...)
	}
	blog.Infof("CreateClusterShieldAlarmTask[%s] ShieldHostAlarmConfig: %+v", taskID, allIPs)

	if len(allIPs) > 0 {
		err = cloudprovider.ShieldHostAlarm(ctx, cluster.BusinessID, masterIPs)
		if err != nil {
			blog.Errorf("CreateClusterShieldAlarmTask[%s] ShieldHostAlarmConfig failed: %v", taskID, err)
		} else {
			blog.Infof("CreateClusterShieldAlarmTask[%s] ShieldHostAlarmConfig successful", taskID)
		}
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}
	return nil
}

type clusterInfo struct {
	systemID  string
	masterIPs []string
	masterIDs []string
	nodeIPs   []string
	nodeIDs   []string
}

// createTkeCluster create tke cluster
func createTkeCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	nodeIPs []string, passwd, operator string) (*clusterInfo, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	var (
		err       error
		masterIDs []string
		nodeIDs   []string
	)

	// get qcloud client
	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("createTkeCluster[%s]: get tke client for cluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		return nil, retErr
	}

	// image info
	imageID, err := transImageNameToImageID(info.CmOption, info.Cluster.ClusterBasicSettings.OS)
	if err != nil {
		blog.Errorf("createTkeCluster[%s]: transImageNameToImageID for cluster[%s] failed, %v",
			taskID, info.Cluster.ClusterID, err)
		retErr := fmt.Errorf("createTkeCluster transImageNameToImageID err, %s", err)
		return nil, retErr
	}

	// passwd
	if passwd == "" {
		passwd = utils.BuildInstancePwd()
	}

	// masterIP list
	masterIPs := cloudprovider.GetClusterMasterIPList(info.Cluster)
	if len(masterIPs) > 0 {
		masterNodes, errTrans := transIPsToInstances(&cloudprovider.ListNodesOption{
			Common:       info.CmOption,
			ClusterVPCID: info.Cluster.VpcID,
		}, masterIPs)
		if errTrans != nil || len(masterNodes) == 0 {
			blog.Errorf("createTkeCluster[%s]: transMasterIPs for cluster[%s] failed: %v",
				taskID, info.Cluster.ClusterID, errTrans)
			retErr := fmt.Errorf("createTkeCluster transMasterIPs err, %s", errTrans)
			return nil, retErr
		}

		for i := range masterNodes {
			masterIDs = append(masterIDs, masterNodes[i].NodeID)
		}
	}

	// handle nodeIPs if exist
	if len(nodeIPs) > 0 {
		nodes, errTrans := transIPsToInstances(&cloudprovider.ListNodesOption{
			Common:       info.CmOption,
			ClusterVPCID: info.Cluster.VpcID,
		}, nodeIPs)
		if errTrans != nil || len(nodes) == 0 {
			blog.Errorf("createTkeCluster[%s] transNodeIPs for cluster[%s] failed: %v",
				taskID, info.Cluster.ClusterID, errTrans)
			retErr := fmt.Errorf("createTkeCluster transNodeIPs err, %s", errTrans)
			return nil, retErr
		}

		for i := range nodes {
			nodeIDs = append(nodeIDs, nodes[i].NodeID)
		}
	}

	clusterCidr, err := generateClusterCIDRInfo(ctx, info.Cluster, info.CmOption)
	if err != nil {
		return nil, err
	}

	// cluster create request
	req := &api.CreateClusterRequest{
		AddNodeMode:     info.Cluster.AutoGenerateMasterNodes,
		Region:          info.Cluster.Region,
		ClusterType:     info.Cluster.ManageType,
		ClusterCIDR:     clusterCidr,
		ClusterBasic:    generateClusterBasicInfo(info.Cluster, imageID, operator),
		ClusterAdvanced: generateClusterAdvancedInfo(info.Cluster),
		InstanceAdvanced: business.GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   strings.Join(nodeIPs, ","),
			Operator: operator,
			Render:   true,
		}, &business.NodeAdvancedOptions{NodeScheduler: true}),
		ExistedInstancesForNode: nil,
		RunInstancesForNode:     nil,
	}

	// 独立集群 和 托管集群
	switch info.Cluster.ManageType {
	case icommon.ClusterManageTypeIndependent:
		if req.AddNodeMode {
			req.RunInstancesForNode = []*api.RunInstancesForNode{
				generateRunInstance(info.Cluster, api.MASTER_ETCD.String(), passwd),
			}
		} else {
			err = handleClusterMasterNodes(ctx, req, info, passwd, masterIDs)
			if err != nil {
				blog.Errorf("createTkeCluster[%s] handleClusterMasterNodes for cluster[%s] failed: %v",
					taskID, info.Cluster.ClusterID, err)
				return nil, err
			}

			err = handleClusterWorkerNodes(ctx, req, info, passwd, nodeIDs, operator)
			if err != nil {
				blog.Errorf("createTkeCluster[%s] handleClusterWorkerNodes for cluster[%s] failed: %v",
					taskID, info.Cluster.ClusterID, err)
				return nil, err
			}
		}
	case icommon.ClusterManageTypeManaged:
		if req.AddNodeMode {
			req.RunInstancesForNode = []*api.RunInstancesForNode{
				generateRunInstance(info.Cluster, api.WORKER.String(), passwd),
			}
		} else {
			err = handleClusterWorkerNodes(ctx, req, info, passwd, nodeIDs, operator)
			if err != nil {
				blog.Errorf("createTkeCluster[%s] createClusterReq for cluster[%s] failed: %v",
					taskID, info.Cluster.ClusterID, err)
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("createTkeCluster[%s] not support manageType[%s]", taskID, info.Cluster.ManageType)
	}

	// handle default addon parameters
	req.Addons = handleTkeDefaultExtensionAddons(ctx, info.CmOption)

	// check cluster if exist
	systemID := info.Cluster.SystemID
	if systemID != "" {
		tkeCluster, errGet := tkeCli.GetTKECluster(info.Cluster.SystemID)
		if errGet != nil {
			blog.Errorf("createTkeCluster[%s] GetTKECluster[%s] failed, %s",
				taskID, info.Cluster.ClusterID, errGet.Error())
			retErr := fmt.Errorf("call GetTKECluster[%s] api err, %s", info.Cluster.ClusterID, errGet.Error())
			return nil, retErr
		}
		systemID = *tkeCluster.ClusterId
	} else {
		resp, errCreate := tkeCli.CreateTKECluster(req)
		if errCreate != nil {
			blog.Errorf("createTkeCluster[%s] call CreateTKECluster[%s] failed, %s",
				taskID, info.Cluster.ClusterID, errCreate.Error())
			retErr := fmt.Errorf("call CreateTKECluster[%s] api err, %s", info.Cluster.ClusterID, errCreate.Error())
			return nil, retErr
		}
		blog.Infof("createTkeCluster[%s] CreateTKECluster[%s] successful", taskID, info.Cluster.ClusterID)

		// update cluster systemID
		err = updateClusterSystemID(info.Cluster.ClusterID, resp.ClusterID)
		if err != nil {
			blog.Errorf("createTkeCluster[%s] updateClusterSystemID[%s] failed %s",
				taskID, info.Cluster.ClusterID, err.Error())
			retErr := fmt.Errorf("call CreateTKECluster updateClusterSystemID[%s] api err: %s",
				info.Cluster.ClusterID, err.Error())
			return nil, retErr
		}
		blog.Infof("createTkeCluster[%s] call CreateTKECluster updateClusterSystemID successful", taskID)
		systemID = resp.ClusterID
	}

	return &clusterInfo{
		systemID:  systemID,
		masterIPs: masterIPs,
		masterIDs: masterIDs,
		nodeIPs:   nodeIPs,
		nodeIDs:   nodeIDs,
	}, nil
}

// CreateTkeClusterTask call qcloud interface to create cluster
func CreateTkeClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateTkeClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateTkeClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")
	passwd := state.Task.CommonParams[cloudprovider.PasswordKey.String()]
	operator := state.Task.CommonParams[cloudprovider.OperatorKey.String()]
	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:      clusterID,
		CloudID:        cloudID,
		NodeTemplateID: nodeTemplateID,
	})
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// create cluster task
	cls, err := createTkeCluster(ctx, dependInfo, nodeIPs, passwd, operator)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s] createTkeCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createTkeCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = cls.systemID
	state.Task.CommonParams[cloudprovider.MasterIPs.String()] = strings.Join(cls.masterIPs, ",")
	state.Task.CommonParams[cloudprovider.MasterIDs.String()] = strings.Join(cls.masterIDs, ",")
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(cls.nodeIPs, ",")
	state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(cls.nodeIDs, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateTkeClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get tke client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		return retErr
	}

	var (
		abnormal = false
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := cli.GetTKECluster(systemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] GetTKECluster failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, *cluster.ClusterStatus)

		switch *cluster.ClusterStatus {
		case api.ClusterStatusRunning:
			return loop.EndLoop
		case api.ClusterStatusAbnormal:
			abnormal = true
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	if abnormal {
		blog.Errorf("checkClusterStatus[%s] GetTKECluster[%s] failed: abnormal", taskID, info.Cluster.ClusterID)
		retErr := fmt.Errorf("cluster[%s] status abnormal", info.Cluster.ClusterID)
		return retErr
	}

	return nil
}

// CheckTkeClusterStatusTask check cluster create status
func CheckTkeClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckTkeClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckTkeClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo, systemID)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// enableTkeClusterVpcCni enable tke cluster vpc-cni mode
func enableTkeClusterVpcCni(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("enableTkeClusterVpcCni[%s] getTkeClient cluster[%s] failed: %v",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		return retErr
	}

	blog.Infof("enableTkeClusterVpcCni[%s]: enableVPCCni %v", taskID, info.Cluster.NetworkSettings.EnableVPCCni)

	var (
		abnormal = false
	)

	if info.Cluster.NetworkSettings.EnableVPCCni {
		err = cli.EnableTKEVpcCniMode(&api.EnableVpcCniInput{
			TkeClusterID:   systemID,
			VpcCniType:     api.TKEDirectEni,
			SubnetsIDs:     info.Cluster.NetworkSettings.EniSubnetIDs,
			EnableStaticIp: info.Cluster.NetworkSettings.IsStaticIpMode,
			ExpiredSeconds: int(info.Cluster.NetworkSettings.ClaimExpiredSeconds),
		})
		if err != nil {
			blog.Errorf("enableTkeClusterVpcCni[%s] tke EnableTKEVpcCniMode for cluster[%s] failed: %v",
				taskID, info.Cluster.ClusterID, err)
			retErr := fmt.Errorf("EnableTKEVpcCniMode failed: %s", err.Error())
			return retErr
		}

		ctxTime, cancel := context.WithTimeout(context.Background(), time.Minute*30)
		defer cancel()

		err = loop.LoopDoFunc(ctxTime, func() error {
			status, errGet := cli.GetEnableVpcCniProgress(systemID)
			if errGet != nil {
				blog.Errorf("enableTkeClusterVpcCni[%s] GetEnableVpcCniProgress failed: %v", taskID, errGet)
				return nil
			}

			blog.Infof("enableTkeClusterVpcCni[%s]: GetEnableVpcCniProgress current status[%s]",
				taskID, status.Status)

			switch status.Status {
			case string(api.Succeed):
				return loop.EndLoop
			case string(api.Failed):
				abnormal = true
				return loop.EndLoop
			}

			return nil
		}, loop.LoopInterval(time.Second*5))
		if err != nil {
			blog.Errorf("enableTkeClusterVpcCni[%s] GetEnableVpcCniProgress failed: %v", taskID, err)
			return err
		}
		if abnormal {
			blog.Errorf("enableTkeClusterVpcCni[%s] GetEnableVpcCniProgress status abnormal", taskID)
			retErr := fmt.Errorf("GetEnableVpcCniProgress[%s] api timeout|abnormal", info.Cluster.ClusterID)
			return retErr
		}
	}

	return nil
}

// EnableTkeClusterVpcCniTask enable on vpc-cni networkMode
func EnableTkeClusterVpcCniTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("EnableTkeClusterVpcCniTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("EnableTkeClusterVpcCniTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// enableTkeClusterVpcCni task
	err = enableTkeClusterVpcCni(ctx, dependInfo, systemID)
	if err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s] enableTkeClusterVpcCni failed: %v",
			taskID, err)
		retErr := fmt.Errorf("enableTkeClusterVpcCni[%s] abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnableTkeClusterVpcCniTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckCreateClusterNodeStatusTask check cluster node status
func CheckCreateClusterNodeStatusTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckCreateClusterNodeStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckCreateClusterNodeStatusTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// get previous step paras
	nodeIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster nodes status
	addSuccessNodes, addFailureNodes, err := business.CheckClusterInstanceStatus(ctx, dependInfo, nodeIDs)
	if err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] CheckClusterInstanceStatus failed, %s",
			taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterInstanceStatus failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CheckCreateClusterNodeStatusTask[%s] addSuccessNodes[%v] addFailureNodes[%v]",
		taskID, addSuccessNodes, addFailureNodes)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// RegisterManageClusterKubeConfigTask register cluster kubeconfig
func RegisterManageClusterKubeConfigTask(taskID string, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterManageClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterManageClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	subnetID := step.Params[cloudprovider.SubnetIDKey.String()]
	isExtranet := step.Params[cloudprovider.IsExtranetKey.String()]

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// cluster vpc subnet selection
	subnet := subnetID
	if subnet == "" {
		subnet, err = getRandomSubnetByVpcID(ctx, dependInfo)
		if err != nil {
			blog.Errorf("RegisterManageClusterKubeConfigTask[%s] getRandomSubnetByVpcID failed: %s", taskID, err.Error())
			retErr := fmt.Errorf("getRandomSubnetByVpcID failed, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}
	blog.Infof("RegisterManageClusterKubeConfigTask[%s] subnet[%s]", taskID, subnet)

	// open tke internal kubeconfig
	err = registerTKEClusterEndpoint(ctx, dependInfo, api.ClusterEndpointConfig{
		IsExtranet: false,
		SubnetId:   subnet,
	})
	if err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s] registerTKEClusterEndpoint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("registerTKEClusterEndpoint failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterManageClusterKubeConfigTask[%s] registerTKEClusterEndpoint success", taskID)

	// 开启admin权限, 并生成kubeconfig
	clusterKube, connectKube, err := openClusterAdminKubeConfig(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s] registerTKEClusterEndpoint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("registerTKEClusterEndpoint failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterManageClusterKubeConfigTask[%s] openClusterAdminKubeConfig[%s] [%s] success",
		taskID, clusterKube, connectKube)

	// retry 重试生成jwt token
	var (
		token string
	)
	err = retry.Do(func() error {
		token, err = providerutils.GenerateSATokenByKubeConfig(ctx, connectKube)
		if err != nil {
			return err
		}
		blog.Infof("RegisterManageClusterKubeConfigTask[%s] GenerateSAToken[%s] success", taskID, token)

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(3*time.Second))
	if err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s] GenerateSAToken failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("GenerateSAToken failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterManageClusterKubeConfigTask[%s] GenerateSAToken[%s] success", taskID, token)

	// import cluster credential
	err = importClusterCredential(ctx, dependInfo, func() bool {
		return isExtranet == icommon.True
	}(), false, token, clusterKube)
	if err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("RegisterManageClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// dynamic inject paras
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.DynamicClusterKubeConfigKey.String()] = clusterKube

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterManageClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// getRandomSubnetByVpcID get random subnet by vpcID
func getRandomSubnetByVpcID(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	vpcClient, err := api.NewVPCClient(info.CmOption)
	if err != nil {
		blog.Errorf("getRandomSubnetByVpcID[%s] newVpcClient failed: %v", taskID, err)
		return "", err
	}

	// filter vpc subnets
	filter := make([]*api.Filter, 0)
	filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{info.Cluster.VpcID}})
	subnets, err := vpcClient.DescribeSubnets(nil, filter)
	if err != nil {
		blog.Errorf("getRandomSubnetByVpcID[%s] failed: %v", taskID, err)
		return "", err
	}

	// pick available subnet
	availableSubnet := make([]*vpc.Subnet, 0)
	for i := range subnets {
		match := utils.MatchPatternSubnet(*subnets[i].SubnetName, info.Cluster.Region)
		if match && *subnets[i].AvailableIpAddressCount > 0 {
			availableSubnet = append(availableSubnet, subnets[i])
		}
	}
	if len(availableSubnet) == 0 {
		return "", fmt.Errorf("region[%s] vpc[%s]无可用匹配子网", info.Cluster.Region, info.Cluster.VpcID)
	}

	rand.Seed(time.Now().Unix())                                           // nolint
	return *availableSubnet[rand.Intn(len(availableSubnet))].SubnetId, nil // nolint
}

// openClusterAdminKubeConfig open account cluster admin perm
func openClusterAdminKubeConfig(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (string, string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] NewTkeClient failed: %v", taskID, err)
		return "", "", err
	}

	// get qcloud account clusterAdminRole
	err = tkeCli.AcquireClusterAdminRole(info.Cluster.SystemID)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] AcquireClusterAdminRole failed: %v", taskID, err)
		return "", "", err
	}

	// get qcloud cluster endpoint
	ep, err := tkeCli.DescribeClusterEndpoints(info.Cluster.SystemID)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] DescribeClusterEndpoints failed: %v", taskID, err)
		return "", "", err
	}

	if ep.ClusterIntranetDomain == "" {
		clusterKube, errLocal := tkeCli.GetTKEClusterKubeConfig(info.Cluster.SystemID, false)
		if errLocal != nil {
			return "", "", errLocal
		}

		return clusterKube, clusterKube, errLocal
	}

	kube, err := tkeCli.GetTKEClusterKubeConfig(info.Cluster.SystemID, false)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] GetTKEClusterKubeConfig failed: %v", taskID, err)
		return "", "", err
	}
	kubeConfig, _ := base64.StdEncoding.DecodeString(kube)
	// parse kubeConfig to Config
	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{YamlContent: string(kubeConfig)})
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] GetKubeConfigFromYAMLBody failed: %v", taskID, err)
		return "", "", err
	}

	if len(config.Clusters) == 0 {
		return "", "", fmt.Errorf("openClusterAdminKubeConfig[%s] yamlConfig[%s] cluster emptp",
			taskID, info.Cluster.SystemID)
	}

	// cluster kubeConfig server by server IP address
	if strings.Contains(config.Clusters[0].Cluster.Server, ep.ClusterIntranetDomain) {
		config.Clusters[0].Cluster.Server = fmt.Sprintf("https://%s", ep.ClusterIntranetEndpoint)
	}
	clusterKubeBytes, _ := yaml.Marshal(config)

	config.Clusters[0].Cluster.InsecureSkipTLSVerify = true
	config.Clusters[0].Cluster.CertificateAuthority = ""
	config.Clusters[0].Cluster.CertificateAuthorityData = []byte("")

	connectKubeBytes, _ := yaml.Marshal(config)

	return base64.StdEncoding.EncodeToString(clusterKubeBytes), base64.StdEncoding.EncodeToString(connectKubeBytes), nil
}

// UpdateCreateClusterDBInfoTask update cluster DB info
func UpdateCreateClusterDBInfoTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateCreateClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]
	nodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	for _, ip := range nodes {
		node, errGet := cloudprovider.GetStorageModel().GetNodeByIP(context.Background(), ip)
		if errGet != nil {
			blog.Errorf("UpdateCreateClusterDBInfoTask[%s] GetNodeByIP[%s] failed: %v",
				taskID, ip, errGet)
			// no import node when found err
			continue
		}
		node.Status = icommon.StatusRunning

		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("UpdateCreateClusterDBInfoTask[%s] UpdateNode[%s] failed: %v",
				taskID, ip, err)
		}
	}

	// sync clusterData to pass-cc
	providerutils.SyncClusterInfoToPassCC(taskID, dependInfo.Cluster)

	// sync cluster perms
	providerutils.AuthClusterResourceCreatorPerm(ctx, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.ClusterName, dependInfo.Cluster.Creator)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}
