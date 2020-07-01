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

// transaction for update application

package scheduler

import (
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// The goroutine function for update application transaction
// You can create a transaction for update application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunUpdateApplication(transaction *Transaction) {

	runAs := transaction.RunAs
	appID := transaction.AppID

	blog.Infof("transaction %s update(%s.%s) run begin", transaction.ID, runAs, appID)
	//var offerIdx int64 = 0
	startedTaskgroup := time.Now()
	startedApp := time.Now()
	for {
		blog.Infof("transaction %s update(%s.%s) run check", transaction.ID, runAs, appID)

		//check begin
		if transaction.CreateTime+transaction.DelayTime > time.Now().Unix() {
			blog.V(3).Infof("transaction %s update(%s.%s) delaytime(%d), cannot do at now",
				transaction.ID, runAs, appID, transaction.DelayTime)
			time.Sleep(3 * time.Second)
			continue
		}

		//check doing
		opData := transaction.OpData.(*TransAPIUpdateOpdata)
		version := opData.Version
		offerOut := s.GetFirstOffer()

		taskGroupID := opData.Taskgroups[opData.LaunchedNum].ID
		for offerOut != nil {
			offerIdx := offerOut.Id
			offer := offerOut.Offer

			curOffer := offerOut
			offerOut = s.GetNextOffer(offerOut)
			blog.V(3).Infof("transaction %s get offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

			isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) && s.IsConstraintsFit(version, offer, taskGroupID) &&
				s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
			if isFit == true {
				blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				if s.UseOffer(curOffer) == true {
					blog.Infof("transaction %s update(%s.%s) use offer(%d) %s||%s", transaction.ID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					launchedNum := opData.LaunchedNum
					s.doUpdateTrans(transaction, curOffer, startedTaskgroup)
					if transaction.Status == types.OPERATION_STATUS_FINISH || transaction.Status == types.OPERATION_STATUS_FAIL {
						blog.Infof("transaction %s update(%s.%s) finish", transaction.ID, runAs, appID)
						goto run_end
					}
					if launchedNum < opData.LaunchedNum {
						startedTaskgroup = time.Now()
					}
				} else {
					blog.Infof("transaction %s use offer(%d) %s||%s fail", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				}
			}
		}

		//check timeout
		if (transaction.CreateTime + transaction.LifePeriod) < time.Now().Unix() {
			blog.Warn("transaction %s update(%s.%s) timeout", transaction.ID, runAs, appID)
			transaction.Status = types.OPERATION_STATUS_TIMEOUT
			goto run_end
		}

		time.Sleep(time.Second)
	}

run_end:
	s.FinishTransaction(transaction)
	reportOperateAppMetrics(transaction.RunAs, transaction.AppID, UpdateApplicationType, startedApp)
	blog.Infof("transaction %s update(%s.%s) run end, result(%s)", transaction.ID, runAs, appID, transaction.Status)

}

func (s *Scheduler) doUpdateTrans(trans *Transaction, outOffer *offer.Offer, started time.Time) {

	offer := outOffer.Offer

	runAs := trans.RunAs
	appID := trans.AppID

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil || app.Created > trans.CreateTime {
		blog.Error("transaction %s, application(%s.%s) not exist", trans.ID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	cpus, mem, disk := s.OfferedResources(offer)

	opData := trans.OpData.(*TransAPIUpdateOpdata)
	version := opData.Version
	resources := task.BuildResources(version.AllResource())
	blog.V(3).Infof("transaction %s: to update instances num(%d), launched(%d)",
		trans.ID, opData.Instances, opData.LaunchedNum)
	blog.Info("transaction %s update application(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.ID, runAs, appID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	if opData.LaunchedNum == opData.Instances {
		blog.Warn("transaction %s update application(%s.%s), but all taskgroup already done", trans.ID, runAs, appID)

		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		err = s.store.SaveApplication(app)
		if err != nil {
			blog.Error("transaction %s save application(%s.%s) err:%s", trans.ID, app.RunAs, app.ID, err.Error())
			s.DeclineResource(offer.Id.Value)
			return
		}

		blog.V(3).Infof("transaction %s save application(%s.%s) succ!", trans.ID, app.RunAs, app.ID)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		return
	}

	var taskGroupInfos []*mesos.TaskGroupInfo
	var oldTaskGroups []*types.TaskGroup

	var taskGroupID string
	var taskgroupName string
	//if opData.LaunchedNum < opData.Instances && version.IsResourceFit(types.Resource{Cpus: cpus, Mem: mem, Disk: disk}) {
	if opData.LaunchedNum < opData.Instances && s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) &&
		s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		taskGroupID = opData.Taskgroups[opData.LaunchedNum].ID
		blog.Info("transaction %s get taskgroup(%s) to do update", trans.ID, taskGroupID)
		var taskGroup *types.TaskGroup
		taskGroup, err = s.store.FetchTaskGroup(taskGroupID)
		if taskGroup == nil {
			blog.Error("transaction %s fetch taskgroup(%s) err(%s)", trans.ID, taskGroupID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}
		taskgroupName = taskGroup.Name

		opData.Taskgroups[opData.LaunchedNum] = taskGroup

		// check old taskGroup end
		isEnd := task.IsTaskGroupEnd(taskGroup)
		if isEnd == false {
			blog.Info("transaction %s update application(%s.%s) pending", trans.ID, runAs, appID)
			if task.CanTaskGroupShutdown(taskGroup) {
				blog.Info("transaction %s pending: kill old taskGroup(%s)", trans.ID, taskGroup.ID)
				s.KillTaskGroup(taskGroup)
			}
			s.DeclineResource(offer.Id.Value)
			return
		}
		blog.Info("transaction %s update application, old taskGroup(%s) is end, continue ...", trans.ID, taskGroup.ID)

		newTaskGroupID, _ := task.ReBuildTaskGroupID(taskGroup.ID)
		var newTaskGroup *types.TaskGroup
		newTaskGroup, err = s.BuildTaskGroup(version, app, newTaskGroupID, "update application")
		if err != nil {
			blog.Error("transaction %s Build taskgroup(%s) failed: %s", trans.ID, newTaskGroupID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}

		newTaskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, newTaskGroup)
		if newTaskGroupInfo == nil {
			blog.Error("transaction %s build taskgroupinfo(%s) failed", trans.ID, newTaskGroup.ID)
			s.DeleteTaskGroup(app, newTaskGroup, "create taskgroupinfo fail")
			//trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}

		if err = s.store.SaveTaskGroup(newTaskGroup); err != nil {
			blog.Error("transaction %s save taskgroup(%s) error %s", trans.ID, newTaskGroup.ID, err.Error())
			s.DeleteTaskGroup(app, newTaskGroup, "save taskgroup fail")
			//trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}

		opData.LaunchedNum++
		taskGroupInfos = append(taskGroupInfos, newTaskGroupInfo)
		oldTaskGroups = append(oldTaskGroups, taskGroup)

		//lock agentsetting
		util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, newTaskGroup.GetAgentIp())
		//update agentsettings taskgroup index info
		agentsetting, _ := s.store.FetchAgentSetting(newTaskGroup.GetAgentIp())
		if agentsetting != nil {
			agentsetting.Pods = append(agentsetting.Pods, newTaskGroup.ID)
			err := s.store.SaveAgentSetting(agentsetting)
			if err != nil {
				blog.Errorf("save agentsetting %s pods error %s", agentsetting.InnerIP, err.Error())
			}
		} else {
			blog.Errorf("fetch agentsetting %s Not Found", newTaskGroup.GetAgentIp())
		}
		util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, newTaskGroup.GetAgentIp())
	}

	if len(taskGroupInfos) <= 0 {
		blog.Error("transaction %s has no taskgroup to launch", trans.ID)
		s.DeclineResource(offer.Id.Value)
		trans.Status = types.OPERATION_STATUS_FAIL
		return
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s launch taskgroups fail: %s", trans.ID, err.Error())
		opData.LaunchedNum = opData.LaunchedNum - len(taskGroupInfos)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s launch taskgroup resp status err code : %d", trans.ID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	for _, taskGroup := range oldTaskGroups {
		if err := s.DeleteTaskGroup(app, taskGroup, "update application"); err != nil {
			blog.Error("transaction %s delete taskgroup %s failed: %s", trans.ID, taskGroup.ID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			return
		}
	}

	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, UpdateTaskgroupType, started)
	if opData.LaunchedNum == opData.Instances {
		blog.Info("transaction %s update application(%s.%s), all taskgroup already done", trans.ID, app.RunAs, app.ID)
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		trans.Status = types.OPERATION_STATUS_FINISH
	}
	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s save application(%s.%s) err:%s", trans.ID, app.RunAs, app.ID, err.Error())
	}

	blog.V(3).Infof("do update transaction %s end", trans.ID)
	return
}
