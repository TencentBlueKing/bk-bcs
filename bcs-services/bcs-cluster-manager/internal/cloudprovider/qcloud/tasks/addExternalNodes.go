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

// Package tasks xxx
package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	pcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ApplyExternalNodeMachinesTask from resource-manager service
func ApplyExternalNodeMachinesTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start apply instance from resource pool")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ApplyExternalNodeMachinesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ApplyExternalNodeMachinesTask[%s] run current step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	operator := step.Params[cloudprovider.OperatorKey.String()]
	scalingNum, err := strconv.Atoi(desiredNodes)
	if err != nil {
		blog.Errorf("ApplyExternalNodeMachinesTask[%s] strconv desiredNodes failed: %v", taskID, err)
		retErr := fmt.Errorf("strconv desiredNodes failed: %v", err.Error())
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// task basic dependInfo
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("ApplyExternalNodeMachinesTask[%s] GetClusterDependBasicInfo for NodeGroup %s to clean "+
			"Node in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// applyInstanceFromResourcePool apply machines from resource pool
	recordInstanceList, err := applyInstanceFromResourcePool(ctx, dependInfo, scalingNum, operator)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("apply instance from resource pool failed [%s]", err))
		blog.Errorf("ApplyExternalNodeMachinesTask[%s] applyInstanceFromResourcePool for NodeGroup %s step %s failed, %s",
			taskID, nodeGroupID, stepName, err.Error())
		retErr := fmt.Errorf("applyInstanceFromResourcePool failed: %s", err.Error())
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// record cluster nodes to db
	err = recordClusterExternalNodeToDB(ctx, dependInfo, state.Task, &RecordInstanceToDBOption{
		InstanceIPs: recordInstanceList.InstanceIPList,
		DeviceIDs:   recordInstanceList.DeviceIDList,
	})
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("record cluster external node to db failed [%s]", err))
		blog.Errorf("ApplyExternalNodeMachinesTask[%s] recordClusterExternalNodeToDB for NodeGroup %s step %s failed, %s",
			taskID, nodeGroupID, stepName, err.Error())
		retErr := fmt.Errorf("ApplyExternalNodeMachinesTask failed, %s", err.Error())
		// rollback nodes
		_, _ = destroyIDCDeviceList(ctx, dependInfo, recordInstanceList.DeviceIDList, operator)
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"record cluster external node to db successful")

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"apply instance from resource pool successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ApplyExternalNodeMachinesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// buildApplyIDCNodesRequest build resource request
func buildApplyIDCNodesRequest(group *proto.NodeGroup, operator string) *resource.ApplyInstanceReq {
	return &resource.ApplyInstanceReq{
		NodeType: resource.IDC,

		CPU:          group.GetLaunchTemplate().GetCPU(),
		Memory:       group.GetLaunchTemplate().GetMem(),
		Gpu:          group.GetLaunchTemplate().GetGPU(),
		Region:       group.GetRegion(),
		VpcID:        group.GetAutoScaling().GetVpcID(),
		ZoneList:     group.GetAutoScaling().GetZones(),
		SubnetList:   group.GetAutoScaling().GetSubnetIDs(),
		InstanceType: group.GetLaunchTemplate().GetInstanceType(),
		PoolID:       group.GetConsumerID(),
		Operator:     operator,
		Selector:     group.GetLaunchTemplate().GetSelector(),
	}
}

// applyInstanceFromResourcePool 申请机器
func applyInstanceFromResourcePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	desired int, operator string) (*RecordInstanceList, error) {

	// consume resource pool
	orderID, err := consumeDevicesFromResourcePool(ctx, info.NodeGroup, desired, operator)
	if err != nil {
		return nil, err
	}

	// check order state
	record, err := checkOrderStateFromResourcePool(ctx, orderID)
	if err != nil {
		return nil, err
	}
	record.OrderID = orderID

	return record, nil
}

// consumeDevicesFromResourcePool apply cvm instances to generate orderID form resource pool
func consumeDevicesFromResourcePool(
	ctx context.Context, group *proto.NodeGroup, nodeNum int, operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ctx = utils.WithTraceIDForContext(ctx, taskID)
	// apply instance
	resp, err := tresource.GetResourceManagerClient().ApplyInstances(ctx, nodeNum,
		buildApplyIDCNodesRequest(group, operator))
	if err != nil {
		blog.Errorf("consumeDevicesFromResourcePool[%s] ApplyInstances failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("consumeDevicesFromResourcePool[%s] success")
	return resp.OrderID, nil
}

// RecordInstanceList instances record
type RecordInstanceList struct {
	OrderID        string
	InstanceIPList []string
	InstanceIDList []string
	DeviceIDList   []string
}

// checkOrderStateFromResourcePool 检查资源池订单状态
func checkOrderStateFromResourcePool(ctx context.Context, orderID string) (*RecordInstanceList, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	ctx = utils.WithTraceIDForContext(ctx, taskID)
	result, err := tresource.GetResourceManagerClient().CheckOrderStatus(ctx, orderID)
	if err != nil {
		blog.Errorf("checkOrderStateFromResourcePool[%s] CheckOrderStatus[%s] failed: %v", taskID, orderID, err)
		return nil, err
	}

	// get IDC device instanceIPs
	if len(result.InstanceIPs) == 0 {
		retErr := fmt.Errorf("checkOrderStateFromResourcePool[%s] return instance empty", taskID)
		blog.Errorf(retErr.Error())
		return nil, retErr
	}

	return &RecordInstanceList{
		InstanceIPList: result.InstanceIPs,
		DeviceIDList:   result.ExtraIDs,
	}, nil
}

// RecordInstanceToDBOption instances db option
type RecordInstanceToDBOption struct {
	Password    string
	InstanceIDs []string
	InstanceIPs []string
	DeviceIDs   []string
}

// recordClusterExternalNodeToDB 录入第三方节点
func recordClusterExternalNodeToDB(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, task *proto.Task, opt *RecordInstanceToDBOption) error {
	var (
		nodes = make([]*proto.Node, 0)
		err   error
	)

	// deviceID Map To InstanceIP
	instanceToDeviceID := make(map[string]string)
	for i := range opt.InstanceIPs {
		if _, ok := instanceToDeviceID[opt.InstanceIPs[i]]; !ok {
			instanceToDeviceID[opt.InstanceIPs[i]] = opt.DeviceIDs[i]
		}
	}

	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// ListExternalNodesByIP list external nodes
	err = retry.Do(func() error {
		nodes, err = business.ListExternalNodesByIP(opt.InstanceIPs, &cloudprovider.ListNodesOption{
			Common: info.CmOption,
		})
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(10))
	if err != nil {
		blog.Errorf("recordClusterExternalNodeToDB[%s] failed: %v", taskID, err)
		return err
	}
	if len(nodes) == 0 {
		blog.Errorf("recordClusterExternalNodeToDB[%s] cmdb sync nodes failed: %v", taskID, opt.InstanceIPs)
		return errors.New("cmdb sync nodes failed")
	}

	// update response information to task common params
	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}
	if len(opt.InstanceIPs) > 0 && len(opt.DeviceIDs) > 0 {
		task.CommonParams[cloudprovider.DeviceIDsKey.String()] = strings.Join(opt.DeviceIDs, ",")
		// Job Result parameter
		task.NodeIPList = opt.InstanceIPs
		task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(opt.InstanceIPs, ",")
		task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(opt.InstanceIPs, ",")
		task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(opt.InstanceIDs, ",")
	}

	// save nodes to db
	for _, n := range nodes {
		n.ClusterID = info.NodeGroup.ClusterID
		n.NodeGroupID = info.NodeGroup.NodeGroupID
		n.Passwd = opt.Password
		n.Status = common.StatusInitialization
		n.DeviceID = instanceToDeviceID[n.InnerIP]
		err = cloudprovider.SaveNodeInfoToDB(ctx, n, true)
		if err != nil {
			blog.Errorf("recordClusterExternalNodeToDB[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, n.InnerIP, err)
		}
	}

	return nil
}

// destroyIDCDeviceList 销毁归还机器
func destroyIDCDeviceList(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, deviceList []string,
	operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	if info == nil || info.NodeGroup == nil || info.Cluster == nil || len(deviceList) == 0 {
		return "", fmt.Errorf("destroyIDCDeviceList[%s] lost validate info", taskID)
	}

	ctx = utils.WithTraceIDForContext(ctx, taskID)

	// destroy devices
	resp, err := tresource.GetResourceManagerClient().DestroyInstances(ctx, &resource.DestroyInstanceReq{
		PoolID:      info.NodeGroup.GetConsumerID(),
		SystemID:    info.Cluster.GetSystemID(),
		InstanceIDs: deviceList,
		Operator:    operator,
	})
	if err != nil {
		blog.Errorf("destroyIDCDeviceList[%s] DestroyInstances failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("destroyIDCDeviceList[%s] call DestroyInstances successfully, orders %v.", resp.OrderID)
	return resp.OrderID, nil
}

// CheckExternalNodesEmptyTask check external node empty
func CheckExternalNodesEmptyTask(taskID string, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckExternalNodesEmptyTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckExternalNodesEmptyTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckExternalNodesEmptyTask[%s] GetClusterDependBasicInfo failed, %s", taskID, err.Error())
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// task common params
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(
		state.Task.CommonParams, cloudprovider.NodeIPsKey.String(), ",")
	deviceIdList := cloudprovider.ParseNodeIpOrIdFromCommonMap(
		state.Task.CommonParams, cloudprovider.DeviceIDsKey.String(), ",")

	if len(ipList) == 0 || len(deviceIdList) == 0 {
		blog.Errorf("CheckExternalNodesEmptyTask[%s] split NodeIPsKey failed: %v", taskID, err)
		retErr := fmt.Errorf("split NodeIPsKey failed: %v", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// parse step params
	bizID := step.Params[cloudprovider.BkSopsBizIDKey.String()]
	templateID := step.Params[cloudprovider.BkSopsTemplateIDKey.String()]
	operator := step.Params[cloudprovider.BkSopsTemplateUserKey.String()]
	templateSource := step.Params[cloudprovider.BkSopsTemplateSourceKey.String()]
	constants := step.Params[cloudprovider.BkSopsConstantsKey.String()]
	taskName := state.Task.CommonParams[cloudprovider.TaskNameKey.String()]

	if bizID == "" || operator == "" || templateID == "" || taskName == "" || constants == "" {
		errMsg := fmt.Sprintf("CheckExternalNodesEmptyTask[%s] validateParameter failed", taskID)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("CheckExternalNodesEmptyTask err, %s", errMsg)
		_ = returnExternalNodes(ctx, dependInfo, ipList, deviceIdList)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// render constants dynamic value parameter
	consMap, err := pcommon.RenderDynamicParaToConstants(state.Task, constants)
	if err != nil {
		errMsg := fmt.Sprintf("CheckExternalNodesEmptyTask[%s] unmarshal constants failed[%v]", taskID, err)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("CheckExternalNodesEmptyTask err, %s", errMsg)
		_ = returnExternalNodes(ctx, dependInfo, ipList, deviceIdList)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	timeOutCtx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	// execBkSopsTask exec sops task
	taskUrl, err := pcommon.ExecBkSopsTask(timeOutCtx, pcommon.CreateBkSopsTaskParas{
		BizID:          bizID,
		TemplateID:     templateID,
		Operator:       operator,
		TemplateSource: templateSource,
		TaskName:       taskName,
		Constants:      consMap,
		StepName:       stepName,
	})
	if err != nil {
		state.TaskURL = taskUrl
		// rollback external nodes
		_ = returnExternalNodes(ctx, dependInfo, ipList, deviceIdList)
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}

	state.TaskURL = taskUrl
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}

// returnExternalNodes return external nodes to resource manager
func returnExternalNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	ips, deviceIds []string) error { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if info == nil || len(ips) == 0 || len(deviceIds) == 0 {
		blog.Infof("returnExternalNodes[%s] info null or ips/deviceIds empty", taskID)
		return nil
	}

	// delete db data record
	for _, ip := range ips {
		err := cloudprovider.GetStorageModel().DeleteClusterNodeByIP(context.Background(), info.Cluster.ClusterID, ip)
		if err != nil {
			blog.Errorf("returnExternalNodes[%s] DeleteClusterNodeByIP[%s] failed: %v", taskID,
				ip, err)
		} else {
			blog.Infof("returnExternalNodes[%s] DeleteClusterNodeByIP success[%+v]", taskID, ip)
		}
	}

	// destroy device to resource manager
	orderID, err := destroyIDCDeviceList(ctx, info, deviceIds, common.ClusterManager)
	if err != nil {
		blog.Errorf("returnExternalNodes[%s] destroyIDCDeviceList failed: %v", taskID, err)
	} else {
		blog.Infof("returnExternalNodes[%s] successful[%v] orderID[%v]", taskID, deviceIds, orderID)
	}

	// rollback nodeGroup desired size
	err = cloudprovider.UpdateNodeGroupDesiredSize(info.NodeGroup.NodeGroupID, len(ips), true)
	if err != nil {
		blog.Errorf("returnExternalNodes[%s] UpdateNodeGroupDesiredSize failed: %v", taskID, err)
	} else {
		blog.Infof("returnExternalNodes[%s] UpdateNodeGroupDesiredSize success[%v]", taskID, len(ips))
	}

	return nil
}

// GetExternalNodeScriptTask get cluster external node script
func GetExternalNodeScriptTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start get cluster external node script")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("GetExternalNodeScriptTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("GetExternalNodeScriptTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	var ipList []string
	if len(step.Params[cloudprovider.NodeIPsKey.String()]) > 0 {
		ipList = strings.Split(step.Params[cloudprovider.NodeIPsKey.String()], ",")
	} else if len(state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]) > 0 {
		ipList = strings.Split(state.Task.CommonParams[cloudprovider.NodeIPsKey.String()], ",")
	}
	if len(ipList) == 0 {
		blog.Errorf("GetExternalNodeScriptTask[%s] split NodeIPsKey failed: %v", taskID, err)
		retErr := fmt.Errorf("split NodeIPsKey failed: %v", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// cluster/cloud/nodeGroup/cloudCredential Info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("GetExternalNodeScriptTask[%s] GetClusterDependBasicInfo task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	clusterExternalNodes, err := business.FilterClusterExternalNodesByIPs(ctx, dependInfo, ipList)
	if err != nil {
		blog.Errorf("GetExternalNodeScriptTask[%s]: FilterClusterExternalInstanceFromNodesIPs for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("FilterClusterExternalInstanceFromNodesIPs err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("GetExternalNodeScriptTask[%s] existedNodeIPs[%v] notExistedNodeIPs[%v]",
		taskID, clusterExternalNodes.ExistInClusterIPs, clusterExternalNodes.NotExistInClusterIPs)

	if len(clusterExternalNodes.NotExistInClusterIPs) == 0 {
		blog.Errorf("GetExternalNodeScriptTask[%s]: nodeIPs all exist in cluster[%s]",
			taskID, clusterID)
		retErr := fmt.Errorf("GetExternalNodeScriptTask err, %s", "nodeIPs all exist in cluster")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get add external nodes script from cluster
	script, err := business.GetClusterExternalNodeScript(ctx, dependInfo,
		GetExternalNgScriptType(dependInfo.NodeGroup))
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("get cluster external node script failed [%s]", err))
		blog.Errorf("GetExternalNodeScriptTask[%s]: GetClusterExternalNodeScript for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("GetClusterExternalNodeScript err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("GetExternalNodeScriptTask[%s] cluster[%s] nodeIPs[%v] script[%v]",
		taskID, clusterID, clusterExternalNodes.NotExistInClusterIPs, script)

	// inject depend data
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	// dynamic inject paras
	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] =
		strings.Join(clusterExternalNodes.NotExistInClusterIPs, ",")
	state.Task.CommonParams[cloudprovider.DynamicNodeScriptKey.String()] = script

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"get cluster external node script successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("GetExternalNodeScriptTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
