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

// Package strategy xxx
package strategy

import (
	"fmt"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/robfig/cron/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// BufferStrategyExecutor buffer type strategy executor
type BufferStrategyExecutor struct {
	opt *Options
}

// NewBufferStrategyExecutor return BufferStrategyExecutor
func NewBufferStrategyExecutor(opt *Options) *BufferStrategyExecutor {
	return &BufferStrategyExecutor{
		opt: opt,
	}
}

// IsAbleToScaleDown check if is able to scale down, return the scale down num
func (e *BufferStrategyExecutor) IsAbleToScaleDown(strategy *storage.NodeGroupMgrStrategy) (int,
	bool, error) {
	if strategy.Strategy.TimeMode != nil && strategy.Strategy.TimeMode.ScaleDownWhenTimeout {
		needScaleIn := needToScaleIn(strategy.Strategy.TimeMode, strategy.Strategy.ScaleDownBeforeDDL)
		if needScaleIn {
			scaleDownNum := 0
			for _, ng := range strategy.ElasticNodeGroups {
				info, getErr := e.opt.Storage.GetNodeGroup(ng.NodeGroupID, &storage.GetOptions{ErrIfNotExist: true})
				if getErr != nil {
					blog.Errorf("get %s nodegroup info from db error:%s", ng.NodeGroupID, getErr.Error())
					continue
				}
				scaleDownNum += info.CmDesiredSize
			}
			if scaleDownNum == 0 {
				return scaleDownNum, false, nil
			}
			return scaleDownNum, true, nil
		}
	}
	blog.Infof("controller handle strategy %s for ResourcePool %s", strategy.Name, strategy.ResourcePool)
	var consumerID string
	if strategy.ReservedNodeGroup == nil || strategy.ReservedNodeGroup.ConsumerID == "" {
		for _, elasticGroup := range strategy.ElasticNodeGroups {
			if elasticGroup.ConsumerID != "" {
				consumerID = elasticGroup.ConsumerID
				break
			}
		}
	} else {
		consumerID = strategy.ReservedNodeGroup.ConsumerID
	}
	if consumerID == "" {
		return 0, false, fmt.Errorf("strategy %s consumer id is empty", strategy.Name)
	}
	// query relative ResourcePool information from resource-manager
	pool, err := e.opt.ResourceManager.GetDeviceListByConsumer(consumerID, nil)
	if err != nil {
		blog.Errorf("controller got ResourcePool %s from resource-manager failed, %s",
			strategy.ResourcePool, err.Error())
		return 0, false, fmt.Errorf("get dependent resourcepool %s failed", strategy.ResourcePool)
	}
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	if strategy.Strategy.Buffer == nil {
		strategy.Strategy.Buffer = &storage.BufferStrategy{
			Low:  0,
			High: 0,
		}
	}
	warnBuffer := float64(strategy.Strategy.Buffer.High)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, consumerID:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ConsumerID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum >= reservedNum {
		// resource is enough, do nothing
		blog.Infof("device group of consumer %s resource is idle %d >= reserved %d, elasticNodeGroup don't scaleDown",
			pool.ConsumerID, idleNum, reservedNum)
		return 0, false, nil
	}
	// buffer resource is not enough, calculate necessary number for scale down
	scaleDownNum := reservedNum - idleNum
	blog.Infof("strategy:%s, scaleDownNum:%d", strategy.Name, scaleDownNum)
	return scaleDownNum, true, nil
}

// IsAbleToScaleUp check if is able to scale up
func (e *BufferStrategyExecutor) IsAbleToScaleUp(strategy *storage.NodeGroupMgrStrategy) (int,
	bool, int, error) {
	blog.Infof("controller handle strategy %s for ResourcePool %s", strategy.Name, strategy.ResourcePool)
	isExecutingTask, err := checkIfTaskExecuting(strategy.Name, e.opt.Storage)
	if err != nil {
		return 0, false, 0, fmt.Errorf("check if task executing failed: %s", err.Error())
	}
	if isExecutingTask {
		blog.Infof("strategy %s is executing scale down task, skip scale up", strategy.Name)
		return 0, false, 0, nil
	}
	// query relative ResourcePool information from resource-manager
	if strategy.Strategy.TimeMode != nil {
		if !checkIfInScaleOutPeriod(strategy.Strategy.TimeMode, strategy.Strategy.ScaleDownBeforeDDL) {
			blog.Infof("strategy %s is not in scale out period, skip", strategy.Name)
			return 0, false, 0, nil
		}
	}
	var consumerID string
	if strategy.ReservedNodeGroup == nil || strategy.ReservedNodeGroup.ConsumerID == "" {
		for _, elasticGroup := range strategy.ElasticNodeGroups {
			if elasticGroup.ConsumerID != "" {
				consumerID = elasticGroup.ConsumerID
				break
			}
		}
	} else {
		consumerID = strategy.ReservedNodeGroup.ConsumerID
	}
	if consumerID == "" {
		return 0, false, 0, fmt.Errorf("strategy %s consumer id is empty", strategy.Name)
	}

	pool, getErr := e.opt.ResourceManager.GetDeviceListByConsumer(consumerID, nil)
	if getErr != nil {
		blog.Errorf("controller got ResourcePool %s from resource-manager failed, %s",
			strategy.ResourcePool, getErr.Error())
		return 0, false, 0, fmt.Errorf("get dependent resourcepool %s failed", strategy.ResourcePool)
	}
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	if strategy.Strategy.Buffer == nil {
		strategy.Strategy.Buffer = &storage.BufferStrategy{
			Low:  0,
			High: 0,
		}
	}
	warnBuffer := float64(strategy.Strategy.Buffer.High)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	allowScaleUpNum := int(math.Ceil(total * float64(strategy.Strategy.Buffer.Low) / 100))
	blog.Infof("strategy %s, consumerID:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ConsumerID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum <= reservedNum {
		// resource is not idle enough
		blog.Infof("device group of consumer %s idle resource %d <= reserved %d, elasticNodeGroup don't scaleUp",
			pool.ConsumerID, idleNum, reservedNum)
		return 0, false, 0, nil
	}
	if idleNum <= allowScaleUpNum {
		blog.Infof("consumer %s idle resource %d <= allowScaleUpNum %d, elasticNodeGroup don't scaleUp",
			pool.ConsumerID, idleNum, allowScaleUpNum)
		return 0, false, 0, nil
	}
	// check resource pool is idle and stable
	now := time.Now()
	diff := now.Sub(pool.UpdatedTime)
	if diff.Seconds() < float64(strategy.Strategy.MaxIdleDelay*60) {
		blog.Infof("device group of consumer %s is not stable enough for elasticNodeGroup scaleUp, now: %.f, target: %d",
			pool.ConsumerID, diff.Seconds(), strategy.Strategy.MaxIdleDelay*60)
		return 0, false, 0, nil
	}
	// resource is more than expected, check if controller can scale up
	scaleUpNum := idleNum - reservedNum
	if allowScaleUpNum > reservedNum {
		scaleUpNum = idleNum - allowScaleUpNum
	}
	if scaleUpNum < strategy.Strategy.MinScaleUpSize {
		blog.Infof("device group of consumer %s idle resource %d is less than MinScaleUpSize %d",
			pool.ConsumerID, scaleUpNum, strategy.Strategy.MinScaleUpSize)
		return 0, false, 0, nil
	}
	blog.Infof("strategy %s scaleUpNum:%d", strategy.Name, scaleUpNum)
	// feature(DeveloperJim): try to check ScaleUpCoolDown
	// if diff.Seconds() < strategy.Strategy.ScaleUpCoolDown do nothing
	return scaleUpNum, true, int(total), nil
}

// HandleNodeMetadata handle node metadata
func (e *BufferStrategyExecutor) HandleNodeMetadata() {
	updateNodeAction, err := e.opt.Storage.ListNodeGroupActionByEvent(storage.UpdateNodeMeta, &storage.ListOptions{})
	if err != nil {
		blog.Errorf("list updateNodeAction error: %s", err.Error())
		return
	}
	for _, action := range updateNodeAction {
		strategy, getErr := e.opt.Storage.GetNodeGroupStrategy(action.Strategy, &storage.GetOptions{
			ErrIfNotExist: true,
		})
		if getErr != nil {
			blog.Errorf("[BufferStrategyExecutor] get strategy %s error:%s", action.Strategy, getErr.Error())
			continue
		}
		nodeGroupLabel := map[string]interface{}{
			storage.NodeGroupLabel: action.NodeGroupID,
		}
		nodeList, listErr := e.opt.ClusterClient.ListNodesByLabel(action.ClusterID, nodeGroupLabel)
		if listErr != nil {
			blog.Errorf("[BufferStrategyExecutor] get cluster %s nodeList by nodegroup %s label failed: %s",
				action.ClusterID, action.NodeGroupID, listErr.Error())
			return
		}
		var drainDelay int
		var deadline string
		if strategy.Strategy.TimeMode != nil {
			deadline, drainDelay = getTimeModeDeadline(strategy.Strategy.TimeMode)
		}
		labels := map[string]interface{}{
			storage.NodeDrainDelayLabel: fmt.Sprintf("%dh", drainDelay),
		}
		annotations := map[string]interface{}{
			storage.NodeDeadlineLabel: deadline,
		}
		blog.Infof("deadline:%s, drain-delay:%d", deadline, drainDelay)
		for node := range nodeList {
			if nodeList[node].Annotations[storage.NodeDeadlineLabel] == deadline {
				continue
			}
			err = e.opt.ClusterClient.UpdateNodeMetadata(action.ClusterID, node, labels, annotations)
			if err != nil {
				blog.Errorf("[BufferStrategyExecutor] update node label error. node:%s, err:%s", node, err.Error())
				return
			}
			blog.Infof("[BufferStrategyExecutor] update node %s success", node)
		}
		nodegroup, getErr := e.opt.Storage.GetNodeGroup(action.NodeGroupID, &storage.GetOptions{ErrIfNotExist: true})
		if err != nil {
			blog.Errorf("[BufferStrategyExecutor] get nodegroup %s error:%s", action.NodeGroupID, getErr.Error())
			continue
		}
		if nodegroup.CmDesiredSize <= len(nodeList) {
			blog.Infof("[BufferStrategyExecutor] updateNodeMetaAction finished, nodegroup:%s", action.NodeGroupID)
			if _, deleteErr := e.opt.Storage.DeleteNodeGroupAction(action, &storage.DeleteOptions{}); deleteErr != nil {
				blog.Errorf("[BufferStrategyExecutor] delete updateNodeMetaAction error:%s", deleteErr.Error())
			}
		} else {
			blog.Infof("[BufferStrategyExecutor] updateNodeMetaAction does not finished, nodegroup:%s, desire:%d, ready:%d",
				action.NodeGroupID, nodegroup.CmDesiredSize, len(nodeList))
		}
	}
}

// CreateNodeUpdateAction create update action
func (e *BufferStrategyExecutor) CreateNodeUpdateAction(strategy *storage.NodeGroupMgrStrategy,
	action *storage.NodeGroupAction) error {
	if action.Event != storage.ScaleUpState || strategy.Strategy.TimeMode == nil {
		blog.Infof("[BufferStrategyExecutor] do not need to update node, action event:%s", action.Event)
		return nil
	}
	updateAction := &storage.NodeGroupAction{
		NodeGroupID: action.NodeGroupID,
		ClusterID:   action.ClusterID,
		CreatedTime: time.Now(),
		Event:       storage.UpdateNodeMeta,
		UpdatedTime: time.Now(),
		Strategy:    strategy.Name,
	}
	err := e.opt.Storage.CreateNodeGroupAction(updateAction, &storage.CreateOptions{OverWriteIfExist: true})
	if err != nil {
		err = fmt.Errorf("create UpdateNodeMeta action error:%s, strategy:%s, nodegroup:%s",
			err.Error(), strategy.Name, action.NodeGroupID)
		blog.Errorf(err.Error())
		return err
	}
	return nil
}

func getTimeModeDeadline(timeMode *storage.BufferTimeMode) (string, int) {
	if timeMode.ReservedHours != 0 && (timeMode.TimePeriods == nil || len(timeMode.TimePeriods) == 0) {
		return time.Now().Add(time.Duration(timeMode.ReservedHours) * time.Hour).Format(time.RFC3339),
			timeMode.ReservedHours
	}
	for _, period := range timeMode.TimePeriods {
		if period.ScaleOutCron != "" {
			nextScaleOut, nextScaleIn := getTimeRangeByCron(period.ScaleOutCron, period.ScaleInCron)
			if nextScaleOut.IsZero() || nextScaleIn.IsZero() {
				blog.Errorf("scale in/out cron error")
				return time.Now().Add(12 * time.Hour).Format(time.RFC3339), 12
			}
			if time.Now().After(nextScaleOut) && time.Now().Before(nextScaleIn) {
				blog.Infof("deadline is %s", nextScaleIn.Format(time.RFC3339))
				return nextScaleIn.Format(time.RFC3339), int(math.Floor(time.Until(nextScaleIn).Hours()))
			}
			continue
		}
		scaleOutTime, err := time.Parse(time.RFC3339, period.ScaleOutTime)
		if err != nil {
			blog.Errorf("scaleOutTime %s parse error:%s", period.ScaleOutTime, err.Error())
			return time.Now().Add(12 * time.Hour).Format(time.RFC3339), 12
		}
		scaleInTime, err := time.Parse(time.RFC3339, period.ScaleInTime)
		if err != nil {
			blog.Errorf("scaleInTime %s parse error:%s", period.ScaleInTime, err.Error())
			return time.Now().Add(12 * time.Hour).Format(time.RFC3339), 12
		}
		if time.Now().After(scaleOutTime) && time.Now().Before(scaleInTime) {
			return scaleInTime.Format(time.RFC3339), int(math.Floor(time.Until(scaleInTime).Hours()))
		}
	}
	// for safety, return deadline after 12 hours
	return time.Now().Add(12 * time.Hour).Format(time.RFC3339), 12
}

func getTimeRangeByCron(scaleOut, scaleIn string) (time.Time, time.Time) {
	scaleOutCron, err := cron.ParseStandard(scaleOut)
	if err != nil {
		blog.Errorf("get scale out scheduler error:%s, scaleOut:%s", err.Error(), scaleOut)
		return time.Time{}, time.Time{}
	}
	scaleInCron, err := cron.ParseStandard(scaleIn)
	if err != nil {
		blog.Errorf("get scale in scheduler error:%s, scaleOut:%s", err.Error(), scaleIn)
		return time.Time{}, time.Time{}
	}
	nextScaleIn := scaleInCron.Next(time.Now())
	nextScaleOut := scaleOutCron.Next(time.Now())
	if nextScaleOut.After(nextScaleIn) {
		return time.Now(), nextScaleIn
	}
	return nextScaleOut, nextScaleIn
}

func checkIfInScaleOutPeriod(timeMode *storage.BufferTimeMode, scaleDownBeforeDDL int) bool {
	if timeMode.TimePeriods == nil || len(timeMode.TimePeriods) == 0 {
		return true
	}
	for _, period := range timeMode.TimePeriods {
		if period.ScaleOutCron != "" {
			nextScaleOut, nextScaleIn := getTimeRangeByCron(period.ScaleOutCron, period.ScaleInCron)
			if nextScaleOut.IsZero() || nextScaleIn.IsZero() {
				blog.Errorf("scale in/out cron error")
				return false
			}
			if time.Now().After(nextScaleOut) && time.Now().Add(time.Duration(scaleDownBeforeDDL)*time.Minute).
				Before(nextScaleIn) {
				return true
			}
		} else {
			scaleOutTime, err := time.Parse(time.RFC3339, period.ScaleOutTime)
			if err != nil {
				blog.Errorf("scaleOutTime %s parse error:%s", period.ScaleOutTime, err.Error())
				return false
			}
			scaleInTime, err := time.Parse(time.RFC3339, period.ScaleInTime)
			if err != nil {
				blog.Errorf("scaleInTime %s parse error:%s", period.ScaleInTime, err.Error())
				return false
			}
			if time.Now().After(scaleOutTime) && time.Now().Before(scaleInTime) {
				return true
			}
		}
	}
	return false
}

func needToScaleIn(timeMode *storage.BufferTimeMode, scaleDownBeforeDDL int) bool {
	inScaleOutPeriod := checkIfInScaleOutPeriod(timeMode, scaleDownBeforeDDL)
	if timeMode.TimePeriods != nil && len(timeMode.TimePeriods) != 0 && inScaleOutPeriod {
		return false
	}
	// if timePeriod is empty, it will return true, need to check by scale up action with reserve hours
	if inScaleOutPeriod {
		// TODO: add logic
		return false
	}
	return true
}
