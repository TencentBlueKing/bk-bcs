/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// NodeGroupController for
type NodeGroupController interface {
	Init(opts ...Option) error
	Options() *Options
	Run(cxt context.Context)
}

// NewController create nodegroup implementation
func NewController(option *Options) NodeGroupController {
	return &control{
		opt: option,
	}
}

// control inner implementation controller for NodeGroupMgrStrategy
type control struct {
	opt *Options
}

// Init NodeGroupController init implementation
func (c *control) Init(opts ...Option) error {
	// init all custom Option
	for _, opt := range opts {
		opt(c.opt)
	}
	// init all dependent resource, such as storage, client and etc.

	if c.opt.ResourceManager == nil {
		blog.Errorf("Controller lost resource-manager interface in Init Stage")
		return fmt.Errorf("controller lost resource-manager instance")
	}
	if c.opt.Storage == nil {
		blog.Errorf("Controller lost storage interface in Init Stage")
		return fmt.Errorf("controller lost storage instance")
	}
	return nil
}

// Options NodeGroupController implementation
func (c *control) Options() *Options {
	return c.opt
}

// Run NodeGroupController implementation
func (c *control) Run(cxt context.Context) {
	tick := time.NewTicker(time.Second * time.Duration(c.opt.Interval))
	for {
		select {
		case now := <-tick.C:
			// main loops
			blog.Infof("##############ticker: %s################", now.Format(time.RFC3339))
			c.controllerLoops()
		case <-cxt.Done():
			blog.Infof("NodeGroupMgr Controller is asked to exit")
			return
		}
	}
}

// controllerLoops for NodeGroupMgr logic loop
func (c *control) controllerLoops() {
	// protection for main logic loop
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("panic in NodeGroup Controller, info: %v", r)
		}
	}()
	// list all strategies from storage
	strategies, err := c.opt.Storage.ListNodeGroupStrategies(&storage.ListOptions{})
	if err != nil {
		blog.Errorf("controller check all nodegroup manage strategies failed, %s", err.Error())
		return
	}
	blog.Infof("controller got %d NodeGroupMgrStrategy", len(strategies))

	for _, strategy := range strategies {
		c.handleStrategy(strategy)
		blog.Infof("strategy %s for ResourcePool %s has been processed completely",
			strategy.Name, strategy.ResourcePool)
	}
}

// handleStrategy handle one strategy for specified ResourcePool
// make decision that elasticNodeGroup need to ScaleUp or ScaleDown
func (c *control) handleStrategy(strategy *storage.NodeGroupMgrStrategy) {
	var msg, status string
	var globalErr error
	defer func() {
		if len(status) == 0 {
			return
		}
		// Controller only records failures of bcs services that it relys on
		// when handling strategy.
		// Controller try best effort to logs failure information when
		// database(storage) is going down.
		if strategy.Status == nil {
			strategy.Status = &storage.State{}
		}
		strategy.Status.UpdatedTime = time.Now()
		strategy.Status.Message = msg
		strategy.Status.LastStatus = strategy.Status.Status
		strategy.Status.Status = status
		if globalErr != nil {
			strategy.Status.Error = globalErr.Error()
		}

		if _, err := c.opt.Storage.UpdateNodeGroupStrategy(strategy, &storage.UpdateOptions{}); err != nil {
			blog.Errorf("controller update strategy %s failed, %s. status %s, message %s",
				strategy.Name, err.Error(), status, msg)
		}
	}()

	blog.Infof("controller handle strategy %s for ResourcePool %s", strategy.Name, strategy.ResourcePool)
	// query relative ResourcePool information from resource-manager
	resourcePool, getErr := c.opt.ResourceManager.GetResourcePool(strategy.ResourcePool, nil)
	if getErr != nil {
		blog.Errorf("controller got ResourcePool %s from resource-manager failed, %s",
			strategy.ResourcePool, getErr.Error())
		globalErr = getErr
		status = storage.ErrState
		msg = fmt.Sprintf("get dependent resourcepool %s failed", strategy.ResourcePool)
		return
	}
	// loading local ResourcePool state & local action
	allActions := make([]*storage.NodeGroupAction, 0)
	for _, nodegroup := range strategy.ElasticNodeGroups {
		actions, err := c.opt.Storage.ListNodeGroupAction(nodegroup.NodeGroupID, &storage.ListOptions{})
		if err != nil {
			blog.Errorf("controller got %s/%s NodeGroupAction failed, %s. wait next tick(best effort to storage)",
				nodegroup.ClusterID, nodegroup.NodeGroupID, err.Error())
			// local storage system error, controller try the best effort to
			// log strategy state into storage
			globalErr = err
			status = storage.ErrState
			msg = fmt.Sprintf("get nodegroup %s relative actions failure", nodegroup.NodeGroupID)
			return
		}
		allActions = append(allActions, actions...)
	}
	// check if controller need to ScaleDown elasticNodeGroup for ResourcePool
	scaleDownNum, isDown := isElasticNodeGroupEssentialForScaleDown(strategy, resourcePool)
	if isDown {
		status = storage.ScaleDownState
		if len(allActions) == 0 {
			// no operation before, handle it simply
			if err := c.handleElasticNodeGroupScaleDown(strategy, scaleDownNum); err != nil {
				// local storage error, controller try the best effort to update Mgr Strategy Status
				blog.Errorf("Controller scaleDown %d resources for resourcePool %s failed: %s. "+
					"wait next tick(best effort to storage)",
					scaleDownNum, strategy.ResourcePool, err.Error())
				globalErr = err
				msg = fmt.Sprintf("storage failure when controller scaledown elasticNodeGroups %d nodes", scaleDownNum)
				return
			}
			// scale down operation success, try to update MgrStrategy Status
			msg = fmt.Sprintf("elastic nodegroup is scaling down %d nodes", scaleDownNum)
			blog.Infof("Controller handle strategy %s for resourcePool %s result, %s",
				strategy.Name, strategy.ResourcePool, msg)
			return
		}
		// tracing scaleDown actions
		if err := c.tracingScaleDownAction(strategy, scaleDownNum, allActions); err != nil {
			blog.Errorf("Controller tracks scaledDown actions under strategy %s failed, "+
				"wait next tick(best effort to storage)", strategy.Name)
			globalErr = err
			msg = fmt.Sprintf("storage failure when controller tracks elasticNodeGroups scaledown %d nodes", scaleDownNum)
			return
		}
		msg = fmt.Sprintf("exist nodegroup actions track %d scaledown resources", scaleDownNum)
		blog.Infof("Controller handles strategy %s for resourcePool %s result,  %s",
			strategy.Name, strategy.ResourcePool, msg)
		return
	}
	// check if controller need to ScaleUp elasticNodeGroup for ResourcePool is idle
	scaleUpNum, isUp := isResourcePoolIdleForScaleUp(strategy, resourcePool)
	if isUp {
		status = storage.ScaleUpState
		if len(allActions) == 0 {
			if err := c.handleElasticNodeGroupScaleUp(strategy, scaleUpNum); err != nil {
				blog.Errorf("Controller scaleUp %d nodes from resourcePool %s failed: %s, "+
					"wait next tick(best effort to storage)", scaleUpNum, strategy.ResourcePool, err.Error())
				globalErr = err
				msg = fmt.Sprintf("storage failure when controller scaleup %d elasticNodeGroups nodes", scaleUpNum)
				return
			}
			// scale up operation success, try to update MgrStrategy Status
			msg = fmt.Sprintf("elastic nodegroup is scaling up %d nodes", scaleUpNum)
			blog.Infof("Controller handle strategy %s for resourcePool %s result, %s",
				strategy.Name, strategy.ResourcePool, msg)
			return
		}
		// tracing scaleUp actions
		if err := c.tracingScaleUpAction(strategy, scaleUpNum, allActions); err != nil {
			blog.Errorf("Controller tracks scaleUp actions under strategy %s failed, "+
				"wait next tick(best effort to storage)", strategy.Name)
			globalErr = err
			msg = fmt.Sprintf("storage failure when controller tracks elasticNodeGroups scaleUp %d nodes", scaleUpNum)
			return
		}
		msg = fmt.Sprintf("exist nodegroup actions track %d scaleUp resources", scaleUpNum)
		blog.Infof("Controller handles strategy %s for resourcePool %s result,  %s",
			strategy.Name, strategy.ResourcePool, msg)
		return
	}
	// no scaleUp or scaleDown action means that resourcePool is stable.
	// try to update NodeGroupMgrStrategy status, waiting for NodeGroupAction processes completely.
	status = storage.StableState
	msg = "no actions required, resourcePool is stable"
	blog.Infof(msg)
	// completely handling NodeGroupAction lifecycle
	for key, action := range allActions {
		clean := false
		if action.IsTerminated() {
			blog.Infof("Nodegroup %s Action %s is terminated when resource pool is stable,",
				action.NodeGroupID, action.Event)
			clean = true
		}
		if action.Event == storage.ScaleDownState && action.IsTimeout(strategy.Strategy.ScaleDownDelay) {
			blog.Infof("Nodegroup %s Action %s is timeout when resource pool is stable,",
				action.NodeGroupID, action.Event)
			clean = true
		}
		if action.Event == storage.ScaleUpState && action.IsTimeout(strategy.Strategy.ScaleUpDelay) {
			blog.Infof("Nodegroup %s Action %s is timeout when resource pool is stable,",
				action.NodeGroupID, action.Event)
			clean = true
		}
		if clean {
			if _, err := c.opt.Storage.DeleteNodeGroupAction(action, &storage.DeleteOptions{}); err != nil {
				blog.Errorf("controller clean NodeGroupAction %s-%s met storage failure, %s. wait next tick",
					action.NodeGroupID, action.Event, err.Error())
			} else {
				blog.Infof("Controller clean NodeGroupAction %s-%s successfully", key, action.Event)
			}
		}
	}
}

// tracingScaleDownAction track exist nodegroup scaledown action
func (c *control) tracingScaleDownAction(strategy *storage.NodeGroupMgrStrategy,
	scaleDownNum int, actions []*storage.NodeGroupAction) error {
	// check exist nodegroup action is suitable for tracing
	trackedActions := make(map[string]*storage.NodeGroupAction)
	for _, action := range actions {
		clean, err := c.cleanUnexpectedNodeGroupActions(action, storage.ScaleDownState)
		if err != nil {
			return fmt.Errorf("clean unexpected nodegroup %s action %s failure, %s",
				action.NodeGroupID, action.Event, err.Error())
		}
		if !clean {
			// expected scaleDown nodegroup actions, keep it tracked
			trackedActions[action.NodeGroupID] = action
			blog.V(5).Infof("nodegroup %s action %s %d nodes is under tracing when try to "+
				"scaleDown elasticNodeGroup %d nodes",
				action.NodeGroupID, action.Event, action.DeltaNum, scaleDownNum)
		}
	}
	nodeGroups, err := c.listElasticNodeGroups(strategy.ElasticNodeGroups)
	if err != nil {
		return fmt.Errorf("load elasticNodeGroups info failed, %s", err.Error())
	}
	// verify nodegroup scaleDown action are still working on
	releaseNum := upComingElasticResources(trackedActions, nodeGroups, storage.ScaleDownState, strategy)
	if releaseNum >= scaleDownNum {
		// update tracked action for next logic tick
		blog.Infof("Upcoming scaleDown %d nodes from elasticNodeGroups can satisfy resourcePool %s "+
			"requirements (%d nodes). wait next tick", releaseNum, strategy.ResourcePool, scaleDownNum)
		if releaseNum-scaleDownNum >= 3 { // todo(DeveloperJim): fix for configuration
			blog.Warnf("ScaleDown too many resources in last decision or resources had been released from "+
				"ReservedNodeGroup, it's risky for stable. details: "+
				"strategy %s, elasticNodeGroup releaseNum %d, resourcePool need %d",
				strategy.Name, releaseNum, scaleDownNum)
		}
		return nil
	}
	// tracked actions can not satisfy scaledown resources,
	// rebalance resources allocation
	scaleMore := scaleDownNum - releaseNum
	blog.Infof("ElasticNodeGroups only release %d nodes, resourcePool still need %d nodes, try to reallocation",
		releaseNum, scaleDownNum)
	scaleDownBalancer := newWeightBalancer(strategy.ElasticNodeGroups, nodeGroups)
	allocations := scaleDownBalancer.distribute(scaleMore)
	for _, allo := range allocations {
		// todo(DeveloperJim): handle if partition == 0
		nodegroup := nodeGroups[allo.NodeGroupID]
		original := nodegroup.DesiredSize
		nodegroup.DesiredSize = original - allo.partition
		if nodegroup.DesiredSize < 0 {
			nodegroup.DesiredSize = 0
		}
		nodegroup.LastScaleUpTime = time.Now()
		nodegroup.LastStatus = nodegroup.Status
		nodegroup.Status = storage.ScaleDownState
		nodegroup.HookConfirm = false
		nodegroup.Message = fmt.Sprintf("nodegroup %s scaledown additional %d nodes from %d to %d",
			nodegroup.NodeGroupID, allo.partition, original, nodegroup.DesiredSize)
		blog.Infof("%s", nodegroup.Message)
		// ready to update NodeGroup information for cluster-autoscaler webhook request
		if _, err := c.opt.Storage.UpdateNodeGroup(nodegroup,
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}); err != nil {
			blog.Errorf("controller update nodegroup %s/%s more scaleDown info into storage failure, %s",
				nodegroup.ClusterID, nodegroup.NodeGroupID, err.Error())
			return err
		}
		// record NodeGroupAction for tracing
		action, found := trackedActions[nodegroup.NodeGroupID]
		var reason string
		if found {
			// found progress NodeGroupAction, partial update
			reason = fmt.Sprintf("scaledown %d more resources, oldDesiredSize %d, newDesiredSize %d",
				allo.partition, original, nodegroup.DesiredSize)
			action.UpdatedTime = time.Now()
			action.NewDesiredNum = nodegroup.DesiredSize
			action.NodeIPs = nodegroup.NodeIPs
			action.DeltaNum = nodegroup.DesiredSize - action.OriginalDesiredNum
			action.Status = storage.ScaleDownState
		} else {
			reason = "tracing empty scaleDown action, lastStatus is scaleUp, recreate scaleDown action"
			action = generateNodeGroupAction(nodegroup, storage.ScaleDownState,
				original, allo.partition, nodegroup.DesiredSize)
			blog.Infof("NodeGroup %s %s", action.NodeGroupID, reason)
		}
		if err := c.opt.Storage.CreateNodeGroupAction(action,
			&storage.CreateOptions{OverWriteIfExist: true}); err != nil {
			blog.Errorf("controller force create nodegroup %s scaleDown action failed, info: %s",
				nodegroup.NodeGroupID, err.Error())
			return fmt.Errorf("force create %s NodeGroupAction to storage failure", nodegroup.NodeGroupID)
		}
		// record NodeGroupEvent for manually tracing
		event := generateNodeGroupEvent(nodegroup, original, storage.ScaleDownState,
			reason, fmt.Sprintf("try to scaledown %d more nodes", allo.partition))
		if err := c.opt.Storage.CreateNodeGroupEvent(event, &storage.CreateOptions{}); err != nil {
			// event only used for administrator tracing issue manually.
			// failure of event operation is tolerable.
			blog.Errorf("controller create nodegroup %s scaleDown record failure, info: %s."+
				"failure is tolerable, controller try best effort for next event record",
				nodegroup.NodeGroupID, err.Error())
		}
		blog.Infof("nodegroup %s tracks ScaleDownAction in storage completely", nodegroup.NodeGroupID)
	}
	return nil
}

// tracingScaleUpAction track
func (c *control) tracingScaleUpAction(strategy *storage.NodeGroupMgrStrategy,
	scaleUpNum int, actions []*storage.NodeGroupAction) error {
	// check exist nodegroup action is suitable for tracing
	trackedActions := make(map[string]*storage.NodeGroupAction)
	for _, action := range actions {
		clean, err := c.cleanUnexpectedNodeGroupActions(action, storage.ScaleUpState)
		if err != nil {
			return err
		}
		if !clean {
			// expect scaleUp nodegroup actions, keep it tracked
			trackedActions[action.NodeGroupID] = action
			blog.Infof("nodegroup %s/%s action %s is expected, keep it tracked",
				action.ClusterID, action.NodeGroupID, action.Event)
		}
	}
	nodeGroups, err := c.listElasticNodeGroups(strategy.ElasticNodeGroups)
	if err != nil {
		return fmt.Errorf("load elasticNodeGroups info failed, %s", err.Error())
	}
	// verify nodegroup scaleUp action ares still working on,
	// and upcoming resources can meet current requirement.
	upComing := upComingElasticResources(trackedActions, nodeGroups, storage.ScaleUpState, strategy)
	if upComing >= scaleUpNum {
		// update tracked action for next logic tick
		blog.Infof("Upcoming scaleUp %d nodes from resourcePool %s to elasticNodeGroups satisfy "+
			"requirements (%d nodes). wait next tick", upComing, strategy.ResourcePool, scaleUpNum)
		if upComing-scaleUpNum >= 3 { // todo(DeveloperJim): fix for configuration
			blog.Warnf("ScaleUp too many resources in last decision or resources had been consumed for "+
				"ReservedNodeGroup, it's risky for stable. details: "+
				"strategy %s, elasticNodeGroup upComing %d, resourcePool can release %d",
				strategy.Name, upComing, scaleUpNum)
		}
		return nil
	}
	// upComing resources can not meet requirement.
	// try to allocate redundant resources to different nodegroups
	scaleMore := scaleUpNum - upComing
	blog.Infof("upComing %d nodes, resourcePool can release %d nodes, try to reallocation", upComing, scaleUpNum)
	scaleUpBalancer := newSimpleBalancer(strategy.ElasticNodeGroups)
	allocations := scaleUpBalancer.distribute(scaleMore)
	for _, allo := range allocations {
		nodegroup := nodeGroups[allo.NodeGroupID]
		original := nodegroup.DesiredSize
		// !pay more attention, controller consider NodeGroup.MaxSize
		// !is closed to resourcePool max size
		nodegroup.DesiredSize = original + allo.partition
		nodegroup.LastScaleUpTime = time.Now()
		nodegroup.LastStatus = nodegroup.Status
		nodegroup.Status = storage.ScaleUpState
		nodegroup.HookConfirm = false
		nodegroup.Message = fmt.Sprintf("nodegroup %s scaleup additional %d nodes from %d to %d",
			nodegroup.NodeGroupID, allo.partition, original, nodegroup.DesiredSize)
		blog.Infof("%s", nodegroup.Message)
		// ready to update NodeGroup information for cluster-autoscaler webhook request
		if _, err := c.opt.Storage.UpdateNodeGroup(nodegroup,
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}); err != nil {
			blog.Errorf("controller update nodegroup %s/%s scaleUp state into storage failure, %s",
				nodegroup.ClusterID, nodegroup.NodeGroupID, err.Error())
			return err
		}
		// record NodeGroupAction for tracing
		action, found := trackedActions[nodegroup.NodeGroupID]
		var reason string
		if found {
			// found progress NodeGroupAction, partial update
			reason = fmt.Sprintf("scaleUp %d more resources, oldDesiredSize %d, newDesiredSize %d",
				allo.partition, original, nodegroup.DesiredSize)
			action.UpdatedTime = time.Now()
			action.NewDesiredNum = nodegroup.DesiredSize
			action.NodeIPs = nodegroup.NodeIPs
			action.DeltaNum = nodegroup.DesiredSize - action.OriginalDesiredNum
			action.Status = storage.ScaleUpState
		} else {
			reason = "tracing empty scaleUp action, lastStatus is scaleDown, recreate scaleUp action"
			action = generateNodeGroupAction(nodegroup, storage.ScaleUpState,
				original, allo.partition, nodegroup.DesiredSize)
			blog.Infof("NodeGroup %s %s", action.NodeGroupID, reason)
		}
		if err := c.opt.Storage.CreateNodeGroupAction(action,
			&storage.CreateOptions{OverWriteIfExist: true}); err != nil {
			blog.Errorf("controller create nodegroup %s scaleUp action failed, info: %s",
				nodegroup.NodeGroupID, err.Error())
			return fmt.Errorf("create %s NodeGroupAction to storage failure", nodegroup.NodeGroupID)
		}

		// record NodeGroupEvent for manually tracing
		event := generateNodeGroupEvent(nodegroup, original, storage.ScaleUpState,
			reason, fmt.Sprintf("try to scaleup %d more nodes", allo.partition))
		if err := c.opt.Storage.CreateNodeGroupEvent(event, &storage.CreateOptions{}); err != nil {
			// event only used for administrator tracing issue manually.
			// failure of event operation is tolerable.
			blog.Errorf("controller create nodegroup %s scaleUp record failure, info: %s."+
				"failure is tolerable, controller try best effort for next event record",
				nodegroup.NodeGroupID, err.Error())
		}
		blog.Infof("nodegroup %s tracks ScaleUpAction in storage completely", nodegroup.NodeGroupID)
	}
	return nil
}

// cleanUnexpectedNodeGroupActions 清理掉操作方向不一致的记录.
// return:
//   bool, true if nodegroupAction was delete， otherwise false
//   error, if any error happened
func (c *control) cleanUnexpectedNodeGroupActions(action *storage.NodeGroupAction, expectedState string) (bool, error) {
	if action.Event != expectedState {
		blog.Infof("exist %s/%s nodegroup action is not %s, clean outdated action. details: %+v",
			action.ClusterID, action.NodeGroupID, expectedState, action)
		// clean nodegroup action
		if _, err := c.opt.Storage.DeleteNodeGroupAction(action, &storage.DeleteOptions{}); err != nil {
			blog.Errorf("controller cleans outdated nodegroupAction %s/%s failed, event %s, %s",
				action.ClusterID, action.NodeGroupID, action.Event, err.Error())
			return false, fmt.Errorf("storage broken, %s", err.Error())
		}
		return true, nil
	}
	return false, nil
}

// listElasticNodeGroups list all specified nodeGroups defined in elastic state.
// !pay more attention: NodeGroup information in storage are only created after
// !cluster-autoscaler making webhook requests.
func (c *control) listElasticNodeGroups(elasticGroups []*storage.GroupInfo) (map[string]*storage.NodeGroup, error) {
	// try to get elastic nodeGroups information
	nodeGroups := make(map[string]*storage.NodeGroup)
	for _, info := range elasticGroups {
		nodegroup, err := c.opt.Storage.GetNodeGroup(info.NodeGroupID, &storage.GetOptions{})
		if err != nil {
			blog.Errorf("controller gets nodeGroups %s/%s in local storage failure, %s",
				info.ClusterID, info.NodeGroupID, err.Error())
			return nil, err
		}
		// todo(DeveloperJim): no nodegroups found in storage, try to query from cluster-manager
		if nodegroup == nil {
			blog.Warnf("Controller get no such nodegroup %s/%s in local storage, waiting for webhook request",
				info.ClusterID, info.NodeGroupID)
			return nil, fmt.Errorf("no nodegroup %s/%s in local storage", info.ClusterID, info.NodeGroupID)
		}
		nodegroup.ClusterID = info.ClusterID
		nodeGroups[info.NodeGroupID] = nodegroup
		blog.Infof("nodegroup %s original information, maxSize: %d, minSize: %d, desiredSize: %d, "+
			"upComing: %d, status: %s, lastStatus: %s",
			nodegroup.NodeGroupID, nodegroup.MaxSize, nodegroup.MinSize, nodegroup.DesiredSize,
			nodegroup.UpcomingSize, nodegroup.Status, nodegroup.LastStatus)
	}
	return nodeGroups, nil
}

// handleElasticNodeGroupScaleUp check
func (c *control) handleElasticNodeGroupScaleUp(strategy *storage.NodeGroupMgrStrategy, scaleUpNum int) error {
	nodegroups, err := c.listElasticNodeGroups(strategy.ElasticNodeGroups)
	if err != nil {
		return err
	}
	// simply balance scaleUp number for each nodegroup
	distribution := newSimpleBalancer(strategy.ElasticNodeGroups)
	nodegrps := distribution.distribute(scaleUpNum)
	// update each nodegroup desiredNum and necessary info to storage
	// for coming webhook requests. all nodegroup actions must be log in storage
	// for next tick to confirm progress.
	for _, ng := range nodegrps {
		nodegroup := nodegroups[ng.NodeGroupID]
		original := nodegroup.DesiredSize
		nodegroup.DesiredSize = original + ng.partition
		nodegroup.LastScaleUpTime = time.Now()
		nodegroup.UpdatedTime = time.Now()
		nodegroup.LastStatus = nodegroup.Status
		nodegroup.Status = storage.ScaleUpState
		nodegroup.Message = fmt.Sprintf("nodegroup %s try to scaleup %d nodes from %d to %d",
			ng.NodeGroupID, ng.partition, original, nodegroup.DesiredSize)
		blog.Infof("%s", nodegroup.Message)
		// ready to update NodeGroup information for cluster-autoscaler webhook request
		if _, err := c.opt.Storage.UpdateNodeGroup(nodegroup, &storage.UpdateOptions{}); err != nil {
			blog.Errorf("controller update nodegroup %s/%s scaleup state into storage failure, %s",
				nodegroup.ClusterID, nodegroup.NodeGroupID, err.Error())
			return err
		}
		// record NodeGroupAction for tracing
		action := generateNodeGroupAction(nodegroup, storage.ScaleUpState, original, ng.partition, nodegroup.DesiredSize)
		if err := c.opt.Storage.CreateNodeGroupAction(action, &storage.CreateOptions{}); err != nil {
			blog.Errorf("controller create nodegroup %s scaleUp action failed, info: %s",
				nodegroup.NodeGroupID, err.Error())
			return fmt.Errorf("create %s NodeGroupAction to storage failure", nodegroup.NodeGroupID)
		}
		// record NodeGroupEvent for manually tracing
		event := generateNodeGroupEvent(nodegroup, original, storage.ScaleUpState,
			"ScaleUp decision making", fmt.Sprintf("try to scaleup %d nodes", ng.partition))
		if err := c.opt.Storage.CreateNodeGroupEvent(event, &storage.CreateOptions{}); err != nil {
			// event only used for administrator tracing issue manually.
			// failure of event operation is tolerable.
			blog.Errorf("controller create nodegroup %s scaleUp tracing event record failure, info: %s."+
				"failure is tolerable, controller try best effort for next event record",
				nodegroup.NodeGroupID, err.Error())
			continue
		}
		blog.Infof("nodegroup %s scaleUp information record in storage completely", nodegroup.NodeGroupID)
	}
	return nil
}

// handleElasticNodeGroupScaleDown check elastic nodeGroups details information,
// then store scaleDown desired nodes to relative nodeGroups, waiting for cluster-autoscaler
// webhook requests. Controller don't create elastic NodeGroup information automatically,
// because controller knows nothing about nodeGroups' maxSize, minSize, desiredSize etc.
// Controller only try to scale down elastic nodeGroups after cluster-autoscaler making webhook requests.
func (c *control) handleElasticNodeGroupScaleDown(
	strategy *storage.NodeGroupMgrStrategy, scaleDownNum int) error {
	nodeGroups, err := c.listElasticNodeGroups(strategy.ElasticNodeGroups)
	if err != nil {
		return err
	}
	// handle scale down number for each nodegroup
	distribution := newWeightBalancer(strategy.ElasticNodeGroups, nodeGroups)
	nodegrps := distribution.distribute(scaleDownNum)
	// update each nodegroup desiredNum
	for _, ng := range nodegrps {
		nodegroup := nodeGroups[ng.NodeGroupID]
		original := nodegroup.DesiredSize
		nodegroup.DesiredSize = original - ng.partition
		if nodegroup.DesiredSize < 0 {
			nodegroup.DesiredSize = 0
		}
		nodegroup.LastScaleDownTime = time.Now()
		nodegroup.LastStatus = nodegroup.Status
		nodegroup.Status = storage.ScaleDownState
		nodegroup.Message = fmt.Sprintf("nodegroup %s try to scaledown %d node from %d to %d",
			ng.NodeGroupID, ng.partition, original, nodegroup.DesiredSize)
		blog.Infof("%s", nodegroup.Message)
		// ready to update NodeGroup information for cluster-autoscaler webhook request
		if _, err := c.opt.Storage.UpdateNodeGroup(nodegroup,
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}); err != nil {
			blog.Errorf("controller update nodegroup %s/%s scaledDown state into storage failed, %s",
				nodegroup.ClusterID, nodegroup.NodeGroupID, err.Error())
			return err
		}
		// record NodeGroupAction for tracing
		action := generateNodeGroupAction(nodegroup, storage.ScaleDownState,
			original, ng.partition, nodegroup.DesiredSize)
		if err := c.opt.Storage.CreateNodeGroupAction(action, &storage.CreateOptions{}); err != nil {
			blog.Errorf("controller create nodegroup %s scale down action failed, info: %s. details: %v",
				nodegroup.NodeGroupID, err.Error(), action)
			return fmt.Errorf("create %s NodeGroupAction to storage failure", nodegroup.NodeGroupID)
		}
		// record NodeGroupEvent for manually tracing
		event := generateNodeGroupEvent(nodegroup, original, storage.ScaleDownState,
			"ScaleDown decision making", fmt.Sprintf("try to scaledown %d nodes", ng.partition))
		if err := c.opt.Storage.CreateNodeGroupEvent(event, &storage.CreateOptions{}); err != nil {
			// event only used for administrator tracing issue manually.
			// failure of event operation is tolerable.
			blog.Errorf("controller create nodegroup %s scaleDown tracing event record failed, info: %s."+
				"failure is tolerable, controller try best effort to next.",
				nodegroup.NodeGroupID, err.Error())
			continue
		}
		blog.Infof("nodegroup %s scaleDown information record in storage completely, action detail: %v",
			nodegroup.NodeGroupID, action)
	}
	return nil
}

// isElasticNodeGroupEssentialForScaleDown check if resources in pool is not enough for ReservedNodeGroup,
// then controller will scaleDown ElasticNodeGroup for releasing resource
func isElasticNodeGroupEssentialForScaleDown(strategy *storage.NodeGroupMgrStrategy,
	pool *storage.ResourcePool) (int, bool) {
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	warnBuffer := float64(strategy.Strategy.Buffer.Low)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, resourcePool:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum >= reservedNum {
		// resource is enough, do nothing
		blog.Infof("ResourcePool %s resource is idle %d >= reserved %d, elasticNodeGroup don't scaleDown",
			pool.ID, idleNum, reservedNum)
		return 0, false
	}
	// buffer resource is not enough, calculate necessary number for scale down
	scaleDownNum := reservedNum - idleNum
	blog.Infof("strategy:%s, scaleDownNum:%d", strategy.Name, scaleDownNum)
	return scaleDownNum, true
}

// isResourcePoolIdleForScaleUp check idle resources in pool, if resource is idle enough,
//  then controller will scaleUp desiredNodes in elastic nodegroup for resource consume
func isResourcePoolIdleForScaleUp(strategy *storage.NodeGroupMgrStrategy, pool *storage.ResourcePool) (int, bool) {
	total := float64(pool.InitNum + pool.IdleNum + pool.ConsumedNum + pool.ReturnedNum)
	idleNum := pool.IdleNum + pool.InitNum + pool.ReturnedNum
	warnBuffer := float64(strategy.Strategy.Buffer.High)
	reservedNum := int(math.Ceil(total * warnBuffer / 100))
	blog.Infof("strategy %s, resourcePool:%s: total:%d, idleNum:%d, warnBuffer:%d, reservedNum:%d",
		strategy.Name, pool.ID, int(total), idleNum, int(warnBuffer), reservedNum)
	if idleNum <= reservedNum {
		// resource is not idle enough
		blog.Infof("ResourcePool %s idle resource %d <= reserved %d, elasticNodeGroup don't scaleUp",
			pool.ID, idleNum, reservedNum)
		return 0, false
	}
	// check resource pool is idle and stable
	now := time.Now()
	diff := now.Sub(pool.UpdatedTime)
	if diff.Seconds() < float64(strategy.Strategy.MaxIdleDelay*60) {
		blog.Infof("ResourcePool %s is not stable enough for elasticNodeGroup scaleUp, now: %.f, target: %d",
			pool.ID, diff.Seconds(), strategy.Strategy.MaxIdleDelay*60)
		return 0, false
	}
	// resource is more than expected, check if controller can scale up
	scaleUpNum := idleNum - reservedNum
	if scaleUpNum < strategy.Strategy.MinScaleUpSize {
		blog.Infof("ResourcePool %s idle resource %d is less than MinScaleUpSize %d",
			pool.ID, scaleUpNum, strategy.Strategy.MinScaleUpSize)
		return 0, false
	}
	blog.Infof("strategy %s scaleUpNum:%d", strategy.Name, scaleUpNum)
	// feature(DeveloperJim): try to check ScaleUpCoolDown
	// if diff.Seconds() < strategy.Strategy.ScaleUpCoolDown do nothing
	return scaleUpNum, true
}

func generateNodeGroupAction(nodegroup *storage.NodeGroup,
	state string, original, deltaNum, newDesired int) *storage.NodeGroupAction {
	return &storage.NodeGroupAction{
		NodeGroupID:        nodegroup.NodeGroupID,
		ClusterID:          nodegroup.ClusterID,
		Event:              state,
		CreatedTime:        time.Now(),
		DeltaNum:           deltaNum,
		NewDesiredNum:      newDesired,
		OriginalDesiredNum: original,
		OriginalNodeNum:    len(nodegroup.NodeIPs),
		NodeIPs:            nodegroup.NodeIPs,
		Process:            0,
		Status:             storage.InitState,
		UpdatedTime:        time.Now(),
	}
}

func generateNodeGroupEvent(nodegroup *storage.NodeGroup, original int,
	event, reason, msg string) *storage.NodeGroupEvent {
	return &storage.NodeGroupEvent{
		NodeGroupID:        nodegroup.NodeGroupID,
		ClusterID:          nodegroup.ClusterID,
		EventTime:          time.Now(),
		Event:              event,
		MaxNum:             nodegroup.MaxSize,
		MinNum:             nodegroup.MinSize,
		DesiredNum:         nodegroup.DesiredSize,
		OriginalDesiredNum: original,
		OriginalNodeNum:    len(nodegroup.NodeIPs),
		Reason:             reason,
		Message:            msg,
	}
}

// upComingElasticResources check number that exist nodegroup actions can consume/release resources.
// all actions have been filtered that they are in same event.
// all nodeGroups were guaranteed that exist by listElasticNodeGroups
func upComingElasticResources(actions map[string]*storage.NodeGroupAction,
	nodeGroups map[string]*storage.NodeGroup, event string, strategy *storage.NodeGroupMgrStrategy) int {
	upComing := 0
	for nodegrpID, action := range actions {
		// check nodegroup action is still working on,
		// if action is timeout, there is no upcoming resources
		now := time.Now()
		gap := now.Sub(action.UpdatedTime)
		// !listElasticNodeGroups guarantee nodegroup exist
		nodeGroup := nodeGroups[nodegrpID]
		if event == storage.ScaleUpState {
			// calculate upComing scaleup resources to for elasticNodeGroup
			if gap.Seconds() >= float64(strategy.Strategy.ScaleUpDelay*60) {
				blog.Warnf("nodegroup %s action %s is timeout, lastUpdated %s, no upcoming resources",
					action.NodeGroupID, action.Event, action.UpdatedTime.Format(time.RFC3339))
				continue
			}
			// realNodes := len(nodeGroup.NodeIPs)
			realNodes := nodeGroup.CmDesiredSize
			if nodeGroup.DesiredSize <= realNodes {
				blog.Errorf("Nodegroup %s DesiredSize %d <= RealNodes %d, lastStatus maybe ScaleDown, "+
					"no upcoming resources.", nodeGroup.NodeGroupID, nodeGroup.DesiredSize, realNodes)
				continue
			}
			nodegroupComing := nodeGroup.DesiredSize - realNodes
			upComing += nodegroupComing
			blog.Infof("ScaleUp NodeGroup %s/%s upComing elastic resources %d",
				nodeGroup.ClusterID, nodeGroup.NodeGroupID, nodegroupComing)
			continue
		}
		if event == storage.ScaleDownState {
			// calculate upComing scaleDown resource to resourcePool
			if gap.Seconds() >= float64(strategy.Strategy.ScaleDownDelay*60) {
				blog.Warnf("nodegroup %s action %s is timeout, lastUpdated %s, no releasing resources",
					action.NodeGroupID, action.Event, action.UpdatedTime.Format(time.RFC3339))
				continue
			}
			// realNodes := len(nodeGroup.NodeIPs)
			realNodes := nodeGroup.CmDesiredSize
			if nodeGroup.DesiredSize >= realNodes {
				blog.Errorf("Nodegroup %s DesiredSize %d >= RealNodes %d, lastStatus maybe ScaleUp, "+
					"no releasing resources.", nodeGroup.NodeGroupID, nodeGroup.DesiredSize, realNodes)
				continue
			}
			nodegroupComing := realNodes - nodeGroup.DesiredSize
			upComing += nodegroupComing
			blog.Infof("ScaleDown NodeGroup %s/%s release elastic resources %d",
				nodeGroup.ClusterID, nodeGroup.NodeGroupID, nodegroupComing)
			continue
		}
	}
	// if no available action, then upcoming is 0
	return upComing
}
