/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"

	"github.com/panjf2000/ants/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

const (
	nodeDrainDelayLabel = "node.bkbcs.tencent.com/drain-delay"
	nodeDeadlineLabel   = "node.bkbcs.tencent.com/deadline"
	nodeDrainTaskLabel  = "node.bkbcs.tencent.com/drain-task-id"
	nodeGroupLabel      = "bkbcs.tencent.com/nodegroupid"
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
	//init all custom Option
	for _, opt := range opts {
		opt(c.opt)
	}
	//init all dependent resource, such as storage, client and etc.
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
			//main loops
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
	//handleExpiredTask
	c.handleExpiredTask()
	// traceExecutingTask
	c.traceExecutingTask()
}

func (c *taskController) handleNormalTask() {
	blog.Infof("[taskController] begin handleNormalTask")
	// 获取所有strategy
	// 根据pool id和其中一个consumer id获取task，确认是哪一个策略需要执行缩容
	nodegroupStrategyList, err := c.opt.Storage.ListNodeGroupStrategiesByType(storage.HierarchicalStrategyType,
		&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list nodegroup strategy from storage failed:%s", err.Error())
		return
	}
	scaleDownTaskList := make(map[string]*storage.ScaleDownTask)
	for _, strategy := range nodegroupStrategyList {
		if strategy.ElasticNodeGroups == nil || len(strategy.ElasticNodeGroups) == 0 {
			continue
		}
		if strategy.ElasticNodeGroups[0].ConsumerID == "" {
			blog.Errorf("[taskController] strategy %s consumerID is empty", strategy.Name)
			continue
		}
		task, listErr := c.opt.ResourceManager.ListTasks(strategy.ResourcePool, strategy.ElasticNodeGroups[0].ConsumerID, nil)
		if err != nil {
			blog.Errorf("[taskController] get strategy %s tasks err:%s", strategy.Name, listErr.Error())
			return
		}
		if task != nil && len(task) != 0 {
			for _, oneTask := range task {
				oneTask.NodeGroupStrategy = strategy.Name
				oneTask.BeginExecuteTime = oneTask.Deadline.Add(-time.Duration(strategy.Strategy.ScaleDownBeforeDDL) * time.Minute)
				scaleDownTaskList[oneTask.TaskID] = oneTask
			}
		}
	}
	// 确认没有返回的task是否terminated，如果是，更新状态
	tasks, err := c.opt.Storage.ListTasks(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("[taskController] list task error:%s", err.Error())
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
	}

	for _, task := range scaleDownTaskList {
		if time.Now().Before(task.BeginExecuteTime) {
			c.handleOneNormalTask(task)
		}
	}
	blog.Infof("[taskController] completed handleNormalTask")
}

func (c *taskController) handleOneNormalTask(task *storage.ScaleDownTask) {
	var err error
	// only update info from resource manager
	updateTask := &storage.ScaleDownTask{
		TaskID:            task.TaskID,
		NodeGroupStrategy: task.NodeGroupStrategy,
		TotalNum:          task.TotalNum,
		DrainDelay:        task.DrainDelay,
		Deadline:          task.Deadline,
		BeginExecuteTime:  task.BeginExecuteTime,
		UpdatedTime:       time.Now(),
		Status:            task.Status,
	}
	updateTask, err = c.opt.Storage.UpdateTask(updateTask, &storage.UpdateOptions{
		CreateIfNotExist: true,
	})
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
	if originTotal == task.TotalNum {
		//先检查ScaleDownGroups里的节点状态是否都是ready，如果不是，从切片里删除
		//检查nodegroup里带标签的节点跟切片里的是否一致，不在切片里但有标签的，去掉
		//如果执行时间在1小时内，不操作？
		wg := sync.WaitGroup{}
		pool, initErr := ants.NewPool(c.opt.Concurrency)
		if err != nil {
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
	// label node
	for _, group := range updateTask.ScaleDownGroups {
		labels := map[string]interface{}{
			nodeDrainDelayLabel: updateTask.DrainDelay,
			nodeDrainTaskLabel:  updateTask.TaskID,
		}
		annotations := map[string]interface{}{
			nodeDeadlineLabel: updateTask.Deadline.Format(time.RFC3339),
		}
		for _, ip := range group.NodeIPs {
			blog.Infof("[taskController] update node %s labels", ip)
			if err = c.opt.ClusterClient.UpdateNodeMetadata(group.ClusterID, ip, labels, annotations); err != nil {
				blog.Errorf("[taskController] UpdateNodeLabels error. task id: %s, clusterID:%s, nodeIP:%s, labels:%s, error:%s",
					task.TaskID, group.ClusterID, ip, labels, err.Error())
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
	for _, task := range tasks {
		if task.BeginExecuteTime.Before(time.Now()) {
			expiredTasks = append(expiredTasks, task)
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
		if err != nil {
			return fmt.Errorf("get nodegroup %s error:%s", scaleDownDetail.NodeGroupID, getErr.Error())
		}
		nodegroupInfo.DesiredSize = nodegroupInfo.DesiredSize - scaleDownDetail.NodeNum
		nodegroupInfo.UpdatedTime = time.Now()
		nodegroupInfo.LastStatus = nodegroupInfo.Status
		nodegroupInfo.Status = storage.ScaleDownByTaskState
		nodegroupInfo.HookConfirm = false
		nodegroupInfo.Message = fmt.Sprintf("scale down %d nodes by task %s", scaleDownDetail.NodeNum, task.TaskID)
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
	for _, action := range actions {
		nodegroup, err := c.opt.Storage.GetNodeGroup(action.NodeGroupID, &storage.GetOptions{ErrIfNotExist: true})
		if err != nil {
			blog.Errorf("[taskController] get nodegroup %s failed:%s", action.NodeGroupID, err.Error())
			return
		}
		completed := checkScaleDownComplete(nodegroup.NodeIPs, action.NodeIPs)
		if !completed {
			blog.Infof("[taskController] scale down action of task %s is not all completed", task.TaskID)
			isAllActionComplete = false
			continue
		}
		action.Process = 100
		if _, err := c.opt.Storage.UpdateNodeGroupAction(action, &storage.UpdateOptions{}); err != nil {
			blog.Errorf("[taskController] update action process failed. action:%v, err:%s",
				action, err.Error())
			return
		}
		blog.Infof("[taskController] scale down action %s of task %s is completed", action.NodeGroupID, task.TaskID)
		if _, err := c.opt.Storage.DeleteNodeGroupAction(action, &storage.DeleteOptions{}); err != nil {
			blog.Errorf("[taskController] delete completed action failed. action:%v, err:%s",
				action, err.Error())
			return
		}

	}
	if isAllActionComplete {
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
	blog.Infof("[taskController] check the nodegroupStrategy, strategyName:%s, reservedDays:%d",
		strategy.Name, strategy.Strategy.Buffer.ReservedDays)
	groupInfos, nodegroups, totalNum, err := c.filterAvailableNodes(task.TaskID, strategy)
	if err != nil {
		return fmt.Errorf("nodeSelector filter available nodes failed. taskId: %s, strategyName:%s, error:%s",
			task.TaskID, strategy.Name, err.Error())
	}
	scaleDownBalancer := newWeightBalancer(groupInfos, nodegroups)
	distributeNum := totalNum
	if task.TotalNum < totalNum {
		distributeNum = task.TotalNum
	}
	result := scaleDownBalancer.distribute(distributeNum)
	for _, allocation := range result {
		if allocation.partition == 0 {
			continue
		}
		scaleDownDetail := &storage.ScaleDownDetail{
			ConsumerID:  allocation.ConsumerID,
			NodeGroupID: allocation.NodeGroupID,
			ClusterID:   allocation.ClusterID,
			NodeIPs:     nodegroups[allocation.NodeGroupID].NodeIPs[0:allocation.partition],
			NodeNum:     allocation.partition,
		}
		task.ScaleDownGroups = append(task.ScaleDownGroups, scaleDownDetail)
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
		nodeDrainDelayLabel: nil,
		nodeDrainTaskLabel:  nil,
	}
	annotations := map[string]interface{}{
		nodeDeadlineLabel: nil,
	}
	nodeList, err := c.opt.ClusterClient.ListClusterNodes(clusterID)
	if err != nil {
		return err
	}
	for _, node := range nodeList {
		if node.Labels != nil && node.Labels[nodeDrainTaskLabel] == taskID {
			if err := c.opt.ClusterClient.UpdateNodeMetadata(clusterID, node.Name, labels, annotations); err != nil {
				return fmt.Errorf("UpdateNodeLabels error. ip:%s, labels:%s, error:%s", node.IP, labels, err.Error())
			}
		}
	}
	return nil
}

func (c *taskController) removeNotReadyNodes(scaleDownDetail *storage.ScaleDownDetail, taskID string) {
	label := map[string]interface{}{
		nodeGroupLabel: scaleDownDetail.NodeGroupID,
	}
	nodeList, err := c.opt.ClusterClient.ListNodesByLabel(scaleDownDetail.ClusterID, label)
	if err != nil {
		blog.Errorf("[taskController] get cluster %s nodeList by nodegroup %s label failed: %s",
			scaleDownDetail.ClusterID, scaleDownDetail.NodeGroupID, err.Error())
		return
	}
	readyNodeIPs := make([]string, 0)
	readyNodeMaps := make(map[string]struct{}, 0)

	// 如果not ready，删掉
	for _, ip := range scaleDownDetail.NodeIPs {
		if nodeList[ip] != nil && nodeList[ip].Status == string(v1.ConditionTrue) {
			readyNodeIPs = append(readyNodeIPs, ip)
		}
	}
	scaleDownDetail.NodeIPs = readyNodeIPs
	scaleDownDetail.NodeNum = len(readyNodeIPs)
	for ip := range nodeList {
		// 如果nodegroup 里有节点带了标签，却不在scaleDownDetail的ip列表里，把标签移除
		if nodeList[ip].Labels != nil && nodeList[ip].Labels[nodeDrainTaskLabel] == taskID {
			if _, ok := readyNodeMaps[ip]; !ok {
				labels := map[string]interface{}{
					nodeDrainDelayLabel: nil,
					nodeDrainTaskLabel:  nil,
				}
				annotations := map[string]interface{}{
					nodeDeadlineLabel: nil,
				}
				if err := c.opt.ClusterClient.UpdateNodeMetadata(scaleDownDetail.ClusterID, ip, label, annotations); err != nil {
					blog.Errorf("[taskController] UpdateNodeLabels error. ip:%s, labels:%s, error:%s", ip, labels, err.Error())
				}
			}
		}
	}
}

// 查每个nodegroup标签符合要求的节点
func (c *taskController) filterAvailableNodes(taskID string,
	strategy *storage.NodeGroupMgrStrategy) ([]*storage.GroupInfo, map[string]*storage.NodeGroup, int, error) {
	selectedGroupInfo := make([]*storage.GroupInfo, 0)
	selectedNodegroup := make(map[string]*storage.NodeGroup)
	totalNum := 0
	for _, groupInfo := range strategy.ElasticNodeGroups {
		nodeList, err := c.opt.ClusterClient.ListNodesByLabel(groupInfo.ClusterID,
			map[string]interface{}{nodeGroupLabel: groupInfo.NodeGroupID})
		if err != nil {
			return nil, nil, 0, fmt.Errorf("list cluster %s nodes by nodegroup label %s err:%s", groupInfo.ClusterID,
				groupInfo.NodeGroupID, err.Error())
		}

		var nodeIPs []string
		for _, node := range nodeList {
			if node.IsOptionalForScaleDown(nodeDrainTaskLabel, taskID) {
				nodeIPs = append(nodeIPs, node.IP)
			}
		}
		if len(nodeIPs) != 0 {
			nodegroup := &storage.NodeGroup{
				NodeGroupID: groupInfo.NodeGroupID,
				ClusterID:   groupInfo.ClusterID,
				NodeIPs:     nodeIPs,
			}
			selectedNodegroup[nodegroup.NodeGroupID] = nodegroup
			selectedGroupInfo = append(selectedGroupInfo, groupInfo)
			totalNum += len(nodeIPs)
		}
	}
	blog.Infof("[taskController] filter %d nodes from strategy %s", totalNum, strategy.Name)
	return selectedGroupInfo, selectedNodegroup, totalNum, nil
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
		taskInfo, err := c.opt.ResourceManager.GetTaskByID(storageTask.TaskID, nil)
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
