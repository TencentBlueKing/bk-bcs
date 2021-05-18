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
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// RunUpdateApplication The goroutine function for update application transaction
// You can create a transaction for update application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunUpdateApplication(transaction *types.Transaction) bool {

	runAs := transaction.Namespace
	appID := transaction.ObjectName

	startedTaskgroup := time.Now()
	blog.Infof("transaction %s update(%s.%s) run check", transaction.TransactionID, runAs, appID)

	// check begin
	if transaction.CreateTime.Add(transaction.DelayTime).After(time.Now()) {
		blog.V(3).Infof("transaction %s update(%s.%s) delaytime(%s), cannot do at now",
			transaction.TransactionID, runAs, appID, transaction.DelayTime.String())
		time.Sleep(3 * time.Second)
		return true
	}

	//check doing
	opData := transaction.CurOp.OpUpdateData
	version := opData.Version
	offerOut := s.GetFirstOffer()

	taskGroupID := opData.Taskgroups[opData.LaunchedNum].ID
	for offerOut != nil {
		offerIdx := offerOut.Id
		offer := offerOut.Offer

		curOffer := offerOut
		offerOut = s.GetNextOffer(offerOut)
		blog.V(3).Infof("transaction %s get offer(%d) %s||%s ",
			transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

		isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) &&
			s.IsConstraintsFit(version, offer, taskGroupID) &&
			s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
		if isFit == true {
			blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ",
				transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
			if s.UseOffer(curOffer) == true {
				blog.Infof("transaction %s update(%s.%s) use offer(%d) %s||%s",
					transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				launchedNum := opData.LaunchedNum
				isContinue := s.doUpdateTrans(transaction, curOffer, startedTaskgroup)
				if !isContinue {
					blog.Infof("transaction %s update(%s.%s) finish", transaction.TransactionID, runAs, appID)
					return false
				}
				if launchedNum < opData.LaunchedNum {
					startedTaskgroup = time.Now()
				}
			} else {
				blog.Infof("transaction %s use offer(%d) %s||%s fail",
					transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
			}
		}
	}
	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doUpdateTrans(trans *types.Transaction, outOffer *offer.Offer, started time.Time) bool {

	offer := outOffer.Offer

	runAs := trans.Namespace
	appID := trans.ObjectName

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil || app.Created > trans.CreateTime.Unix() {
		blog.Error("transaction %s, application(%s.%s) not exist", trans.TransactionID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "fetch appliction failed"
		s.DeclineResource(offer.Id.Value)
		return false
	}

	cpus, mem, disk := s.OfferedResources(offer)

	opData := trans.CurOp.OpUpdateData
	version := opData.Version
	resources := task.BuildResources(version.AllResource())
	blog.V(3).Infof("transaction %s: to update instances num(%d), launched(%d)",
		trans.TransactionID, opData.Instances, opData.LaunchedNum)
	blog.Info("transaction %s update application(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.TransactionID, runAs, appID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	if opData.LaunchedNum == opData.Instances {
		blog.Warn("transaction %s update application(%s.%s), but all taskgroup already done",
			trans.TransactionID, runAs, appID)

		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		err = s.store.SaveApplication(app)
		if err != nil {
			blog.Error("transaction %s save application(%s.%s) err:%s",
				trans.TransactionID, app.RunAs, app.ID, err.Error())
			s.DeclineResource(offer.Id.Value)
			return true
		}

		blog.V(3).Infof("transaction %s save application(%s.%s) succ!",
			trans.TransactionID, app.RunAs, app.ID)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		return false
	}

	var taskGroupInfos []*mesos.TaskGroupInfo
	var oldTaskGroups []*types.TaskGroup

	var taskGroupID string
	var taskgroupName string
	if opData.LaunchedNum < opData.Instances && s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) &&
		s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		taskGroupID = opData.Taskgroups[opData.LaunchedNum].ID
		blog.Info("transaction %s get taskgroup(%s) to do update", trans.TransactionID, taskGroupID)
		var taskGroup *types.TaskGroup
		taskGroup, err = s.store.FetchTaskGroup(taskGroupID)
		if taskGroup == nil {
			blog.Error("transaction %s fetch taskgroup(%s) err(%s)", trans.TransactionID, taskGroupID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			trans.Message = fmt.Sprintf("transaction %s fetch taskgroup(%s) err(%s)",
				trans.TransactionID, taskGroupID, err.Error())
			s.DeclineResource(offer.Id.Value)
			return false
		}
		taskgroupName = taskGroup.Name

		opData.Taskgroups[opData.LaunchedNum] = taskGroup

		// check old taskGroup end
		isEnd := task.IsTaskGroupEnd(taskGroup)
		if isEnd == false {
			blog.Info("transaction %s update application(%s.%s) pending", trans.TransactionID, runAs, appID)
			if task.CanTaskGroupShutdown(taskGroup) {
				blog.Info("transaction %s pending: kill old taskGroup(%s)", trans.TransactionID, taskGroup.ID)
				s.KillTaskGroup(taskGroup)
			}
			s.DeclineResource(offer.Id.Value)
			return true
		}
		blog.Info("transaction %s update application, old taskGroup(%s) is end, continue ...",
			trans.TransactionID, taskGroup.ID)

		newTaskGroupID, _ := task.ReBuildTaskGroupID(taskGroup.ID)
		var newTaskGroup *types.TaskGroup
		newTaskGroup, err = s.BuildTaskGroup(version, app, newTaskGroupID, "update application")
		if err != nil {
			blog.Error("transaction %s Build taskgroup(%s) failed: %s",
				trans.TransactionID, newTaskGroupID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			trans.Message = fmt.Sprintf("transaction %s Build taskgroup(%s) failed: %s",
				trans.TransactionID, newTaskGroupID, err.Error())
			s.DeclineResource(offer.Id.Value)
			return false
		}

		newTaskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, newTaskGroup)
		if newTaskGroupInfo == nil {
			blog.Error("transaction %s build taskgroupinfo(%s) failed", trans.TransactionID, newTaskGroup.ID)
			s.DeleteTaskGroup(app, newTaskGroup, "create taskgroupinfo fail")
			s.DeclineResource(offer.Id.Value)
			return true
		}

		if err = s.store.SaveTaskGroup(newTaskGroup); err != nil {
			blog.Error("transaction %s save taskgroup(%s) error %s", trans.TransactionID, newTaskGroup.ID, err.Error())
			s.DeleteTaskGroup(app, newTaskGroup, "save taskgroup fail")
			s.DeclineResource(offer.Id.Value)
			return true
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
		blog.Error("transaction %s has no taskgroup to launch", trans.TransactionID)
		s.DeclineResource(offer.Id.Value)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "has no taskgroup to launch"
		return false
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s launch taskgroups fail: %s", trans.TransactionID, err.Error())
		opData.LaunchedNum = opData.LaunchedNum - len(taskGroupInfos)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("transaction %s launch taskgroups fail: %s", trans.TransactionID, err.Error())
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s launch taskgroup resp status err code : %d", trans.TransactionID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("transaction %s launch taskgroup resp status err code : %d",
			trans.TransactionID, resp.StatusCode)
		s.DeclineResource(offer.Id.Value)
		return false
	}

	for _, taskGroup := range oldTaskGroups {
		if err := s.DeleteTaskGroup(app, taskGroup, "update application"); err != nil {
			blog.Error("transaction %s delete taskgroup %s failed: %s", trans.TransactionID, taskGroup.ID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			trans.Message = fmt.Sprintf("transaction %s delete taskgroup %s failed: %s",
				trans.TransactionID, taskGroup.ID, err.Error())
			return false
		}
	}

	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, UpdateTaskgroupType, started)
	if opData.LaunchedNum == opData.Instances {
		blog.Info("transaction %s update application(%s.%s), all taskgroup already done",
			trans.TransactionID, app.RunAs, app.ID)
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		trans.Status = types.OPERATION_STATUS_FINISH
	}
	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s save application(%s.%s) err:%s", trans.TransactionID, app.RunAs, app.ID, err.Error())
	}

	blog.V(3).Infof("do update transaction %s, launched number %d", trans.TransactionID, opData.LaunchedNum)
	if opData.LaunchedNum == opData.Instances {
		return false
	}
	return true
}
