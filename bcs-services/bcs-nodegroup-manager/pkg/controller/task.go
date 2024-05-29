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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/panjf2000/ants/v2"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// control inner implementation controller for NodeGroupMgrStrategy
type taskController struct {
	opt *Options
}

// NewTaskController return task controller
func NewTaskController(option *Options) Controller {
	return &taskController{
		opt: option,
	}
}

// Init task controller
func (c *taskController) Init(opts ...Option) error {
	// init all custom Option
	for _, opt := range opts {
		opt(c.opt)
	}
	// init all dependent resource, such as storage, client and etc.
	if c.opt.ResourceManager == nil {
		blog.Errorf("[taskController] Controller lost resource-manager interface in Init Stage")
		return fmt.Errorf("controller lost resource-manager instance")
	}
	if c.opt.Storage == nil {
		blog.Errorf("[taskController] Controller lost storage interface in Init Stage")
		return fmt.Errorf("controller lost storage instance")
	}
	if c.opt.ClusterClient == nil {
		blog.Errorf("[taskController] Controller lost cluster client interface in Init Stage")
		return fmt.Errorf("controller lost cluster client instance")
	}
	return nil
}

// Options taskController implement controller interface
func (c *taskController) Options() *Options {
	return c.opt
}

// Run taskController implement controller interface
func (c *taskController) Run(cxt context.Context) {
	tick := time.NewTicker(time.Second * time.Duration(c.opt.Interval))
	for {
		select {
		case now := <-tick.C:
			// main loops
			blog.Infof("[taskController] ############## task controller ticker: %s################", now.Format(time.RFC3339))
			c.controllerLoops()
		case <-cxt.Done():
			blog.Infof("[taskController] task Controller is asked to exit")
			return
		}
	}
}

func (c *taskController) controllerLoops() {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("[taskController] panic in taskController, info: %v", r)
		}
	}()
	// handleNewTask
	c.handleNormalTask()
	// handleTerminatedTask
	c.handleTerminatedTask()
	// handleExpiredTask
	c.handleExpiredTask()
	// traceExecutingTask
	c.traceExecutingTask()
}

func (c *taskController) handleNormalTask() {
	blog.Infof("[taskController] begin handleNormalTask")
	// 获取所有strategy
	// 根据pool id和其中一个consumer id获取task，确认是哪一个策略需要执行缩容
	nodegroupStrategyList, err := c.opt.Storage.ListNodeGroupStrategies(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list nodegroup strategy from storage failed:%s", err.Error())
		return
	}
	scaleDownTaskList := make(map[string]*storage.ScaleDownTask)
	c.getCrTasks(nodegroupStrategyList, scaleDownTaskList)
	blog.Infof("len of scaleDownTaskList: %d", len(scaleDownTaskList))
	blog.Infof("scaleDownTaskList: %v", scaleDownTaskList)
	getSelfTaskErr := c.getSelfTasks(nodegroupStrategyList, scaleDownTaskList)
	if getSelfTaskErr != nil {
		blog.Errorf("get cr task err:%s", getSelfTaskErr.Error())
		return
	}
	blog.Infof("len of scaleDownTaskList: %d", len(scaleDownTaskList))
	blog.Infof("scaleDownTaskList: %v", scaleDownTaskList)
	// 确认没有返回的task是否terminated，如果是，更新状态
	tasks, listErr := c.opt.Storage.ListTasks(&storage.ListOptions{})
	if listErr != nil {
		blog.Errorf("[taskController] list task error:%s", listErr.Error())
		return
	}
	terminatedTasks := c.getTerminatedTasks(tasks, scaleDownTaskList)
	for _, task := range terminatedTasks {
		task.Status = storage.TaskFinishedState
		_, err = c.opt.Storage.UpdateTask(task, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("[taskController]update task %s status to terminated failed:%s", task.TaskID, err.Error())
			return
		}
		blog.Infof("[taskController] update task %s status to finished success", task.TaskID)
		metric.ReportTerminatedTaskNumMetric(task.NodeGroupStrategy, task.DrainDelay)
	}
	for taskID := range scaleDownTaskList {
		if time.Now().Before(scaleDownTaskList[taskID].BeginExecuteTime) ||
			(time.Now().Before(scaleDownTaskList[taskID].Deadline) && !scaleDownTaskList[taskID].IsExecuted) {
			c.handleOneNormalTask(scaleDownTaskList[taskID])
		}
	}
	blog.Infof("[taskController] completed handleNormalTask")
}

// NOCC:golint/funlen(设计如此)
// nolint
func (c *taskController) handleOneNormalTask(task *storage.ScaleDownTask) {
	blog.Infof("begin handle task %s", task.TaskID)
	var err error
	// only update info from resource manager
	updateTask := &storage.ScaleDownTask{
		TaskID:            task.TaskID,
		NodeGroupStrategy: task.NodeGroupStrategy,
		DevicePoolID:      task.DevicePoolID,
		TotalNum:          task.TotalNum,
		DrainDelay:        task.DrainDelay,
		Deadline:          task.Deadline,
		BeginExecuteTime:  task.BeginExecuteTime,
		UpdatedTime:       time.Now(),
		Status:            task.Status,
		DeviceList:        task.DeviceList,
		SpecifyScaleDown:  task.SpecifyScaleDown,
		ScaleDownGroups:   task.ScaleDownGroups,
	}
	updateTask, err = c.opt.Storage.UpdateTask(updateTask, &storage.UpdateOptions{CreateIfNotExist: true})
	if err != nil {
		blog.Errorf("[taskController] update task %s error:%s", task.TaskID, err.Error())
		return
	}
	blog.Infof("[taskController] updateTask:%v", updateTask)
	// check chosen nodes if satisfy task, select nodegroup and node ip
	originTotal := 0
	if updateTask.ScaleDownGroups != nil {
		for _, scaleDownDetail := range updateTask.ScaleDownGroups {
			if scaleDownDetail != nil {
				originTotal += len(scaleDownDetail.NodeIPs)
			}
		}
	}
	blog.Infof("[taskController] originalTotal:%d", originTotal)
	// 如果ip数量和任务的数量一致，说明上一轮没有ip被剔除，此时检查节点状态
	if !task.SpecifyScaleDown {
		if originTotal == task.TotalNum {
			// 先检查ScaleDownGroups里的节点状态是否都是ready，如果不是，从切片里删除
			// 检查nodegroup里带标签的节点跟切片里的是否一致，不在切片里但有标签的，去掉
			// 如果执行时间在1小时内，不操作？
			wg := sync.WaitGroup{}
			pool, initErr := ants.NewPool(c.opt.Concurrency)
			if initErr != nil {
				blog.Errorf("[taskController] task controller internal error, ants.NewPool failed: %s", initErr.Error())
				return
			}
			defer pool.Release()
			for _, scaleDownDetail := range updateTask.ScaleDownGroups {
				wg.Add(1)
				detail := scaleDownDetail
				err = pool.Submit(func() {
					c.removeNotReadyNodes(detail, task.TaskID)
					wg.Done()
				})
				if err != nil {
					blog.Errorf("[taskController] submit task to ch pool err:%v", err)
				}
			}
			wg.Wait()
			blog.Infof("[taskController] finish check task %s node status", task.TaskID)
		} else {
			// 数量不一致，要重选ip
			if err = c.nodeSelector(updateTask); err != nil {
				blog.Errorf("[taskController] execute task %s node select failed:%s", task.TaskID, err.Error())
				return
			}
		}
	}
	// label node
	for _, group := range updateTask.ScaleDownGroups {
		labels := map[string]interface{}{storage.NodeDrainTaskLabel: updateTask.TaskID}
		if strings.Contains(updateTask.TaskID, "-specifyDevice-") {
			labels[storage.NodeDrainTaskLabel] = strings.Split(updateTask.TaskID, "-specifyDevice-")[0]
		}
		annotations := map[string]interface{}{storage.NodeDeadlineLabel: updateTask.Deadline.Format(time.RFC3339)}
		nodegroupLabel := map[string]interface{}{storage.NodeGroupLabel: group.NodeGroupID}
		nodeList, err := c.opt.ClusterClient.ListNodesByLabel(group.ClusterID, nodegroupLabel)
		if err != nil {
			blog.Errorf("[taskController] get cluster %s nodeList by nodegroup %s label failed: %s",
				group.ClusterID, group.NodeGroupID, err.Error())
			return
		}
		for _, ip := range group.NodeIPs {
			if nodeList[ip].Annotations[storage.NodeDeadlineLabel] == updateTask.Deadline.Format(time.RFC3339) &&
				nodeList[ip].Labels[storage.NodeDrainTaskLabel] == updateTask.TaskID {
				continue
			}
			blog.Infof("label deadline:%s, deadline:%s, label task:%s, task:%s",
				nodeList[ip].Annotations[storage.NodeDeadlineLabel], updateTask.Deadline.Format(time.RFC3339),
				nodeList[ip].Labels[storage.NodeDrainTaskLabel], labels[storage.NodeDrainTaskLabel])
			blog.Infof("[taskController] update node %s labels", ip)
			if err = c.opt.ClusterClient.UpdateNodeMetadata(group.ClusterID, ip, labels, annotations); err != nil {
				blog.Errorf("[taskController] UpdateNodeLabels error. task id: %s, clusterID:%s, nodeIP:%s, "+
					"labels:%s, error:%s", task.TaskID, group.ClusterID, ip, labels, err.Error())
				break
			}
		}
	}
}

func (c *taskController) handleExpiredTask() {
	blog.Infof("[taskController] begin handleExpiredTask")
	// check all tasks' expired time
	tasks, err := c.opt.Storage.ListTasks(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list task error:%s", err.Error())
		return
	}
	expiredTasks := checkExpiredTask(tasks)
	// create scale down action
	for _, task := range expiredTasks {
		if task.IsExecuted {
			continue
		}
		if err := c.handleOneExpiredTask(task); err != nil {
			blog.Errorf("[taskController] handleOneExpiredTask error:%s", err.Error())
			continue
		}
		task.IsExecuted = true
		if _, err := c.opt.Storage.UpdateTask(task, &storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}); err != nil {
			blog.Errorf("[taskController] update task %s error:%s", task.TaskID, err.Error())
		}
	}
	blog.Infof("[taskController] completed handleExpiredTask")
}

func (c *taskController) handleTerminatedTask() {
	blog.Infof("[taskController] begin handleTerminatedTask")
	tasks, err := c.opt.Storage.ListTasks(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list tasks error:%s", err.Error())
		return
	}
	for _, task := range tasks {
		if task.IsTerminated() {
			for _, scaleDownGroup := range task.ScaleDownGroups {
				err = c.removeLabel(scaleDownGroup.ClusterID, task.TaskID)
				if err != nil {
					blog.Errorf("[taskController] remove task %s label error. cluster id:%s, nodegroup id:%s, error:%s",
						task.TaskID, scaleDownGroup.ClusterID, scaleDownGroup.NodeGroupID, err.Error())
					return
				}
			}
			if _, err := c.opt.Storage.DeleteTask(task.TaskID, &storage.DeleteOptions{}); err != nil {
				blog.Errorf("[taskController] delete task %s error:%s", task.TaskID, err.Error())
				return
			}
			blog.Infof("[taskController] delete terminated task %s success", task.TaskID)
		}
	}
	blog.Infof("[taskController] completed handleTerminatedTask")
}

func checkExpiredTask(tasks []*storage.ScaleDownTask) []*storage.ScaleDownTask {
	expiredTasks := make([]*storage.ScaleDownTask, 0)
	for index := range tasks {
		if tasks[index].BeginExecuteTime.Before(time.Now()) {
			expired := &storage.ScaleDownTask{}
			bytes, err := json.Marshal(tasks[index])
			if err != nil {
				blog.Errorf("marshal task to bytes error:%s", err.Error())
				continue
			}
			err = json.Unmarshal(bytes, expired)
			if err != nil {
				blog.Errorf("unmarshal task error:%s", err.Error())
				continue
			}
			expiredTasks = append(expiredTasks, expired)
		}
	}
	return expiredTasks
}

func (c *taskController) handleOneExpiredTask(task *storage.ScaleDownTask) error {
	var err error
	strategyName := task.NodeGroupStrategy
	strategyDetail, err := c.opt.Storage.GetNodeGroupStrategy(strategyName, &storage.GetOptions{})
	if err != nil {
		return fmt.Errorf("[taskController] get strategy %s detail error:%s", strategyName, err.Error())
	}
	// clear scale up action
	blog.Infof("[taskController] clean up scale up action")
	for _, nodegroup := range strategyDetail.ElasticNodeGroups {
		if _, err = c.opt.Storage.DeleteNodeGroupAction(&storage.NodeGroupAction{
			NodeGroupID: nodegroup.NodeGroupID,
			Event:       storage.ScaleUpState,
		}, &storage.DeleteOptions{}); err != nil {
			return fmt.Errorf("clean up scale up action of nodegroup %s error:%s",
				nodegroup.NodeGroupID, err.Error())
		}
	}

	for _, scaleDownDetail := range task.ScaleDownGroups {
		blog.Infof("[taskController] begin handle %s scale down nodes", scaleDownDetail.NodeGroupID)
		action := &storage.NodeGroupAction{
			NodeGroupID:        scaleDownDetail.NodeGroupID,
			ClusterID:          scaleDownDetail.ClusterID,
			TaskID:             task.TaskID,
			CreatedTime:        time.Now(),
			Event:              storage.ScaleDownByTaskState,
			DeltaNum:           len(scaleDownDetail.NodeIPs),
			NewDesiredNum:      0,
			OriginalDesiredNum: 0,
			OriginalNodeNum:    0,
			NodeIPs:            scaleDownDetail.NodeIPs,
			Process:            0,
			Status:             storage.InitState,
			UpdatedTime:        time.Now(),
			IsDeleted:          false,
		}
		_, err = c.opt.Storage.UpdateNodeGroupAction(action, &storage.UpdateOptions{CreateIfNotExist: true})
		if err != nil {
			blog.Errorf("[taskController] update scaleDown by task action err. taskID:%s, nodegroupID:%s, clusterID:%s, "+
				"err:%s", task.TaskID, scaleDownDetail.NodeGroupID, scaleDownDetail.ClusterID, err.Error())
			return err
		}
		metric.ReportActionNumMetric(task.NodeGroupStrategy, action.ClusterID, action.NodeGroupID, action.Event)
		event := &storage.NodeGroupEvent{
			NodeGroupID: scaleDownDetail.NodeGroupID,
			ClusterID:   scaleDownDetail.ClusterID,
			EventTime:   time.Now(),
			Event:       storage.ScaleDownByTaskState,
			Reason:      fmt.Sprintf("trigger by task %s", task.TaskID),
			Message:     fmt.Sprintf("trigger by task %s", task.TaskID),
			IsDeleted:   false,
		}
		if err = c.opt.Storage.CreateNodeGroupEvent(event, &storage.CreateOptions{}); err != nil {
			// event only used for administrator tracing issue manually.
			// failure of event operation is tolerable.
			blog.Errorf("[taskController] controller create nodegroup %s scaleDown record failure, info: %s."+
				"failure is tolerable, controller try best effort for next event record",
				scaleDownDetail.NodeGroupID, err.Error())
		}
		nodegroupInfo, getErr := c.opt.Storage.GetNodeGroup(scaleDownDetail.NodeGroupID, &storage.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("[taskController]get nodegroup %s error:%s", scaleDownDetail.NodeGroupID, getErr.Error())
		}
		nodegroupInfo.DesiredSize -= scaleDownDetail.NodeNum
		nodegroupInfo.UpdatedTime = time.Now()
		nodegroupInfo.LastStatus = nodegroupInfo.Status
		nodegroupInfo.Status = storage.ScaleDownByTaskState
		nodegroupInfo.HookConfirm = false
		nodegroupInfo.Message = fmt.Sprintf("scale down %d nodes by task %s", scaleDownDetail.NodeNum, task.TaskID)
		// nolint
		if _, err = c.opt.Storage.UpdateNodeGroup(nodegroupInfo, &storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}); err != nil {
			return fmt.Errorf("update nodegroup %s desire size error:%s", scaleDownDetail.NodeGroupID, err.Error())
		}
	}
	return nil
}

func (c *taskController) traceExecutingTask() {
	blog.Infof("[taskController] begin to traceExecutingTask")
	tasks, err := c.opt.Storage.ListTasks(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list tasks error:%s", err.Error())
		return
	}
	wg := sync.WaitGroup{}
	pool, err := ants.NewPool(c.opt.Concurrency)
	if err != nil {
		blog.Errorf("[taskController] task controller internal error, ants.NewPool failed: %s", err.Error())
		return
	}
	defer pool.Release()
	for key := range tasks {
		wg.Add(1)
		task := tasks[key]
		err := pool.Submit(func() {
			c.traceAndUpdateStatus(task)
			wg.Done()
		})
		if err != nil {
			blog.Errorf("[taskController] submit task to ch pool err:%v", err)
		}
	}
	wg.Wait()
	blog.Infof("[taskController] completed traceExecutingTask")
}

func (c *taskController) traceAndUpdateStatus(task *storage.ScaleDownTask) {
	if !task.IsExecuting() {
		return
	}
	actions, err := c.opt.Storage.ListNodeGroupActionByTaskID(task.TaskID, &storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list action by taskID %s failed: %s", task.TaskID, err.Error())
		return
	}
	if len(actions) == 0 {
		blog.Errorf("[taskController] trace executing task %s but action not found", task.TaskID)
		return
	}
	isAllActionComplete := true
	scaleDownDevices := make([]string, 0)
	for _, action := range actions {
		scaleDownDevices = append(scaleDownDevices, action.NodeIPs...)
		nodegroup, err := c.opt.Storage.GetNodeGroup(action.NodeGroupID, &storage.GetOptions{ErrIfNotExist: true})
		if err != nil {
			blog.Errorf("[taskController] get nodegroup %s failed:%s", action.NodeGroupID, err.Error())
			return
		}
		completed := checkScaleDownComplete(nodegroup.NodeIPs, action.NodeIPs)
		if !completed {
			blog.Infof("[taskController] scale down action of task %s is not all completed", task.TaskID)
			isAllActionComplete = false
			metric.ReportActionHandleLatencyMetric(task.NodeGroupStrategy, action.Event, storage.ActionRunningState,
				action.ClusterID, action.NodeGroupID, action.CreatedTime)
			continue
		}
		action.Process = 100
		if _, err := c.opt.Storage.UpdateNodeGroupAction(action, &storage.UpdateOptions{}); err != nil {
			blog.Errorf("[taskController] update action process failed. action:%v, err:%s",
				action, err.Error())
			return
		}
		blog.Infof("[taskController] scale down action %s of task %s is completed", action.NodeGroupID, task.TaskID)
		metric.ReportActionHandleLatencyMetric(task.NodeGroupStrategy, action.Event, storage.ActionFinishedState,
			action.ClusterID, action.NodeGroupID, action.CreatedTime)
		if _, err := c.opt.Storage.DeleteNodeGroupAction(action, &storage.DeleteOptions{}); err != nil {
			blog.Errorf("[taskController] delete completed action failed. action:%v, err:%s",
				action, err.Error())
			return
		}
	}
	if isAllActionComplete {
		metric.ReportTaskFinishedMetric(task.NodeGroupStrategy, task.DrainDelay)
		metric.ReportTaskHandleLatencyMetric(task.NodeGroupStrategy, task.TaskID, storage.TaskFinishedState,
			task.DrainDelay, task.BeginExecuteTime)
		if !task.SpecifyScaleDown {
			if err := c.opt.ResourceManager.FillDeviceRecordIp(task.TaskID, scaleDownDevices); err != nil {
				blog.Errorf("[taskController] FillDeviceRecordIp error:%s", err.Error())
			}
		}
		_, err := c.opt.Storage.DeleteTask(task.TaskID, &storage.DeleteOptions{})
		if err != nil {
			blog.Errorf("[taskController] delete completed task %s failed: %s", task.TaskID, err.Error())
			return
		}
		blog.Infof("[taskController] all action is completed, delete task %s", task.TaskID)
	}
}

// nodeSelector 根据resource pool得到可以缩容的nodegroup，再根据权重计算出各nodegroup 可以缩容的节点数，最后选出具体node ip
func (c *taskController) nodeSelector(task *storage.ScaleDownTask) error {
	blog.Infof("[taskController] node Selector begins to choose node from strategy %s for scaleDown task %s",
		task.NodeGroupStrategy, task.TaskID)
	strategy, err := c.opt.Storage.GetNodeGroupStrategy(task.NodeGroupStrategy, &storage.GetOptions{})
	if err != nil {
		return fmt.Errorf("get nodegroup strategy by name %s error: %s", task.NodeGroupStrategy, err.Error())
	}
	// 根据标签筛选出可以缩容的节点数
	// 权重打分
	// 打分后确认nodegroup 节点ip
	// 先选择跟task的drainDelay一致的节点，如果不够，再选择drainDelay小于task的节点
	// 比如，有需要下架100台3天的节点，目前只有80台30天，20台2天，20台1天，那需要从2天中再选择20台
	selectedNodeGroups, err := c.filterAvailableNodes(task, strategy)
	if err != nil {
		return fmt.Errorf("nodeSelector filter available nodes failed. taskId: %s, strategyName:%s, error:%s",
			task.TaskID, strategy.Name, err.Error())
	}
	remainNum := task.TotalNum
	scaleDownDetailMap := make(map[string]*storage.ScaleDownDetail, 0)
	for index := range selectedNodeGroups {
		scaleDownNodegroup := selectedNodeGroups[index]
		scaleDownBalancer := newWeightBalancer(scaleDownNodegroup.GroupInfos, scaleDownNodegroup.NodeGroups)
		distributeNum := scaleDownNodegroup.Total
		if remainNum < distributeNum {
			distributeNum = remainNum
		}
		remainNum -= distributeNum
		result := scaleDownBalancer.distribute(distributeNum)
		for _, allocation := range result {
			if allocation.partition == 0 {
				continue
			}
			if _, ok := scaleDownDetailMap[allocation.NodeGroupID]; ok {
				scaleDownDetailMap[allocation.NodeGroupID].NodeIPs = append(scaleDownDetailMap[allocation.NodeGroupID].NodeIPs,
					scaleDownNodegroup.NodeGroups[allocation.NodeGroupID].NodeIPs[0:allocation.partition]...)
				scaleDownDetailMap[allocation.NodeGroupID].NodeNum += allocation.partition
			} else {
				scaleDownDetail := &storage.ScaleDownDetail{
					ConsumerID:  allocation.ConsumerID,
					NodeGroupID: allocation.NodeGroupID,
					ClusterID:   allocation.ClusterID,
					NodeIPs:     scaleDownNodegroup.NodeGroups[allocation.NodeGroupID].NodeIPs[0:allocation.partition],
					NodeNum:     allocation.partition,
				}
				scaleDownDetailMap[allocation.NodeGroupID] = scaleDownDetail
			}
		}
		if remainNum <= 0 {
			break
		}
	}
	for nodegroupID := range scaleDownDetailMap {
		task.ScaleDownGroups = append(task.ScaleDownGroups, scaleDownDetailMap[nodegroupID])
	}
	blog.Infof("[taskController] nodeSelector update task:%v", task)
	task, err = c.opt.Storage.UpdateTask(task, &storage.UpdateOptions{
		CreateIfNotExist: true,
	})
	if err != nil {
		blog.Errorf("[taskController] nodeSelector update task %s error:%s", task.TaskID, err.Error())
		return err
	}
	return nil
}

func (c *taskController) removeLabel(clusterID, taskID string) error {
	labels := map[string]interface{}{
		storage.NodeDrainTaskLabel: nil,
	}
	annotations := map[string]interface{}{
		storage.NodeDeadlineLabel: nil,
	}
	nodeList, err := c.opt.ClusterClient.ListClusterNodes(clusterID)
	if err != nil {
		return err
	}
	for _, node := range nodeList {
		if node.Labels != nil && node.Labels[storage.NodeDrainTaskLabel] == taskID {
			if err := c.opt.ClusterClient.UpdateNodeMetadata(clusterID, node.Name, labels, annotations); err != nil {
				return fmt.Errorf("UpdateNodeLabels error. ip:%s, labels:%s, error:%s", node.IP, labels, err.Error())
			}
		}
	}
	return nil
}

func (c *taskController) removeNotReadyNodes(scaleDownDetail *storage.ScaleDownDetail, taskID string) {
	label := map[string]interface{}{
		storage.NodeGroupLabel: scaleDownDetail.NodeGroupID,
	}
	nodeList, err := c.opt.ClusterClient.ListNodesByLabel(scaleDownDetail.ClusterID, label)
	if err != nil {
		blog.Errorf("[taskController] get cluster %s nodeList by nodegroup %s label failed: %s",
			scaleDownDetail.ClusterID, scaleDownDetail.NodeGroupID, err.Error())
		return
	}
	readyNodeIPs := make([]string, 0)
	readyNodeMaps := make(map[string]bool, 0)

	// 如果not ready，删掉
	for _, ip := range scaleDownDetail.NodeIPs {
		if nodeList[ip] != nil && nodeList[ip].Status == string(v1.ConditionTrue) {
			readyNodeIPs = append(readyNodeIPs, ip)
			readyNodeMaps[ip] = true
		}
	}
	scaleDownDetail.NodeIPs = readyNodeIPs
	scaleDownDetail.NodeNum = len(readyNodeIPs)
	for ip := range nodeList {
		// 如果nodegroup 里有节点带了标签，却不在scaleDownDetail的ip列表里，把标签移除
		if nodeList[ip].Labels != nil && nodeList[ip].Labels[storage.NodeDrainTaskLabel] == taskID {
			if _, ok := readyNodeMaps[ip]; !ok {
				labels := map[string]interface{}{
					storage.NodeDrainTaskLabel: nil,
				}
				annotations := map[string]interface{}{
					storage.NodeDeadlineLabel: nil,
				}
				if updateErr := c.opt.ClusterClient.UpdateNodeMetadata(scaleDownDetail.ClusterID, ip, labels,
					annotations); updateErr != nil {
					blog.Errorf("[taskController] UpdateNodeLabels error. ip:%s, error:%s", ip, updateErr.Error())
				}
			}
		}
	}
}

// NOCC:golint/funlen(设计如此)
// nolint
// 查每个nodegroup标签符合要求的节点
func (c *taskController) filterAvailableNodes(task *storage.ScaleDownTask,
	strategy *storage.NodeGroupMgrStrategy) ([]*storage.ScaleDownNodegroup, error) {
	selectNodeGroups := make([]*storage.ScaleDownNodegroup, 0)
	backupMap := make(map[int]*storage.ScaleDownNodegroup)
	totalNum := 0
	matchNodeGroups := &storage.ScaleDownNodegroup{
		GroupInfos: make([]*storage.GroupInfo, 0),
		NodeGroups: make(map[string]*storage.NodeGroup),
	}
	for _, groupInfo := range strategy.ElasticNodeGroups {
		nodeList, err := c.opt.ClusterClient.ListNodesByLabel(groupInfo.ClusterID,
			map[string]interface{}{storage.NodeGroupLabel: groupInfo.NodeGroupID})
		if err != nil {
			return nil, fmt.Errorf("list cluster %s nodes by nodegroup label %s err:%s", groupInfo.ClusterID,
				groupInfo.NodeGroupID, err.Error())
		}
		deviceNodeList, err := c.opt.ResourceManager.GetDeviceListByPoolID(groupInfo.ConsumerID,
			[]string{task.DevicePoolID}, nil)
		if err != nil {
			return nil, err
		}
		deviceMap := c.convertDeviceListToIP(deviceNodeList.Resources)
		var matchNodeIPs []string
		backUpIPMap := make(map[int][]string)
		for _, node := range nodeList {
			if _, ok := deviceMap[node.IP]; !ok {
				blog.Errorf("%s does not belong to devicepool %s", node.IP, task.DevicePoolID)
				continue
			}
			match, backup, backupDrainDelay := node.IsOptionalForScaleDown(storage.NodeDrainTaskLabel,
				storage.NodeDrainDelayLabel, task.TaskID, task.DrainDelay)
			if match {
				matchNodeIPs = append(matchNodeIPs, node.IP)
				continue
			}
			if backup {
				if _, ok := backUpIPMap[backupDrainDelay]; ok {
					backUpIPMap[backupDrainDelay] = append(backUpIPMap[backupDrainDelay], node.IP)
					continue
				}
				backupIPs := make([]string, 0)
				backupIPs = append(backupIPs, node.IP)
				backUpIPMap[backupDrainDelay] = backupIPs
			}
		}
		if len(matchNodeIPs) != 0 {
			nodegroup := &storage.NodeGroup{
				NodeGroupID: groupInfo.NodeGroupID,
				ClusterID:   groupInfo.ClusterID,
				NodeIPs:     matchNodeIPs,
			}
			matchNodeGroups.NodeGroups[nodegroup.NodeGroupID] = nodegroup
			matchNodeGroups.GroupInfos = append(matchNodeGroups.GroupInfos, groupInfo)
			matchNodeGroups.Total += len(matchNodeIPs)
			drainDelayStr := strings.Split(task.DrainDelay, "h")[0]
			drainDelayHour, _ := strconv.Atoi(drainDelayStr)
			matchNodeGroups.DrainDelayHour = drainDelayHour
			totalNum += len(matchNodeIPs)
		}

		for backupDrainDelay := range backUpIPMap {
			nodegroup := &storage.NodeGroup{
				NodeGroupID: groupInfo.NodeGroupID,
				ClusterID:   groupInfo.ClusterID,
				NodeIPs:     backUpIPMap[backupDrainDelay],
			}
			if _, ok := backupMap[backupDrainDelay]; ok {
				backupMap[backupDrainDelay].NodeGroups[nodegroup.NodeGroupID] = nodegroup
				backupMap[backupDrainDelay].GroupInfos = append(backupMap[backupDrainDelay].GroupInfos, groupInfo)
				backupMap[backupDrainDelay].Total += len(backUpIPMap[backupDrainDelay])
			} else {
				backUpInfo := &storage.ScaleDownNodegroup{
					GroupInfos: make([]*storage.GroupInfo, 0),
					NodeGroups: make(map[string]*storage.NodeGroup),
				}
				backUpInfo.DrainDelayHour = backupDrainDelay
				backUpInfo.NodeGroups[nodegroup.NodeGroupID] = nodegroup
				backUpInfo.GroupInfos = append(backUpInfo.GroupInfos, groupInfo)
				backUpInfo.Total += len(backUpIPMap[backupDrainDelay])
				backupMap[backupDrainDelay] = backUpInfo
			}
			totalNum += len(backUpIPMap[backupDrainDelay])
		}
	}
	selectNodeGroups = append(selectNodeGroups, matchNodeGroups)
	drainDelayGroup := make([]int, 0)
	for hour := range backupMap {
		drainDelayGroup = append(drainDelayGroup, hour)
	}
	sort.Ints(drainDelayGroup)
	sort.Sort(sort.Reverse(sort.IntSlice(drainDelayGroup)))
	for _, hour := range drainDelayGroup {
		selectNodeGroups = append(selectNodeGroups, backupMap[hour])
	}
	blog.Infof("[taskController] filter %d nodes from strategy %s", totalNum, strategy.Name)
	return selectNodeGroups, nil
}

func checkScaleDownComplete(nodegroupSlice, actionSlice []string) bool {
	nodegroupMap := make(map[string]struct{})
	for _, ip := range nodegroupSlice {
		nodegroupMap[ip] = struct{}{}
	}
	for _, ip := range actionSlice {
		if _, ok := nodegroupMap[ip]; ok {
			return false
		}
	}
	return true
}

func (c *taskController) getTerminatedTasks(storageTasks []*storage.ScaleDownTask,
	resourceTasks map[string]*storage.ScaleDownTask) []*storage.ScaleDownTask {
	terminatedTasks := make([]*storage.ScaleDownTask, 0)
	for _, storageTask := range storageTasks {
		if _, ok := resourceTasks[storageTask.TaskID]; ok {
			continue
		}
		realTaskId := strings.Split(storageTask.TaskID, "-specifyDevice-")[0]
		taskInfo, err := c.opt.ResourceManager.GetTaskByID(realTaskId, nil)
		if err != nil {
			blog.Errorf(err.Error())
			return nil
		}
		if taskInfo.Status != storage.TaskRequestingState {
			terminatedTasks = append(terminatedTasks, storageTask)
		}
	}
	return terminatedTasks
}

func (c *taskController) convertDeviceListToIP(deviceList []*storage.Resource) map[string]bool {
	ipMap := make(map[string]bool)
	for index := range deviceList {
		ipMap[deviceList[index].InnerIP] = true
	}
	return ipMap
}

func (c *taskController) getCrTasks(nodegroupStrategyList []*storage.NodeGroupMgrStrategy,
	scaleDownTaskList map[string]*storage.ScaleDownTask) {
	for _, strategy := range nodegroupStrategyList {
		if strategy.ElasticNodeGroups == nil || len(strategy.ElasticNodeGroups) == 0 {
			continue
		}
		if strategy.ElasticNodeGroups[0].ConsumerID == "" {
			blog.Errorf("[taskController] strategy %s consumerID is empty", strategy.Name)
			continue
		}
		task, listErr := c.opt.ResourceManager.ListTasksByConsumer(strategy.ElasticNodeGroups[0].ConsumerID, nil)
		if listErr != nil {
			blog.Errorf("[taskController] get strategy %s tasks err:%s", strategy.Name, listErr.Error())
			continue
		}
		if len(task) != 0 {
			for _, oneTask := range task {
				// 检查是否有指定节点下架的任务
				oneTask.NodeGroupStrategy = strategy.Name
				oneTask.BeginExecuteTime = oneTask.Deadline.Add(-time.Duration(strategy.Strategy.ScaleDownBeforeDDL) * time.Minute)
				scaleDownTaskList[oneTask.TaskID] = oneTask
				blog.Infof("add one task: %v", oneTask)
			}
		}
	}
}

func (c *taskController) getSelfTasks(nodegroupStrategyList []*storage.NodeGroupMgrStrategy,
	scaleDownTaskList map[string]*storage.ScaleDownTask) error {
	recordType := []int64{6}
	status := []int64{1}
	task, listErr := c.opt.ResourceManager.ListTasksByCond(recordType, status)
	if listErr != nil {
		blog.Errorf("[taskController] list strategy err:%s", listErr.Error())
		return listErr
	}
	for _, strategy := range nodegroupStrategyList {
		if strategy.ElasticNodeGroups == nil || len(strategy.ElasticNodeGroups) == 0 {
			continue
		}
		if strategy.ElasticNodeGroups[0].ConsumerID == "" {
			blog.Errorf("[taskController] strategy %s consumerID is empty", strategy.Name)
			continue
		}
		rangeTask := task

		for index := range rangeTask {
			bytes, marshalErr := json.Marshal(rangeTask[index])
			if marshalErr != nil {
				blog.Errorf("marshal task to bytes error:%s", marshalErr.Error())
				continue
			}
			copyTask := &storage.ScaleDownTask{}
			unmarshalErr := json.Unmarshal(bytes, copyTask)
			if unmarshalErr != nil {
				blog.Errorf("unmarshal task to copyTask error:%s", unmarshalErr.Error())
				continue
			}
			// 检查是否有指定节点下架的任务
			if copyTask.DeviceList != nil && len(copyTask.DeviceList) != 0 {
				matchDevice := make([]string, 0)
				scaleDownDetail := make(map[string]*storage.ScaleDownDetail)
				// 检查下架节点是否属于本策略
				for _, ip := range copyTask.DeviceList {
					nodeDetail, getErr := c.opt.ClusterClient.GetNodeDetail(ip)
					if getErr != nil {
						blog.Errorf(getErr.Error())
						continue
					}
					nodegroupId := nodeDetail.NodeGroupID
					for _, elasticGroup := range strategy.ElasticNodeGroups {
						if elasticGroup.NodeGroupID == nodegroupId {
							matchDevice = append(matchDevice, ip)
							if scaleDownDetail[nodegroupId] == nil {
								scaleDownDetail[nodegroupId] = &storage.ScaleDownDetail{
									ConsumerID:  elasticGroup.ConsumerID,
									NodeGroupID: nodegroupId,
									ClusterID:   elasticGroup.ClusterID,
									NodeIPs:     make([]string, 0),
									NodeNum:     0,
								}
							}
							scaleDownDetail[nodegroupId].NodeIPs = append(scaleDownDetail[nodegroupId].NodeIPs, ip)
							scaleDownDetail[nodegroupId].NodeNum++
						}
					}
				}
				if len(matchDevice) == 0 {
					blog.Infof("the devices of task %s are not belong to strategy %s, skip", rangeTask[index].TaskID,
						strategy.Name)
					continue
				}
				rangeTask[index].AllocatedNum += len(matchDevice)
				copyTask.TotalNum = len(matchDevice)
				copyTask.DeviceList = matchDevice
				copyTask.SpecifyScaleDown = true
				copyTask.ScaleDownGroups = make([]*storage.ScaleDownDetail, 0)
				for ng := range scaleDownDetail {
					copyTask.ScaleDownGroups = append(copyTask.ScaleDownGroups, scaleDownDetail[ng])
				}
				// 可能一个device record 包含的device来自不同的ng，不同的strategy，这里做个区分
				copyTask.TaskID = fmt.Sprintf("%s-specifyDevice-%s", rangeTask[index].TaskID, strategy.Name)
			}
			copyTask.NodeGroupStrategy = strategy.Name
			if strategy.Strategy.ScaleDownBeforeDDL != 0 {
				copyTask.BeginExecuteTime = rangeTask[index].Deadline.Add(
					-time.Duration(strategy.Strategy.ScaleDownBeforeDDL) * time.Minute)
			} else {
				copyTask.BeginExecuteTime = rangeTask[index].Deadline.Add(-5 * time.Minute)
			}
			scaleDownTaskList[copyTask.TaskID] = copyTask
			blog.Infof("add one task: %v", copyTask)
		}
	}
	return nil
}
