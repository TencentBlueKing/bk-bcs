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

// Package handler xxx
package handler

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/mitchellh/mapstructure"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	nodegroupmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
)

// NodegroupManager defines nodegroup manager handler
type NodegroupManager struct {
	storage storage.Storage
}

// New create grpc service handler
func New(storage storage.Storage) *NodegroupManager {
	return &NodegroupManager{
		storage: storage,
	}
}

// GetClusterAutoscalerReview handles cluster autoscaler review request
// get nodegroup scale policy, request by ca
// scale up: num
// scale down: num/ip
func (e *NodegroupManager) GetClusterAutoscalerReview(ctx context.Context,
	req *nodegroupmgr.ClusterAutoscalerReview, rsp *nodegroupmgr.ClusterAutoscalerReview) error {
	var err error
	startTime := time.Now()
	blog.Info("Received NodegroupManager.GetClusterAutoscalerReview request. RequestId:%s", req.Request.Uid)
	reqNodeGroups := req.Request.NodeGroups
	rsp.Request = req.Request
	rsp.Response = &nodegroupmgr.AutoscalerReviewResponse{}
	rsp.Response.Uid = req.Request.Uid
	scaleUpPolicies := make([]*nodegroupmgr.NodeScaleUpPolicy, 0)
	scaleDownPolicies := make([]*nodegroupmgr.NodeScaleDownPolicy, 0)
	defer func() {
		blog.Infof("caReview response: %v", rsp.Response)
		if err == nil {
			e.updateNodeGroupStatus(rsp)
		}
		metric.ReportAPIRequestMetric("grpc", "GetClusterAutoscalerReview", err, startTime)
	}()
	for nodegroupId := range reqNodeGroups {
		scaleUpPolicy, scaleDownPolicy, err := e.handleNodeGroup(reqNodeGroups[nodegroupId], req.Request.Uid)
		if err != nil {
			blog.Errorf("handleNodegroup error:%v", err)
			return err
		}
		if scaleUpPolicy != nil {
			scaleUpPolicies = append(scaleUpPolicies, scaleUpPolicy...)
		}
		if scaleDownPolicy != nil {
			scaleDownPolicies = append(scaleDownPolicies, scaleDownPolicy...)
		}
	}
	rsp.Response.ScaleUps = scaleUpPolicies
	rsp.Response.ScaleDowns = scaleDownPolicies
	return nil
}

// CreateNodePoolMgrStrategy create nodePoolMgrStrategy
// createOption cannot be nil
// when overwrite is false, it will return error if strategy exists
func (e *NodegroupManager) CreateNodePoolMgrStrategy(ctx context.Context,
	req *nodegroupmgr.CreateNodePoolMgrStrategyReq, rsp *nodegroupmgr.CreateNodePoolMgrStrategyRsp) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportAPIRequestMetric("grpc", "CreateNodePoolMgrStrategy", err, startTime)
	}()
	if req.Option == nil {
		errMessage := "CreateOptions cannot be nil"
		blog.Errorf(errMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errMessage
		rsp.Result = false
		return nil
	}
	blog.Infof("Received BcsNodegroupManager.CreateNodePoolMgrStrategy request. type:%s, name:%s, operator:%s, "+
		"overwrite:%s", req.Strategy.Kind, req.Strategy.Name, req.Option.Operator, req.Option.OverWriteIfExist)
	storageStrategy := transferToStorageStrategy(req.Strategy)
	err = e.storage.CreateNodeGroupStrategy(storageStrategy,
		&storage.CreateOptions{OverWriteIfExist: req.Option.OverWriteIfExist})
	if err != nil {
		errMessage := fmt.Sprintf("CreateNodePoolMgrStrategy error:%v", err)
		blog.Errorf(errMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errMessage
		rsp.Result = false
		return nil
	}
	blog.Infof("CreateNodePoolMgrStrategy success. StrategyName:%s", req.Strategy.Name)
	rsp.Code = 0
	rsp.Message = "success"
	rsp.Result = true
	return nil
}

// UpdateNodePoolMgrStrategy update nodePoolMgrStrategy
// updateOption cannot be nil
// when CreateIfNotExist is false, it will return error if strategy does not exist
func (e *NodegroupManager) UpdateNodePoolMgrStrategy(ctx context.Context,
	req *nodegroupmgr.UpdateNodePoolMgrStrategyReq, rsp *nodegroupmgr.CreateNodePoolMgrStrategyRsp) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportAPIRequestMetric("grpc", "UpdateNodePoolMgrStrategy", err, startTime)
	}()
	if req.Option == nil {
		errMessage := "UpdateOptions cannot be nil"
		blog.Errorf(errMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errMessage
		rsp.Result = false
		return nil
	}
	blog.Info("Received BcsNodegroupManager.UpdateNodePoolMgrStrategy request.name:%s, options:%s",
		req.Strategy.Name, req.Option.String())
	storageStrategy := transferToStorageStrategy(req.Strategy)
	_, err = e.storage.UpdateNodeGroupStrategy(storageStrategy, &storage.UpdateOptions{
		CreateIfNotExist:        req.Option.CreateIfNotExist,
		OverwriteZeroOrEmptyStr: req.Option.OverwriteZeroOrEmptyStr,
	})
	if err != nil {
		errMessage := fmt.Sprintf("UpdateNodePoolMgrStrategy error:%v", err)
		blog.Errorf(errMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errMessage
		rsp.Result = false
		return nil
	}
	blog.Infof("UpdateNodePoolMgrStrategy success. StrategyName:%s", req.Strategy.Name)
	rsp.Code = 0
	rsp.Message = "success"
	rsp.Result = true
	return nil
}

// GetNodePoolMgrStrategy get nodePoolMgrStrategy
// strategy name cannot be empty
func (e *NodegroupManager) GetNodePoolMgrStrategy(ctx context.Context,
	req *nodegroupmgr.GetNodePoolMgrStrategyReq, rsp *nodegroupmgr.GetNodePoolMgrStrategyRsp) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportAPIRequestMetric("grpc", "GetNodePoolMgrStrategy", err, startTime)
	}()
	blog.Info("Received BcsNodegroupManager.GetNodePoolMgrStrategy request. name:%s", req.Name)
	strategy, err := e.storage.GetNodeGroupStrategy(req.Name, &storage.GetOptions{})
	if err != nil {
		errorMessage := fmt.Sprintf("GetNodePoolMgrStrategy error, err:%v", err)
		blog.Errorf(errorMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errorMessage
		rsp.Data = nil
		return nil
	}
	if strategy != nil {
		rsp.Data = transferToHandlerStrategy(strategy)
	}
	message := "GetNodePoolMgrStrategy success"
	blog.Infof(message)
	rsp.Code = 0
	rsp.Message = message
	return nil
}

// ListNodePoolMgrStrategies get nodePoolMgrStrategy list
// page: default 0
// size: default 10
func (e *NodegroupManager) ListNodePoolMgrStrategies(ctx context.Context,
	req *nodegroupmgr.ListNodePoolMgrStrategyReq, rsp *nodegroupmgr.ListNodePoolMgrStrategyRsp) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportAPIRequestMetric("grpc", "ListNodePoolMgrStrategies", err, startTime)
	}()
	page := int(req.Page)
	size := int(req.Limit)
	blog.Info("Received BcsNodegroupManager.GetNodePoolMgrStrategyList request. Page:%d, size:%d",
		page, size)
	strategyList, err := e.storage.ListNodeGroupStrategies(&storage.ListOptions{
		Limit:                  size,
		Page:                   page,
		ReturnSoftDeletedItems: false,
	})
	if err != nil {
		errorMessage := fmt.Sprintf("ListNodePoolMgrStrategies error, err:%v", err)
		blog.Errorf(errorMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errorMessage
		rsp.Data = nil
		return nil
	}
	result := make([]*nodegroupmgr.NodeGroupStrategy, 0)
	for _, strategy := range strategyList {
		ngStrategy := transferToHandlerStrategy(strategy)
		result = append(result, ngStrategy)
	}
	message := "ListNodePoolMgrStrategies success"
	blog.Infof(message)
	rsp.Code = 0
	rsp.Data = result
	rsp.Message = message
	return nil
}

// DeleteNodePoolMgrStrategy delete nodePoolMgrStrategy
// strategy name cannot be empty
func (e *NodegroupManager) DeleteNodePoolMgrStrategy(ctx context.Context,
	req *nodegroupmgr.DeleteNodePoolMgrStrategyReq, rsp *nodegroupmgr.DeleteNodePoolMgrStrategyRsp) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportAPIRequestMetric("grpc", "DeleteNodePoolMgrStrategy", err, startTime)
	}()
	blog.Infof("Received BcsNodegroupManager.DeleteNodePoolMgrStrategy request. Name: %s, operator:%s",
		req.Name, req.Operator)
	strategy, err := e.storage.DeleteNodeGroupStrategy(req.Name, &storage.DeleteOptions{})
	if err != nil {
		errorMessage := fmt.Sprintf("DeleteNodePoolMgrStrategy error, name:%s, err:%v", req.Name, err)
		blog.Errorf(errorMessage)
		rsp.Code = common.AdditionErrorCode + 500
		rsp.Message = errorMessage
		rsp.Result = false
		return nil
	}
	message := "delete nodeGroupStrategy success"
	if strategy == nil {
		message = "nodeGroupStrategy does not exist"
	}
	// 联动删除action
	if strategy != nil {
		for _, ng := range strategy.ElasticNodeGroups {
			scaleDownAction, getErr := e.storage.GetNodeGroupAction(ng.NodeGroupID,
				storage.ScaleDownState, &storage.GetOptions{})
			if getErr != nil {
				blog.Errorf("get %s scale down action err:%s", ng.NodeGroupID, getErr.Error())
			}
			_, deleteErr := e.storage.DeleteNodeGroupAction(scaleDownAction, &storage.DeleteOptions{})
			if deleteErr != nil {
				blog.Errorf("delete %s scale down action err:%s", ng.NodeGroupID, deleteErr.Error())
			}
			scaleUpAction, getErr := e.storage.GetNodeGroupAction(ng.NodeGroupID, storage.ScaleUpState, &storage.GetOptions{})
			if getErr != nil {
				blog.Errorf("get %s scale up action err:%s", ng.NodeGroupID, getErr.Error())
			}
			_, deleteErr = e.storage.DeleteNodeGroupAction(scaleUpAction, &storage.DeleteOptions{})
			if deleteErr != nil {
				blog.Errorf("delete %s scale up action err:%s", ng.NodeGroupID, deleteErr.Error())
			}
		}
	}
	blog.Infof(message)
	rsp.Code = 0
	rsp.Result = true
	rsp.Message = message
	return nil
}

// transferToHandlerStrategy transfer storage strategy struct to proto struct
func transferToHandlerStrategy(original *storage.NodeGroupMgrStrategy) *nodegroupmgr.NodeGroupStrategy {
	reservedNodeGroup := &nodegroupmgr.ReservedNodeGroup{}
	if original.ReservedNodeGroup != nil {
		reservedNodeGroup.ClusterId = original.ReservedNodeGroup.ClusterID
		reservedNodeGroup.NodeGroup = original.ReservedNodeGroup.NodeGroupID
		reservedNodeGroup.ConsumerId = original.ReservedNodeGroup.ConsumerID
	}
	elasticNodeGroups := make([]*nodegroupmgr.ElasticNodeGroup, 0)
	if original.ElasticNodeGroups != nil {
		for _, group := range original.ElasticNodeGroups {
			elasticGroup := &nodegroupmgr.ElasticNodeGroup{
				ClusterId:  group.ClusterID,
				NodeGroup:  group.NodeGroupID,
				ConsumerId: group.ConsumerID,
				Weight:     int32(group.Weight),
			}
			elasticGroup.Limit = &nodegroupmgr.NodegroupLimit{}
			if err := mapstructure.Decode(group.Limit, elasticGroup.Limit); err != nil {
				blog.Errorf("Error during decoding limit to storage limit:", err.Error())
			}
			elasticNodeGroups = append(elasticNodeGroups, elasticGroup)
		}
	}
	strategy := &nodegroupmgr.Strategy{}
	if original.Strategy != nil {
		strategy.Type = original.Strategy.Type
		strategy.ReservedTimeRange = original.Strategy.ReservedTimeRange
		strategy.MaxIdleDelay = int32(original.Strategy.MaxIdleDelay)
		strategy.MinScaleUpSize = int32(original.Strategy.MinScaleUpSize)
		strategy.ScaleUpDelay = int32(original.Strategy.ScaleUpDelay)
		strategy.ScaleDownDelay = int32(original.Strategy.ScaleDownDelay)
		strategy.ScaleUpCoolDown = int32(original.Strategy.ScaleUpCoolDown)
		strategy.ScaleDownBeforeDDL = int32(original.Strategy.ScaleDownBeforeDDL)
		if original.Strategy.Buffer != nil {
			strategy.Buffer = &nodegroupmgr.Buffer{}
			strategy.Buffer.Low = int32(original.Strategy.Buffer.Low)
			strategy.Buffer.High = int32(original.Strategy.Buffer.High)
		}
		if original.Strategy.TimeMode != nil {
			strategy.TimeMode = &nodegroupmgr.TimeMode{}
			if err := mapstructure.Decode(original.Strategy.TimeMode, strategy.TimeMode); err != nil {
				blog.Errorf("Error during decoding timeMode to pbTimeMode:", err.Error())
			}
		}
		if original.Strategy.NodegroupBuffer != nil {
			strategy.NodegroupBuffer = make(map[string]*nodegroupmgr.BufferParam)
			if err := mapstructure.Decode(original.Strategy.NodegroupBuffer, &strategy.NodegroupBuffer); err != nil {
				blog.Errorf("Error during decoding nodegroupBuffer to pbNodegroupBuffer:%s", err.Error())
			}
		}
	}
	return &nodegroupmgr.NodeGroupStrategy{
		Kind:              "NodeGroupStrategy",
		Name:              original.Name,
		Labels:            original.Labels,
		ResourcePool:      original.ResourcePool,
		ReservedNodeGroup: reservedNodeGroup,
		ElasticNodeGroups: elasticNodeGroups,
		Strategy:          strategy,
	}
}

// transferToStorageStrategy transfer proto strategy struct to storage local struct
func transferToStorageStrategy(original *nodegroupmgr.NodeGroupStrategy) *storage.NodeGroupMgrStrategy {
	reservedNodeGroup := &storage.GroupInfo{}
	if original.ReservedNodeGroup != nil {
		reservedNodeGroup.ClusterID = original.ReservedNodeGroup.ClusterId
		reservedNodeGroup.NodeGroupID = original.ReservedNodeGroup.NodeGroup
		reservedNodeGroup.ConsumerID = original.ReservedNodeGroup.ConsumerId
	}
	elasticNodeGroups := make([]*storage.GroupInfo, 0)
	if original.ElasticNodeGroups != nil {
		for _, group := range original.ElasticNodeGroups {
			elasticGroup := &storage.GroupInfo{
				ClusterID:   group.ClusterId,
				ConsumerID:  group.ConsumerId,
				NodeGroupID: group.NodeGroup,
				Weight:      int(group.Weight),
			}
			elasticGroup.Limit = &storage.NodegroupLimit{}
			if err := mapstructure.Decode(group.Limit, elasticGroup.Limit); err != nil {
				blog.Errorf("Error during decoding limit to storage limit:", err.Error())
			}
			elasticNodeGroups = append(elasticNodeGroups, elasticGroup)
		}
	}
	strategy := &storage.Strategy{}
	if original.Strategy != nil {
		strategy.Type = original.Strategy.Type
		strategy.ReservedTimeRange = original.Strategy.ReservedTimeRange
		strategy.MaxIdleDelay = int(original.Strategy.MaxIdleDelay)
		strategy.MinScaleUpSize = int(original.Strategy.MinScaleUpSize)
		strategy.ScaleUpDelay = int(original.Strategy.ScaleUpDelay)
		strategy.ScaleDownDelay = int(original.Strategy.ScaleDownDelay)
		strategy.ScaleUpCoolDown = int(original.Strategy.ScaleUpCoolDown)
		strategy.ScaleDownBeforeDDL = int(original.Strategy.ScaleDownBeforeDDL)
		if original.Strategy.Buffer != nil {
			strategy.Buffer = &storage.BufferStrategy{}
			strategy.Buffer.Low = int(original.Strategy.Buffer.Low)
			strategy.Buffer.High = int(original.Strategy.Buffer.High)
		}
		if original.Strategy.TimeMode != nil {
			strategy.TimeMode = &storage.BufferTimeMode{}
			if err := mapstructure.Decode(original.Strategy.TimeMode, strategy.TimeMode); err != nil {
				blog.Errorf("Error during decoding timeMode to storage timeMode:", err.Error())
			}
		}
		if original.Strategy.NodegroupBuffer != nil {
			strategy.NodegroupBuffer = make(map[string]*storage.NodegroupBuffer)
			if err := mapstructure.Decode(original.Strategy.NodegroupBuffer, &strategy.NodegroupBuffer); err != nil {
				blog.Errorf("Error during decoding nodegroupBuffer to storage nodegroupBuffer:%s", err.Error())
			}
		}
	}
	status := &storage.State{
		Status:      storage.InitState,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	return &storage.NodeGroupMgrStrategy{
		Name:              original.Name,
		Labels:            original.Labels,
		ResourcePool:      original.ResourcePool,
		ReservedNodeGroup: reservedNodeGroup,
		ElasticNodeGroups: elasticNodeGroups,
		Strategy:          strategy,
		Status:            status,
	}
}

// transferToStorageNodegroup transfer proto nodegroup struct to storage local struct
func transferToStorageNodegroup(origin *nodegroupmgr.NodeGroup,
	storageNodegroup *storage.NodeGroup) (*storage.NodeGroup, bool) {
	retNodegroup := &storage.NodeGroup{}
	if storageNodegroup != nil {
		retNodegroup = storageNodegroup
		if checkNodeGroupEqual(origin, storageNodegroup) {
			return nil, true
		}
	}
	retNodegroup.NodeGroupID = origin.NodeGroupID
	retNodegroup.MaxSize = int(origin.MaxSize)
	retNodegroup.MinSize = int(origin.MinSize)
	retNodegroup.CmDesiredSize = int(origin.DesiredSize)
	retNodegroup.UpcomingSize = int(origin.UpcomingSize)
	retNodegroup.NodeIPs = origin.NodeIPs
	retNodegroup.UpdatedTime = time.Now()
	return retNodegroup, false
}

// checkNodeGroupEqual check if pb.nodegroup is equal to storage.nodegroup
func checkNodeGroupEqual(origin *nodegroupmgr.NodeGroup, storageNodegroup *storage.NodeGroup) bool {
	if origin.DesiredSize != int32(storageNodegroup.CmDesiredSize) {
		return false
	}
	if origin.MaxSize != int32(storageNodegroup.MaxSize) {
		return false
	}
	if origin.MinSize != int32(storageNodegroup.MinSize) {
		return false
	}
	if origin.UpcomingSize != int32(storageNodegroup.UpcomingSize) {
		return false
	}
	if len(origin.NodeIPs) != len(storageNodegroup.NodeIPs) {
		return false
	}
	sort.Strings(origin.NodeIPs)
	sort.Strings(storageNodegroup.NodeIPs)
	for index := range origin.NodeIPs {
		if origin.NodeIPs[index] != storageNodegroup.NodeIPs[index] {
			return false
		}
	}
	return true
}

// handleNodeGroup get scale down/up action of particular nodegroup
func (e *NodegroupManager) handleNodeGroup(nodegroup *nodegroupmgr.NodeGroup,
	uid string) ([]*nodegroupmgr.NodeScaleUpPolicy, []*nodegroupmgr.NodeScaleDownPolicy, error) {
	nodegroupId := nodegroup.NodeGroupID
	storageNodeGroup, err := e.storage.GetNodeGroup(nodegroupId, &storage.GetOptions{})
	if err != nil {
		errMessage := fmt.Sprintf("[%s]get nodegroup[%s] from storage err:%v", uid, nodegroupId, err)
		blog.Error(errMessage)
		return nil, nil, fmt.Errorf(errMessage)
	}
	storageUpdateNodegroup := storageNodeGroup
	updateNodeGroup, equal := transferToStorageNodegroup(nodegroup, storageNodeGroup)
	if !equal {
		blog.Infof("[%s] nodegroup[%s] change.", nodegroup.NodeGroupID, uid)
		storageUpdateNodegroup, err = e.storage.UpdateNodeGroup(updateNodeGroup, &storage.UpdateOptions{
			CreateIfNotExist:        true,
			OverwriteZeroOrEmptyStr: true,
		})
		if err != nil {
			errMessage := fmt.Sprintf("[%s]update nodegroup[%s] err:%v", uid, nodegroupId, err)
			blog.Error(errMessage)
			return nil, nil, fmt.Errorf(errMessage)
		}
	}
	scaleUpPolicies := make([]*nodegroupmgr.NodeScaleUpPolicy, 0)
	scaleDownPolicies := make([]*nodegroupmgr.NodeScaleDownPolicy, 0)
	actions, err := e.storage.ListNodeGroupAction(nodegroupId, &storage.ListOptions{})
	if err != nil {
		errMessage := fmt.Sprintf("[%s]list nodegroup[%s] actions err:%v", uid, nodegroupId, err)
		blog.Error(errMessage)
		return nil, nil, fmt.Errorf(errMessage)
	}
	for _, action := range actions {
		switch action.Event {
		case storage.ScaleUpState:
			scaleUp := &nodegroupmgr.NodeScaleUpPolicy{
				NodeGroupID: action.NodeGroupID,
				DesiredSize: int32(storageUpdateNodegroup.DesiredSize),
			}
			scaleUpPolicies = append(scaleUpPolicies, scaleUp)
		case storage.ScaleDownState:
			scaleDown := &nodegroupmgr.NodeScaleDownPolicy{
				NodeGroupID: action.NodeGroupID,
				Type:        "NodeNum",
				NodeNum:     int32(storageUpdateNodegroup.DesiredSize),
			}
			scaleDownPolicies = append(scaleDownPolicies, scaleDown)
		case storage.ScaleDownByTaskState:
			scaleDown := &nodegroupmgr.NodeScaleDownPolicy{
				NodeGroupID: action.NodeGroupID,
				Type:        "NodeIPs",
				NodeIPs:     action.NodeIPs,
			}
			scaleDownPolicies = append(scaleDownPolicies, scaleDown)
		}
	}
	return scaleUpPolicies, scaleDownPolicies, nil
}

// updateNodeGroupStatus
// update the process of  scale down/up action
// update nodegroup hookConfirm
func (e *NodegroupManager) updateNodeGroupStatus(response *nodegroupmgr.ClusterAutoscalerReview) {
	for _, scaleUp := range response.Response.ScaleUps {
		desireNum := int(scaleUp.DesiredSize)
		currentNum := len(response.Request.NodeGroups[scaleUp.NodeGroupID].NodeIPs)
		process := calculateProcess(currentNum, desireNum)
		blog.Infof("calculate process: currentNum:%d, desireNum:%d, process:%d", currentNum, desireNum, process)
		action := &storage.NodeGroupAction{
			NodeGroupID: scaleUp.NodeGroupID,
			Event:       storage.ScaleUpState,
			Process:     process,
			UpdatedTime: time.Now(),
		}
		_, err := e.storage.UpdateNodeGroupAction(action, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("update nodegroup[%s] scale up action error:%v", scaleUp.NodeGroupID, err)
		}
		_, err = e.storage.UpdateNodeGroup(&storage.NodeGroup{
			NodeGroupID: scaleUp.NodeGroupID,
			HookConfirm: true,
			UpdatedTime: time.Now(),
		}, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("update nodegroup[%s] error:%v", scaleUp.NodeGroupID, err)
		}
	}
	for _, scaleDown := range response.Response.ScaleDowns {
		if scaleDown.Type == "NodeNum" {
			desireNum := int(scaleDown.NodeNum)
			currentNum := len(response.Request.NodeGroups[scaleDown.NodeGroupID].NodeIPs)
			process := calculateProcess(currentNum, desireNum)
			blog.Infof("calculate process: currentNum:%d, desireNum:%d, process:%d", currentNum, desireNum, process)
			action := &storage.NodeGroupAction{
				NodeGroupID: scaleDown.NodeGroupID,
				Event:       storage.ScaleDownState,
				Process:     process,
				UpdatedTime: time.Now(),
			}
			_, err := e.storage.UpdateNodeGroupAction(action, &storage.UpdateOptions{})
			if err != nil {
				blog.Errorf("update nodegroup[%s] scale down action error:%v", scaleDown.NodeGroupID, err)
			}
		}
		_, err := e.storage.UpdateNodeGroup(&storage.NodeGroup{
			NodeGroupID: scaleDown.NodeGroupID,
			HookConfirm: true,
			UpdatedTime: time.Now(),
		}, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("update nodegroup[%s] error:%v", scaleDown.NodeGroupID, err)
		}
	}
}

// calculate scale down/up process
func calculateProcess(current, desired int) int {
	if desired == 0 {
		if current == 0 {
			return 100
		}
		return 0
	}
	if current < desired {
		return current * 100 / desired
	}
	return 100 - (current-desired)*100/desired
}
