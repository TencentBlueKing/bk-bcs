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

// transaction for resched task group

package scheduler

import (
	"fmt"
	"net/http"
	"time"

	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// RunRescheduleTaskgroup The goroutine function for reschedule taskgroup transaction
// You can create a transaction for reschedule taskgroup, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunRescheduleTaskgroup(transaction *types.Transaction) bool {
	rescheduleOpdata := transaction.CurOp.OpRescheduleData
	taskGroupID := rescheduleOpdata.TaskGroupID

	transaction.CurOp.OpRescheduleData.SchedulerNum++
	blog.Infof("transaction %s reschedule(%s) run check", transaction.TransactionID, taskGroupID)

	// check begin
	if transaction.CreateTime.Add(transaction.DelayTime).After(time.Now()) {
		blog.V(3).Infof("transaction %s reschedule(%s) delaytime(%d), cannot do at now",
			transaction.TransactionID, taskGroupID, transaction.DelayTime)
		return true
	}

	// precheck application&taskgroup status
	taskgroup, _ := s.store.FetchTaskGroup(taskGroupID)
	if taskgroup == nil {
		blog.Infof("transaction %s fail: fetch taskgroup(%s) return nil", transaction.TransactionID, taskGroupID)
		transaction.Status = types.OPERATION_STATUS_FAIL
		return false
	}
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app, _ := s.store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Infof("transaction %s fail: fetch application(%s.%s) return nil",
			transaction.TransactionID, runAs, appID)
		transaction.Status = types.OPERATION_STATUS_FAIL
		return false
	}
	// check doing
	opData := transaction.CurOp.OpRescheduleData
	version := opData.Version

	// if inner rescheduler, then check app status
	if opData.IsInner {
		if app.Status == types.APP_STATUS_OPERATING {
			blog.Infof("transaction %s pending: app(%s.%s) is in status(%s:%s), cannot do at now",
				transaction.TransactionID, runAs, appID, app.Status, app.SubStatus)
			return true
		}
		if app.Status == types.APP_STATUS_ROLLINGUPDATE && app.SubStatus != types.APP_SUBSTATUS_ROLLINGUPDATE_UP {
			blog.Infof("transaction %s pending: app(%s.%s) is in status(%s:%s), cannot do at now",
				transaction.TransactionID, runAs, appID, app.Status, app.SubStatus)
			return true
		}
	}

	hostRetain := false
	if transaction.CreateTime.Unix()+opData.HostRetainTime > time.Now().Unix() {
		hostRetain = true
	}

	isContinue := true
	offerOut := s.GetFirstOffer()
	for offerOut != nil {
		offerIdx := offerOut.Id
		offer := offerOut.Offer
		blog.V(3).Infof("transaction %s get offer(%d) %s||%s ",
			transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

		curOffer := offerOut
		offerOut = s.GetNextOffer(offerOut)
		if hostRetain == false || offer.GetHostname() == opData.HostRetain {
			isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) &&
				s.IsConstraintsFit(version, offer, taskGroupID) &&
				s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
			if isFit == true {
				blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ",
					transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				if s.UseOffer(curOffer) == true {
					blog.Info("transaction %s reschedule(%s) use offer(%d) %s||%s",
						transaction.TransactionID, taskGroupID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					isContinue = s.doRescheduleTrans(transaction, curOffer)
					break
				} else {
					blog.Info("transaction %s use offer(%d) %s||%s fail",
						transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				}
			}
		}
	}

	// check end
	if !isContinue {
		blog.Infof("transaction %s reschedule(%s) finish", transaction.TransactionID, taskGroupID)
		return false
	}

	// when scheduler taskgroup number>=10, then report resources insufficient message
	if transaction.CurOp.OpRescheduleData.SchedulerNum == 10 {
		s.store.LockApplication(runAs + "." + appID)
		taskGroup, _ := s.store.FetchTaskGroup(taskGroupID)
		if taskGroup != nil {
			taskGroup.Message = "don't have fit resources to reschedule this taskgroup"
			s.store.SaveTaskGroup(taskGroup)
		}
		s.store.UnLockApplication(runAs + "." + appID)
	}
	return true
}

func (s *Scheduler) passRescheduleCheck(trans *types.Transaction, app *types.Application) bool {

	if app.Created > trans.CreateTime.Unix() {
		blog.Warn("transaction %s fail: application(%s.%s) Created(%d) > transaction.CreateTime(%d)",
			trans.TransactionID, app.RunAs, app.ID, app.Created, trans.CreateTime.Unix())
		trans.Status = types.OPERATION_STATUS_FAIL
		return false
	}

	opData := trans.CurOp.OpRescheduleData
	if !opData.IsInner {
		return true
	}

	if app.Status == types.APP_STATUS_OPERATING {
		blog.Warn("transaction %s pending: app(%s.%s) is in status(%s:%s), redo later",
			trans.TransactionID, app.RunAs, app.ID, app.Status, app.SubStatus)
		return false
	}
	if app.Status == types.APP_STATUS_ROLLINGUPDATE && app.SubStatus != types.APP_SUBSTATUS_ROLLINGUPDATE_UP {
		blog.Warn("transaction %s pending: app(%s.%s) is in status(%s:%s), redo later",
			trans.TransactionID, app.RunAs, app.ID, app.Status, app.SubStatus)
		return false
	}

	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doRescheduleTrans(trans *types.Transaction, outOffer *offer.Offer) bool {
	blog.Infof("do transaction %s begin", trans.TransactionID)

	offer := outOffer.Offer
	rescheduleOpdata := trans.CurOp.OpRescheduleData
	taskGroupID := rescheduleOpdata.TaskGroupID
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)

	cpus, mem, disk := s.OfferedResources(offer)
	blog.Info("transaction %s reschedule taskgroup(%s), offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.TransactionID, taskGroupID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Warn("transaction %s fail: fetch application(%s.%s) return nil", trans.TransactionID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}

	if !s.passRescheduleCheck(trans, app) {
		s.DeclineResource(offer.Id.Value)
		return true
	}

	version, _ := s.store.GetVersion(runAs, appID)
	if version == nil {
		blog.Error("transaction %s fail: no version for application(%s.%s)", trans.TransactionID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}

	taskGroup, err := s.store.FetchTaskGroup(taskGroupID)
	if taskGroup == nil {
		blog.Info("transaction %s fetch taskGroup(%s) fail: %s", trans.TransactionID, taskGroupID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}

	if rescheduleOpdata.IsInner && taskGroup.Status == types.TASKGROUP_STATUS_RUNNING {
		blog.Info("transaction %s finish: lost taskGroup(%s) recover to running",
			trans.TransactionID, taskGroupID)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		var alarmTimeval uint16 = 600
		s.SendHealthMsg(alarm.InfoKind, taskGroup.RunAs,
			"taskgroup recover to running, id:"+taskGroupID,
			taskGroup.RunAs+"."+taskGroup.Name+"-restart", &alarmTimeval)
		return false
	}

	// if old taskgroup cannot be rescheduled
	isEnd := task.CanTaskGroupReschedule(taskGroup)
	if !isEnd && !rescheduleOpdata.Force {
		blog.Info("transaction %s pending: old taskGroup(%s) not end at current",
			trans.TransactionID, taskGroup.ID)
		if task.CanTaskGroupShutdown(taskGroup) {
			blog.Info("transaction %s pending: kill old taskGroup(%s)", trans.TransactionID, taskGroup.ID)
			s.KillTaskGroup(taskGroup)
		}
		s.DeclineResource(offer.Id.Value)
		return true
	}

	blog.Infof("kill taskgroup(%s) before do reschedule", taskGroup.ID)
	s.KillTaskGroup(taskGroup)

	reschededTimes := taskGroup.ReschededTimes

	// build new taskgroup
	newTaskGroupID, _ := task.ReBuildTaskGroupID(taskGroupID)
	newTaskGroup, err := s.BuildTaskGroup(version, app, newTaskGroupID, "reschedule taskgroup")
	if err != nil {
		blog.Error("transaction %s fail: Build task failed: %s", trans.TransactionID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}
	blog.Info("transaction %s build group(%s) for reschedule", trans.TransactionID, newTaskGroup.ID)

	// remember reschedule information
	newTaskGroup.ReschededTimes = reschededTimes + 1
	newTaskGroup.LastReschedTime = time.Now().Unix()

	resources := task.BuildResources(version.AllResource())

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

	// lock agentsetting
	util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, newTaskGroup.GetAgentIp())
	// update agentsettings taskgroup index info
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

	// delete old taskgroup
	if err = s.DeleteTaskGroup(app, taskGroup, "do reschedule"); err != nil {
		blog.Error("transaction %s delete taskgroup(%s) failed: %s", trans.TransactionID, taskGroup.ID, err.Error())
	}

	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s save application(%s.%s) error %s",
			trans.TransactionID, app.RunAs, app.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}

	resp, err := s.LaunchTaskGroup(offer, newTaskGroupInfo, version)
	if err != nil {
		blog.Error("transaction %s launch taskgroup(%s) failed: %s", trans.TransactionID, newTaskGroup.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s request mesos resp code: %s", trans.TransactionID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return false
	}

	var alarmTimeval uint16 = 600
	s.SendHealthMsg(alarm.InfoKind, newTaskGroup.RunAs,
		fmt.Sprintf("taskgroup restarted, id:%s->%s", taskGroupID, newTaskGroup.ID),
		fmt.Sprintf("%s.%s-restart", newTaskGroup.RunAs, newTaskGroup.Name), &alarmTimeval)
	// launch taskgroup success, and metrics
	reportScheduleTaskgroupMetrics(trans.Namespace, trans.ObjectName,
		newTaskGroup.Name, RescheduleTaskgroupType, trans.CreateTime)
	trans.Status = types.OPERATION_STATUS_FINISH
	return false
}
