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

package scheduler

import (
	"fmt"
	"net/http"
	"time"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/sched"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/strategy"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"

	"github.com/golang/protobuf/proto"
)

// BuildTaskGroup Build an taskgroup for application:
// If ID is empty, the taskgroup's ID will created and its index will be app.Instances,
// If ID is not empty, the taskgroup's ID will be inputted ID
// You can input the reason to decribe why the taskgrop is built.
// The taskgroup will be created in DB, application, and also will be outputted in related service
func (s *Scheduler) BuildTaskGroup(version *types.Version, app *types.Application, id string, reason string) (
	*types.TaskGroup, error) {

	taskgroup, err := task.CreateTaskGroup(version, id, app.Instances, app.ClusterId, reason, s.store)
	if taskgroup == nil {
		blog.Errorf("create taskgroup err: %s", err.Error())
		return nil, err
	}

	err = s.store.SaveTaskGroup(taskgroup)
	if err != nil {
		blog.Error("save taskgroup(%s) err: %s", taskgroup.ID, err.Error())
		return nil, err
	}
	blog.Infof("build taskgroup %s for %s", taskgroup.ID, reason)
	s.ServiceMgr.TaskgroupAdd(taskgroup)
	podIndex := new(bcstype.BcsPodIndex)
	podIndex.Name = taskgroup.ID
	app.Pods = append(app.Pods, podIndex)
	app.UpdateTime = time.Now().Unix()
	if id == "" {
		app.Instances = uint64(len(app.Pods))
	}

	// add events
	for _, task := range taskgroup.Taskgroup {
		s.produceEvent(*task)
	}

	return taskgroup, nil
}

// LaunchTaskGroup Launch an taskgroup with offered slave resource
func (s *Scheduler) LaunchTaskGroup(offer *mesos.Offer, taskGroup *mesos.TaskGroupInfo,
	version *types.Version) (*http.Response, error) {

	blog.Infof("launch %d tasks with offer %s", len(taskGroup.Tasks), *offer.GetId().Value)

	executor, err := task.CreateBcsExecutorInfo(offer, task.GetTaskGroupID(taskGroup), version, s.store)
	if err != nil {
		return nil, fmt.Errorf("create executor from version(%s.%s.%s) failed, %s",
			version.RunAs, version.ID, version.Name, err.Error())
	}

	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesos.OfferID{
				offer.GetId(),
			},
			Operations: []*mesos.Offer_Operation{
				{
					Type: mesos.Offer_Operation_LAUNCH_GROUP.Enum(),
					LaunchGroup: &mesos.Offer_Operation_LaunchGroup{
						TaskGroup: taskGroup,
						Executor:  executor,
					},
				},
			},
			Filters: &mesos.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	return s.send(call)
}

// LaunchTaskGroups Launch taskgroups with offered slave resource
func (s *Scheduler) LaunchTaskGroups(offer *mesos.Offer, taskGroups []*mesos.TaskGroupInfo,
	version *types.Version) (*http.Response, error) {

	blog.Infof("launch %d taskgroup with offer %s", len(taskGroups), *offer.GetId().Value)

	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds: []*mesos.OfferID{
				offer.GetId(),
			},
			Operations: []*mesos.Offer_Operation{},
			Filters:    &mesos.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	for _, taskGroup := range taskGroups {

		executor, err := task.CreateBcsExecutorInfo(offer, task.GetTaskGroupID(taskGroup), version, s.store)
		if err != nil {
			return nil, fmt.Errorf("create executor from version(%s.%s.%s) failed, %s",
				version.RunAs, version.ID, version.Name, err.Error())
		}

		ops := &mesos.Offer_Operation{
			Type: mesos.Offer_Operation_LAUNCH_GROUP.Enum(),
			LaunchGroup: &mesos.Offer_Operation_LaunchGroup{
				TaskGroup: taskGroup,
				Executor:  executor,
			},
		}
		call.Accept.Operations = append(call.Accept.Operations, ops)
	}

	return s.send(call)
}

// KillTaskGroup Kill a taskgroup with taskgroup information
func (s *Scheduler) KillTaskGroup(taskGroup *types.TaskGroup) (*http.Response, error) {
	blog.Info("kill taskgroup(%s) on ExecutorID(%s), AgentID(%s)",
		taskGroup.ID, taskGroup.ExecutorID, taskGroup.AgentID)
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_SHUTDOWN.Enum(),
		Shutdown: &sched.Call_Shutdown{
			ExecutorId: &mesos.ExecutorID{
				Value: &taskGroup.ExecutorID,
			},
			AgentId: &mesos.AgentID{
				Value: &taskGroup.AgentID,
			},
		},
	}

	return s.send(call)
}

// KillExecutor Kill a taskgroup with the agent and executor ID
func (s *Scheduler) KillExecutor(agentID, executerID string) (*http.Response, error) {
	blog.Info("kill taskgroup on AgentID(%s), ExecutorID(%s)", agentID, executerID)
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_SHUTDOWN.Enum(),
		Shutdown: &sched.Call_Shutdown{
			ExecutorId: &mesos.ExecutorID{
				Value: &executerID,
			},
			AgentId: &mesos.AgentID{
				Value: &agentID,
			},
		},
	}

	return s.send(call)
}

// DeleteTaskGroup Delete a taskgroup:
// the taskgroup will delete from DB, application and service
func (s *Scheduler) DeleteTaskGroup(app *types.Application, taskGroup *types.TaskGroup, reason string) error {
	blog.Info("delete taskgroup %s for %s", taskGroup.ID, reason)
	s.ServiceMgr.TaskgroupDelete(taskGroup)
	// update app taskgroup index info
	if app != nil {
		delete := -1
		for index, currPod := range app.Pods {
			if currPod.Name == taskGroup.ID {
				delete = index
				break
			}
		}
		if delete != -1 {
			app.UpdateTime = time.Now().Unix()
			app.Pods = append(app.Pods[:delete], app.Pods[delete+1:]...)
		}
	}

	return s.deleteTaskGroup(taskGroup)
}

// Delete a taskgroup:
// the taskgroup will delete from DB
func (s *Scheduler) deleteTaskGroup(taskGroup *types.TaskGroup) error {
	blog.Infof("delete taskgroup(%s) in store", taskGroup.ID)
	// release taskgroup's DeltaCPU, DeltaDisk, DeltaMem
	if err := s.UpdateAgentSchedInfo(taskGroup.HostName, taskGroup.ID, nil); err != nil {
		blog.Errorf("update agent sched info %s failed when delete taskgroup %s", taskGroup.HostName, taskGroup.ID)
	}

	// update agentsetting taskgroup index info
	nodeIP := taskGroup.GetAgentIp()
	if nodeIP == "" {
		blog.Errorf("taskgroup %s don't have nodeIP", taskGroup.ID)
	} else {
		// lock agentsetting
		util.Lock.Lock(bcstype.BcsClusterAgentSetting{}, nodeIP)
		defer util.Lock.UnLock(bcstype.BcsClusterAgentSetting{}, nodeIP)

		agentsetting, err := s.store.FetchAgentSetting(nodeIP)
		if err != nil && !errors.Is(err, store.ErrNoFound) {
			blog.Errorf("fetch agentsetting %s failed: %s", nodeIP, err.Error())
			return fmt.Errorf("fetch agentsetting %s failed: %s", nodeIP, err.Error())
		}
		if agentsetting == nil {
			blog.Errorf("fetch agentsetting %s Not found", nodeIP)
		} else {
			delete := -1
			for index, currPod := range agentsetting.Pods {
				if currPod == taskGroup.ID {
					delete = index
					break
				}
			}
			if delete != -1 {
				agentsetting.Pods = append(agentsetting.Pods[:delete], agentsetting.Pods[delete+1:]...)
				// TODO: to deal with save agent setting error
				// save operation may failed when multiple operations are on same agent setting
				err = s.store.SaveAgentSetting(agentsetting)
				if err != nil {
					blog.Errorf("save agentsetting %s failed when delete taskgroup, err %s", nodeIP, err.Error())
				}
			}
		}
	}
	// delete taskgroup from store
	if err := s.store.DeleteTaskGroup(taskGroup.ID); err != nil {
		blog.Errorf("delete taskgroup(%s) err: %s", taskGroup.ID, err.Error())
		return fmt.Errorf("delete taskgroup(%s) err: %s", taskGroup.ID, err.Error())
	}
	return nil
}

// IsOfferResourceFitLaunch Check whether the offer is match required resource for launching a taskgroup
func (s *Scheduler) IsOfferResourceFitLaunch(needResource *types.Resource, outOffer *offer.Offer) bool {

	inOffer := outOffer.Offer
	cpus, mem, disk := s.OfferedResources(inOffer)

	if outOffer.DeltaCPU > 0 {
		blog.V(3).Infof("offer %s CPU(offer: %f, delta: %f)",
			*(inOffer.Id.Value), cpus, outOffer.DeltaCPU)
		cpus = cpus - outOffer.DeltaCPU
	}
	if outOffer.DeltaMem > 0 {
		blog.V(3).Infof("offer %s MEM(offer: %f, delta: %f)",
			*(inOffer.Id.Value), mem, outOffer.DeltaMem)
		mem = mem - outOffer.DeltaMem
	}
	if outOffer.DeltaDisk > 0 {
		blog.V(3).Infof("offer %s DISK(offer: %f, delta: %f)",
			*(inOffer.Id.Value), disk, outOffer.DeltaDisk)
		disk = disk - outOffer.DeltaDisk
	}

	if needResource.Cpus <= cpus && needResource.Mem <= mem && needResource.Disk <= disk {
		return true
	}

	blog.V(3).Infof("offer %s resource not enough: need(%f, %f, %f), offer(%f, %f, %f)",
		*(inOffer.Id.Value), needResource.Cpus, needResource.Mem, needResource.Disk, cpus, mem, disk)

	return false
}

// IsOfferExtendedResourcesFitLaunch check whether the offer is match extended resources for launching a taskgroup
func (s *Scheduler) IsOfferExtendedResourcesFitLaunch(
	needs map[string]*bcstype.ExtendedResource, outOffer *offer.Offer) bool {
	// if version don't have extended resources, then return true
	if needs == nil || len(needs) == 0 {
		return true
	}

	for _, need := range needs {
		resource := s.getNeedResourceOfOffer(outOffer.Offer, need.Name)
		if resource == nil {
			blog.V(3).Infof("offer %s don't have extended resources %s", outOffer.Offer.GetHostname(), need.Name)
			return false
		}
		// if offer extended resources not enough, then return false
		if need.Value > resource.GetScalar().GetValue() {
			blog.V(3).Infof("offer %s extended resources %s not enough: need(%f), offer(%f)",
				outOffer.Offer.GetHostname(), need.Name, need.Value, resource.GetScalar().GetValue())
			return false
		}
	}
	// if offer extended resources match fit, then return true
	return true
}

func (s *Scheduler) getNeedResourceOfOffer(o *mesos.Offer, name string) *mesos.Resource {
	for _, re := range o.GetResources() {
		if re.GetName() == name {
			return re
		}
	}

	return nil
}

// IsOfferResourceFitUpdate Check whether the offer is match required resource for updating a taskgroup's resource
func (s *Scheduler) IsOfferResourceFitUpdate(needResource *types.Resource, outOffer *offer.Offer) bool {

	inOffer := outOffer.Offer
	cpus, mem, disk := s.OfferedResources(inOffer)

	cpus = cpus - outOffer.DeltaCPU
	mem = mem - outOffer.DeltaMem
	disk = disk - outOffer.DeltaDisk

	if needResource.Cpus <= cpus && needResource.Mem <= mem && needResource.Disk <= disk {
		return true
	}

	blog.V(3).Infof("offer %s resource not enough: need(%f, %f, %f), offer(-delta)(%f, %f, %f)",
		*(inOffer.Id.Value), needResource.Cpus, needResource.Mem, needResource.Disk, cpus, mem, disk)

	return false
}

// IsConstraintsFit Check whether the offer match version constraints
func (s *Scheduler) IsConstraintsFit(version *types.Version, offer *mesos.Offer, taskgroupID string) bool {

	isFit, _ := strategy.ConstraintsFit(version, offer, s.store, taskgroupID)
	return isFit
}
