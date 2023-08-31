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
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
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
)

// BuildTransferHostModuleStep build common transfer module step
func BuildTransferHostModuleStep(task *proto.Task, businessID string, moduleID string) {
	transStep := cloudprovider.InitTaskStep(transferHostModuleStep)

	transStep.Params[cloudprovider.BKBizIDKey.String()] = businessID
	transStep.Params[cloudprovider.BKModuleIDKey.String()] = moduleID

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

// TransferHostModuleTask transfer host module task
func TransferHostModuleTask(taskID string, stepName string) error {
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
	// get moduleID
	moduleIDString := step.Params[cloudprovider.BKModuleIDKey.String()]

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

	nodeManClient := nodeman.GetNodeManClient()

	if nodeManClient == nil {
		blog.Errorf("TransferHostModule %s failed, nodeman client is not init", taskID)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("TransferHostModule %s failed, cmdb client is not init", taskID)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("cmdb client is not init"))
		return nil
	}

	// get host id from host list
	ips := strings.Split(nodeIPs, ",")
	var hostIDs []int

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		var errGet error
		hostIDs, errGet = nodeManClient.GetHostIDByIPs(bkBizID, ips)
		if errGet != nil {
			blog.Errorf("TransferHostModule %v failed, list nodeman hosts err %s", bkBizID, errGet.Error())
			return errGet
		}
		if len(hostIDs) == len(ips) {
			return loop.EndLoop
		}
		blog.Infof("TransferHostModule %s can't get all host id, waiting", taskID)
		return nil
	}, loop.LoopInterval(3*time.Second))
	if err != nil {
		blog.Errorf("TransferHostModule %s get host id failed: %v", taskID, err)
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("get host id err %s", err.Error()))
		return nil
	}

	if err = cmdbClient.TransferHostModule(bkBizID, hostIDs, []int{moduleID}, false); err != nil {
		blog.Errorf("TransferHostModule %s failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.UpdateStepFailure(start, stepName,
			fmt.Errorf("TransferHostModule failed, bkBizID %d, hosts %v, err %s", bkBizID, hostIDs, err.Error()))
		return nil
	}

	blog.Infof("TransferHostModule %s successful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// RemoveHostFromCMDBTask remove host from cmdb task
func RemoveHostFromCMDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}
	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, nodeman client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, cmdb client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("cmdb client is not init"))
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

	// get host id from host list
	ips := strings.Split(nodeIPs, ",")
	hostIDs, err := nodeManClient.GetHostIDByIPs(bkBizID, ips)
	if err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, list nodeman hosts err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("list nodeman hosts err %s", err.Error()))
		return nil
	}

	if len(hostIDs) == 0 {
		blog.Warnf("RemoveHostFromCMDBTask %s skip, cause of empty host", taskID)
		_ = state.UpdateStepSucc(start, stepName)
		return nil
	}

	if err := cmdbClient.TransferHostToIdleModule(bkBizID, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName,
			fmt.Errorf("TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s", bkBizID, hostIDs, err.Error()))
		return nil
	}

	if err := cmdbClient.TransferHostToResourceModule(bkBizID, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName,
			fmt.Errorf("TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
				bkBizID, hostIDs, err.Error()))
		return nil
	}

	if err := cmdbClient.DeleteHost(hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s DeleteHost %v failed, %s", taskID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("DeleteHost %v failed, %s", hostIDs, err.Error()))
		return nil
	}
	blog.Infof("RemoveHostFromCMDBTask %s successful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}
