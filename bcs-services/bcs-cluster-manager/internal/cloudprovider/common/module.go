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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"github.com/kirito41dd/xslice"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	defaultLimit = 10
)

var (
	transferHostModuleStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.TransferHostModuleAction,
		StepName:   "转移主机模块",
	}

	removeHostFromCmdbStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.RemoveHostFromCmdbAction,
		StepName:   "移除主机",
	}

	checkNodeIpsInCmdbStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.CheckNodeIpsInCmdbAction,
		StepName:   "检测节点同步至cmdb",
	}
)

// BuildTransferHostModuleStep build common transfer module step
func BuildTransferHostModuleStep(task *proto.Task, businessID string, moduleID string, masterModuleID string) {
	transStep := cloudprovider.InitTaskStep(transferHostModuleStep)

	transStep.Params[cloudprovider.BKBizIDKey.String()] = businessID
	transStep.Params[cloudprovider.BKModuleIDKey.String()] = moduleID
	transStep.Params[cloudprovider.BKMasterModuleIDKey.String()] = masterModuleID

	task.Steps[transferHostModuleStep.StepMethod] = transStep
	task.StepSequence = append(task.StepSequence, transferHostModuleStep.StepMethod)
}

// BuildRemoveHostStep build common remove host from cmdb step
func BuildRemoveHostStep(task *proto.Task, bizID string, nodeIPs []string) {
	removeStep := cloudprovider.InitTaskStep(removeHostFromCmdbStep, cloudprovider.WithStepSkipFailed(true))

	removeStep.Params[cloudprovider.BKBizIDKey.String()] = bizID
	removeStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	task.Steps[removeHostFromCmdbStep.StepMethod] = removeStep
	task.StepSequence = append(task.StepSequence, removeHostFromCmdbStep.StepMethod)
}

// BuildCheckNodeIpsInCmdbStep check node ips sync to cmdb step
func BuildCheckNodeIpsInCmdbStep(task *proto.Task, cluster *proto.Cluster) {
	checkCmdbStep := cloudprovider.InitTaskStep(checkNodeIpsInCmdbStep)

	checkCmdbStep.Params[cloudprovider.CloudIDKey.String()] = cluster.Provider
	checkCmdbStep.Params[cloudprovider.ClusterIDKey.String()] = cluster.ClusterID

	task.Steps[checkNodeIpsInCmdbStep.StepMethod] = checkCmdbStep
	task.StepSequence = append(task.StepSequence, checkNodeIpsInCmdbStep.StepMethod)
}

// TransferHostModuleTask transfer host module task
func TransferHostModuleTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start transfer host module")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get nodeIPs
	nodeIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")
	// get moduleID
	moduleIDString := step.Params[cloudprovider.BKModuleIDKey.String()]

	// get moduleID
	masterModuleIDString := step.Params[cloudprovider.BKMasterModuleIDKey.String()]
	masterIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.MasterNodeIPsKey.String(), ",")

	if len(nodeIPs) == 0 {
		blog.Warnf("TransferHostModule %s skip, cause of empty node", taskID)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}

	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("TransferHostModule %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}
	moduleID, err := strconv.Atoi(moduleIDString)
	if err != nil {
		blog.Errorf("TransferHostModule %s failed, invalid moduleID, err %s", taskID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("invalid moduleID, err %s", err.Error()))
		return nil
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	// check exist master nodes, trans master nodes module if exist
	if len(masterModuleIDString) != 0 && len(masterIPs) > 0 {
		masterModuleID, _ := strconv.Atoi(masterModuleIDString)
		err = TransBizNodeModule(ctx, bkBizID, masterModuleID, masterIPs)
		if err != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("transfer master host module failed [%d]", err))
			blog.Errorf("TransferHostModule transBizNodeModule master[%v] failed: %v", masterIPs, err)
		}

		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			"transfer master host module successful")
	}

	// transfer nodes
	err = TransBizNodeModule(ctx, bkBizID, moduleID, func() []string {
		filterNodeIps := make([]string, 0)
		for i := range nodeIPs {
			if utils.StringInSlice(nodeIPs[i], masterIPs) {
				continue
			}

			filterNodeIps = append(filterNodeIps, nodeIPs[i])
		}

		return filterNodeIps
	}())
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("transfer host module failed [%s]", err))
		blog.Errorf("TransferHostModule %s failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, nodeIPs, err.Error())
		_ = state.UpdateStepFailure(start, stepName,
			fmt.Errorf("TransferHostModule failed, bkBizID %d, hosts %v, err %s", bkBizID, nodeIPs, err.Error()))
		return nil
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"transfer host module successful")

	blog.Infof("TransferHostModule %s successful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// TransBizNodeModule trans hostIPs to module. if module is zero, thus trans hostIPs to idle module
func TransBizNodeModule(ctx context.Context, biz, module int, hostIPs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("TransBizNodeModule %s failed, cmdb client is not init", taskID)
		return nil
	}

	// get host id from host list
	var hostIDs []int

	// 要从 bkcc 获取 hostID
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	err := retry.Do(func() error {
		var errGet error
		/*
			// hostIPs may be notIn biz cmdb && only operate exist hosts
			hostIDs, errGet = nodeManClient.GetHostIDByIPs(biz, hostIPs)
		*/

		hosts, errGet := cmdbClient.FetchAllHostsByBizID(biz, false)
		if errGet != nil {
			blog.Errorf("TransBizNodeModule %v failed, cmdb fetchAllHostsByBizID err %s", biz, errGet.Error())
			return errGet
		}
		for i := range hosts {
			if utils.StringInSlice(hosts[i].BKHostInnerIP, hostIPs) {
				hostIDs = append(hostIDs, int(hosts[i].BKHostID))
			}
		}

		blog.Infof("TransBizNodeModule %s get hosts id success", taskID)
		return nil
	}, retry.Attempts(3), retry.Context(ctx), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second))
	if err != nil {
		blog.Errorf("TransBizNodeModule %s get host id failed: %v", taskID, err)
		return err
	}

	blog.Infof("TransBizNodeModule %s hostIPs(%v) %+v hostIds(%v) %+v",
		taskID, len(hostIPs), hostIPs, len(hostIDs), hostIDs)

	err = cmdbClient.TransferHostToIdleModule(biz, hostIDs)
	if err != nil {
		blog.Errorf("TransBizNodeModule %s failed, bkBizID %d, hosts %v, err %s",
			taskID, biz, hostIDs, err.Error())
		return err
	}

	if module > 0 {
		err = cmdbClient.TransferHostModule(biz, hostIDs, []int{module}, false)
		if err != nil {
			blog.Errorf("TransBizNodeModule %s failed, bkBizID %d, hosts %v, err %s",
				taskID, biz, hostIDs, err.Error())
			return err
		}
	}

	return nil
}

// RemoveHostFromCMDBTask remove host from cmdb task
func RemoveHostFromCMDBTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"remove host from cmdb")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]

	if len(nodeIPs) == 0 {
		blog.Infof("RemoveHostFromCMDBTask %s skip, cause of empty node", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}
	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	err = RemoveHostFromCmdb(ctx, bkBizID, nodeIPs)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("remove host from cmdb failed [%s]", err))
		blog.Errorf("RemoveHostFromCmdb[%s] failed: %v", taskID, err)
		_ = state.SkipFailure(start, stepName, err)
		return nil
	}
	blog.Infof("RemoveHostFromCMDBTask %s successful", taskID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"remove host from cmdb successful")

	// update step
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}

// RemoveHostFromCmdb remove host from cmdb
func RemoveHostFromCmdb(ctx context.Context, biz int, nodeIPs string) error {
	taskID, stepName := cloudprovider.GetTaskIDAndStepNameFromContext(ctx)

	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, nodeman client is not init", taskID)
		return fmt.Errorf("nodeman client is not init")
	}
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, cmdb client is not init", taskID)
		return fmt.Errorf("cmdb client is not init")
	}

	// get host id from host list
	ips := strings.Split(nodeIPs, ",")
	hostIDs, err := nodeManClient.GetHostIDByIPs(biz, ips)
	if err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, list nodeman hosts err %s", taskID, err.Error())
		return fmt.Errorf("list nodeman hosts err %s", err.Error())
	}

	if len(hostIDs) == 0 {
		blog.Warnf("RemoveHostFromCMDBTask %s skip, cause of empty host", taskID)
		return nil
	}

	if err := cmdbClient.TransferHostToIdleModule(biz, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s",
			taskID, biz, hostIDs, err.Error())
		return fmt.Errorf("TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s",
			biz, hostIDs, err.Error())
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"transfer host to idle module successful")

	if err := cmdbClient.TransferHostToResourceModule(biz, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
			taskID, biz, hostIDs, err.Error())
		return fmt.Errorf("TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
			biz, hostIDs, err.Error())
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"transfer host to resource module successful")

	if err := cmdbClient.DeleteHost(hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s DeleteHost %v failed, %s", taskID, hostIDs, err.Error())
		return fmt.Errorf("DeleteHost %v failed, %s", hostIDs, err.Error())
	}

	return nil
}

// CheckNodeIpsInCMDBTask check nodes exist in cmdb task
func CheckNodeIpsInCMDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get nodeIPs
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.GetCommonParams(),
		cloudprovider.NodeIPsKey.String(), ",")
	if len(ipList) == 0 {
		blog.Infof("CheckNodeIpsInCMDBTask[%s] nodeIPs empty", taskID)
		return nil
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	err = CheckIPsInCmdb(ctx, ipList)
	if err != nil {
		blog.Errorf("CheckNodeIpsInCMDBTask[%s] failed: %v", taskID, err)
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}
	blog.Infof("CheckNodeIpsInCMDBTask %s successful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}

// CheckIPsInCmdb check cluster nodeIPs sync to cmdb
func CheckIPsInCmdb(ctx context.Context, nodeIPs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	var err error
	// check nodeIPs if exist in cmdb
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		cmdbClient := cmdb.GetCmdbClient()
		if cmdbClient == nil {
			blog.Errorf("checkIPsInCmdb[%s] failed, cmdb client is not init", taskID)
			return nil
		}
		detailHosts, errLocal := cmdbClient.QueryAllHostInfoWithoutBiz(nodeIPs)
		if errLocal != nil {
			blog.Errorf("checkIPsInCmdb[%s] QueryAllHostInfoWithoutBiz failed: %s", taskID, errLocal.Error())
			return nil
		}

		blog.Infof("checkIPsInCmdb[%s] QueryAllHostInfoWithoutBiz sourceIps(%v) cmdb(%v)",
			taskID, len(nodeIPs), len(detailHosts))

		if len(detailHosts) == len(nodeIPs) {
			return loop.EndLoop
		}
		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkIPsInCmdb[%s] failed: %v", taskID, err)
		return err
	}

	return nil
}

// HostInfo host info
type HostInfo struct {
	HostId    int64
	HostIp    string
	BkCloudId int
}

// return Host Ids
func returnHostIds(hosts []HostInfo) []int64 {
	hostIds := make([]int64, 0)

	for i := range hosts {
		hostIds = append(hostIds, hosts[i].HostId)
	}
	return hostIds
}

// return Host Ips
func returnHostIps(hosts []HostInfo) []string {
	hostIps := make([]string, 0)

	for i := range hosts {
		hostIps = append(hostIps, hosts[i].HostIp)
	}
	return hostIps
}

// ip In Host Infos
func ipInHostInfos(ip string, hosts []HostInfo) bool {
	for i := range hosts {
		if hosts[i].HostIp == ip {
			return true
		}
	}
	return false
}

// SplitHostsChunks split hosts chunk
func SplitHostsChunks(hostList []HostInfo, limit int) [][]HostInfo {
	if limit <= 0 || len(hostList) == 0 {
		return nil
	}
	i := xslice.SplitToChunks(hostList, limit)
	ss, ok := i.([][]HostInfo)
	if !ok {
		return nil
	}

	return ss
}

// SyncIpsInfoToCmdb sync ips info to cmdb
func SyncIpsInfoToCmdb(ctx context.Context, dependInfo *cloudprovider.CloudDependBasicInfo, nodeIPs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	inCmdb, notInCmdb, err := splitNodeIPsFromCmdb(ctx, nodeIPs)
	if err != nil {
		return err
	}

	resourcePoolType := ""
	if dependInfo.NodeGroup.GetExtraInfo() != nil {
		t, ok := dependInfo.NodeGroup.GetExtraInfo()[resource.ResourcePoolType]
		if ok {
			resourcePoolType = t
		}
	}

	blog.Infof("SyncIpsInfoToCmdb[%s] resourceType[%s]", resourcePoolType)

	switch resourcePoolType {
	case resource.SelfPool:
		if len(notInCmdb) > 0 {
			err = handleInCmdbFromCmpyNodeIps(ctx, notInCmdb)
			if err != nil {
				blog.Errorf("SyncIpsInfoToCmdb[%s] handleNotInCmdbNodeIps failed: %v", taskID, err)
			}
		}
	default:
		if len(notInCmdb) > 0 {
			err = handleNotInCmdbNodeIps(ctx, notInCmdb)
			if err != nil {
				blog.Errorf("SyncIpsInfoToCmdb[%s] handleNotInCmdbNodeIps failed: %v", taskID, err)
			}
		}
	}

	if len(inCmdb) > 0 {
		err = handleInCmdbNodeIps(ctx, inCmdb)
		if err != nil {
			blog.Errorf("task[%s] SyncIpsInfoToCmdb handleInCmdbNodeIps failed: %v", taskID, err)
		}
	}

	return nil
}

// split Node IPs From Cmdb
func splitNodeIPsFromCmdb(ctx context.Context, nodeIPs []string) ([]HostInfo, []HostInfo, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	var (
		nodeInCmdb    = make([]HostInfo, 0)
		nodeNotInCmdb = make([]HostInfo, 0)
	)

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("checkIPsInCmdb[%s] failed, cmdb client is not init", taskID)
		return nil, nil, fmt.Errorf("cmdbClient is not init")
	}
	detailHosts, errLocal := cmdbClient.QueryAllHostInfoWithoutBiz(nodeIPs)
	if errLocal != nil {
		blog.Errorf("checkIPsInCmdb[%s] QueryAllHostInfoWithoutBiz failed: %s", taskID, errLocal.Error())
		return nil, nil, errLocal
	}

	// nodeInCmdb nodeIPs
	for i := range detailHosts {
		nodeInCmdb = append(nodeInCmdb, HostInfo{
			HostId:    detailHosts[i].BKHostID,
			HostIp:    detailHosts[i].BKHostInnerIP,
			BkCloudId: detailHosts[i].BKHostCloudID,
		})
	}

	blog.Infof("task[%s] splitNodeIPsFromCmdb[%v] nodeInCmdb[%v]", taskID, len(nodeInCmdb), nodeInCmdb)

	for _, ip := range nodeIPs {
		if ipInHostInfos(ip, nodeInCmdb) {
			continue
		}

		nodeNotInCmdb = append(nodeNotInCmdb, HostInfo{
			HostIp: ip,
		})
	}
	blog.Infof("task[%s] splitNodeIPsFromCmdb[%v] nodeNotInCmdb[%v]", taskID, len(nodeNotInCmdb), nodeNotInCmdb)

	return nodeInCmdb, nodeNotInCmdb, nil
}

// handle In Cmdb From Cmpy Node Ips
func handleInCmdbFromCmpyNodeIps(ctx context.Context, inCmdbIps []HostInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("handleInCmdbFromCmpyNodeIps[%s] failed, cmdb client is not init", taskID)
		return fmt.Errorf("cmdbClient is not init")
	}
	hostIps := returnHostIps(inCmdbIps)

	blog.Infof("handleInCmdbFromCmpyNodeIps[%s] hostIps[%v]", taskID, hostIps)
	servers, err := cmdbClient.GetAssetIdsByIps(hostIps)
	if err != nil {
		blog.Errorf("handleInCmdbFromCmpyNodeIps[%s] failed: %v", taskID, err)
		return err
	}
	// 固资号
	assetIds := make([]string, 0)
	for _, s := range servers {
		assetIds = append(assetIds, s.ServerAssetId)
	}
	blog.Infof("handleInCmdbFromCmpyNodeIps[%s] assetIds[%v]", taskID, assetIds)

	// hostIds
	hosts, err := cmdbClient.QueryAllHostInfoByAssetIdWithoutBiz(assetIds)
	if err != nil {
		blog.Errorf("handleInCmdbFromCmpyNodeIps[%s] failed: %v", taskID, err)
		return err
	}

	// 主机ID
	hostIds := make([]int64, 0)
	for i := range hosts {
		hostIds = append(hostIds, hosts[i].BKHostID)
	}
	blog.Infof("handleInCmdbFromCmpyNodeIps[%s] hostIds[%v]", taskID, hostIds)

	// 同步公司cmdb信息至bkcc
	hostIdsChunks := utils.SplitInt64sChunks(hostIds, defaultLimit)
	for i := range hostIdsChunks {
		if len(hostIdsChunks[i]) == 0 {
			continue
		}

		errLocal := cmdbClient.SyncHostInfoFromCmpy(0, hostIdsChunks[i])
		if errLocal != nil {
			blog.Errorf("handleInCmdbFromCmpyNodeIps[%s] [%v] failed: %v", taskID, hostIdsChunks[i], err)
			continue
		}

		blog.Infof("handleInCmdbFromCmpyNodeIps[%s] [%v] success", taskID, hostIdsChunks[i])
	}

	return nil
}

// handle In Cmdb Node Ips
func handleInCmdbNodeIps(ctx context.Context, inCmdbIps []HostInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("handleInCmdbNodeIps[%s] failed, cmdb client is not init", taskID)
		return fmt.Errorf("cmdbClient is not init")
	}

	hostsChunks := SplitHostsChunks(inCmdbIps, defaultLimit)
	for i := range hostsChunks {
		hostIds := returnHostIds(hostsChunks[i])
		hostIps := returnHostIps(hostsChunks[i])
		if len(hostIds) == 0 {
			continue
		}

		err := cmdbClient.SyncHostInfoFromCmpy(0, hostIds)
		if err != nil {
			blog.Errorf("handleInCmdbNodeIps[%s] [%v] failed: %v", taskID, hostIds, err)
			continue
		}

		blog.Infof("handleInCmdbNodeIps[%s] [%v] [%v] success", taskID, hostIds, hostIps)
	}

	return nil
}

// handle Not In Cmdb Node Ips
func handleNotInCmdbNodeIps(ctx context.Context, notInCmdbIps []HostInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("handleNotInCmdbNodeIps[%s] failed, cmdb client is not init", taskID)
		return fmt.Errorf("cmdbClient is not init")
	}

	hostsChunks := SplitHostsChunks(notInCmdbIps, defaultLimit)
	for i := range hostsChunks {
		hostIds := returnHostIds(hostsChunks[i])
		hostIps := returnHostIps(hostsChunks[i])
		if len(hostIds) == 0 {
			continue
		}

		err := cmdbClient.AddHostFromCmpy(nil, hostIps, nil)
		if err != nil {
			blog.Errorf("handleNotInCmdbNodeIps[%s] [%v] failed: %v", taskID, hostIps, err)
			continue
		}

		blog.Infof("handleNotInCmdbNodeIps[%s] [%v] success", taskID, hostIps)
	}

	return nil
}
