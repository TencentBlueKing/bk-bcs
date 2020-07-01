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
	"net/http"
	"time"

	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/offer"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// The goroutine function for reschedule taskgroup transaction
// You can create a transaction for reschedule taskgroup, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunRescheduleTaskgroup(transaction *Transaction) {

	rescheduleOpdata := transaction.OpData.(*TransRescheduleOpData)
	taskGroupID := rescheduleOpdata.TaskGroupID

	blog.Infof("transaction %s reschedule(%s) run begin", transaction.ID, taskGroupID)

	started := time.Now()
	var schedulerNumber int64
	for {
		schedulerNumber++
		blog.Infof("transaction %s reschedule(%s) run check", transaction.ID, taskGroupID)

		//check begin
		if transaction.CreateTime+transaction.DelayTime > time.Now().Unix() {
			blog.V(3).Infof("transaction %s reschedule(%s) delaytime(%d), cannot do at now",
				transaction.ID, taskGroupID, transaction.DelayTime)
			time.Sleep(3 * time.Second)
			continue
		}

		//precheck application&taskgroup status
		taskgroup, _ := s.store.FetchTaskGroup(taskGroupID)
		if taskgroup == nil {
			blog.Infof("transaction %s fail: fetch taskgroup(%s) return nil", transaction.ID, taskGroupID)
			transaction.Status = types.OPERATION_STATUS_FAIL
			break
		}
		runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
		app, _ := s.store.FetchApplication(runAs, appID)
		if app == nil {
			blog.Infof("transaction %s fail: fetch application(%s.%s) return nil", transaction.ID, runAs, appID)
			transaction.Status = types.OPERATION_STATUS_FAIL
			break
		}
		//check doing
		opData := transaction.OpData.(*TransRescheduleOpData)
		version := opData.Version

		//if inner rescheduler, then check app status
		if opData.IsInner {
			if app.Status == types.APP_STATUS_OPERATING {
				blog.Infof("transaction %s pending: app(%s.%s) is in status(%s:%s), cannot do at now",
					transaction.ID, runAs, appID, app.Status, app.SubStatus)
				time.Sleep(3 * time.Second)
				continue
			}
			if app.Status == types.APP_STATUS_ROLLINGUPDATE && app.SubStatus != types.APP_SUBSTATUS_ROLLINGUPDATE_UP {
				blog.Infof("transaction %s pending: app(%s.%s) is in status(%s:%s), cannot do at now",
					transaction.ID, runAs, appID, app.Status, app.SubStatus)
				time.Sleep(3 * time.Second)
				continue
			}
		}

		hostRetain := false
		if transaction.CreateTime+opData.HostRetainTime > time.Now().Unix() {
			hostRetain = true
		}

		offerOut := s.GetFirstOffer()
		for offerOut != nil {
			offerIdx := offerOut.Id
			offer := offerOut.Offer
			blog.V(3).Infof("transaction %s get offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

			curOffer := offerOut
			offerOut = s.GetNextOffer(offerOut)
			if hostRetain == false || offer.GetHostname() == opData.HostRetain {
				isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) && s.IsConstraintsFit(version, offer, taskGroupID) &&
					s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
				if isFit == true {
					blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					if s.UseOffer(curOffer) == true {
						blog.Info("transaction %s reschedule(%s) use offer(%d) %s||%s", transaction.ID, taskGroupID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
						s.doRescheduleTrans(transaction, curOffer, started)
						break
					} else {
						blog.Info("transaction %s use offer(%d) %s||%s fail", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					}
				}
			}

		}

		//check end
		if transaction.Status == types.OPERATION_STATUS_FINISH || transaction.Status == types.OPERATION_STATUS_FAIL {
			blog.Infof("transaction %s reschedule(%s) finish", transaction.ID, taskGroupID)
			break
		}

		// when scheduler taskgroup number>=10, then report resources insufficient message
		if schedulerNumber == 10 {
			s.store.LockApplication(runAs + "." + appID)
			taskGroup, _ := s.store.FetchTaskGroup(taskGroupID)
			if taskGroup != nil {
				taskGroup.Message = "don't have fit resources to reschedule this taskgroup"
				s.store.SaveTaskGroup(taskGroup)
			}
			s.store.UnLockApplication(runAs + "." + appID)
		}

		//check timeout
		if (transaction.CreateTime + transaction.LifePeriod) < time.Now().Unix() {
			blog.Warn("transaction %s reschedule(%s) timeout", transaction.ID, taskGroupID)
			transaction.Status = types.OPERATION_STATUS_TIMEOUT
			break
		}

		time.Sleep(3 * time.Second)
	}

	s.FinishTransaction(transaction)
	blog.Infof("transaction %s reschedule(%s) run end, result(%s)", transaction.ID, taskGroupID, transaction.Status)

	return
}

func (s *Scheduler) passRescheduleCheck(trans *Transaction, app *types.Application) bool {

	if app.Created > trans.CreateTime {
		blog.Warn("transaction %s fail: application(%s.%s) Created(%d) > transaction.CreateTime(%d)",
			trans.ID, app.RunAs, app.ID, app.Created, trans.CreateTime)
		trans.Status = types.OPERATION_STATUS_FAIL
		return false
	}

	opData := trans.OpData.(*TransRescheduleOpData)
	if !opData.IsInner {
		return true
	}

	if app.Status == types.APP_STATUS_OPERATING {
		blog.Warn("transaction %s pending: app(%s.%s) is in status(%s:%s), redo later",
			trans.ID, app.RunAs, app.ID, app.Status, app.SubStatus)
		return false
	}
	if app.Status == types.APP_STATUS_ROLLINGUPDATE && app.SubStatus != types.APP_SUBSTATUS_ROLLINGUPDATE_UP {
		blog.Warn("transaction %s pending: app(%s.%s) is in status(%s:%s), redo later",
			trans.ID, app.RunAs, app.ID, app.Status, app.SubStatus)
		return false
	}

	return true
}

func (s *Scheduler) doRescheduleTrans(trans *Transaction, outOffer *offer.Offer, started time.Time) {

	blog.Infof("do transaction %s begin", trans.ID)

	offer := outOffer.Offer

	rescheduleOpdata := trans.OpData.(*TransRescheduleOpData)
	taskGroupID := rescheduleOpdata.TaskGroupID
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)

	cpus, mem, disk := s.OfferedResources(offer)
	blog.Info("transaction %s reschedule taskgroup(%s), offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.ID, taskGroupID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Warn("transaction %s fail: fetch application(%s.%s) return nil", trans.ID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	if !s.passRescheduleCheck(trans, app) {
		s.DeclineResource(offer.Id.Value)
		return
	}

	version, _ := s.store.GetVersion(runAs, appID)
	if version == nil {
		blog.Error("transaction %s fail: no version for application(%s.%s)", trans.ID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	var taskgroupName string
	//reschededTimes := 0
	taskGroup, err := s.store.FetchTaskGroup(taskGroupID)
	if taskGroup == nil {
		blog.Info("transaction %s fetch taskGroup(%s) fail: %s", trans.ID, taskGroupID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	taskgroupName = taskGroup.Name

	if rescheduleOpdata.IsInner && taskGroup.Status == types.TASKGROUP_STATUS_RUNNING {
		blog.Info("transaction %s finish: lost taskGroup(%s) recover to running",
			trans.ID, taskGroupID)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		var alarmTimeval uint16 = 600
		s.SendHealthMsg(alarm.InfoKind, taskGroup.RunAs, "taskgroup recover to running, id:"+taskGroupID, taskGroup.RunAs+"."+taskGroup.Name+"-restart", &alarmTimeval)
		return
	}

	isEnd := task.CanTaskGroupReschedule(taskGroup)
	if !isEnd && !rescheduleOpdata.Force {
		blog.Info("transaction %s pending: old taskGroup(%s) not end at current",
			trans.ID, taskGroup.ID)
		if task.CanTaskGroupShutdown(taskGroup) {
			blog.Info("transaction %s pending: kill old taskGroup(%s)", trans.ID, taskGroup.ID)
			s.KillTaskGroup(taskGroup)
		}
		s.DeclineResource(offer.Id.Value)
		return
	}

	blog.Infof("kill taskgroup(%s) before do reschedule", taskGroup.ID)
	s.KillTaskGroup(taskGroup)

	reschededTimes := taskGroup.ReschededTimes

	// build new taskgroup
	newTaskGroupID, _ := task.ReBuildTaskGroupID(taskGroupID)
	newTaskGroup, err := s.BuildTaskGroup(version, app, newTaskGroupID, "reschedule taskgroup")
	if err != nil {
		blog.Error("transaction %s fail: Build task failed: %s", trans.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	blog.Info("transaction %s build group(%s) for reschedule", trans.ID, newTaskGroup.ID)

	// remember reschedule information
	newTaskGroup.ReschededTimes = reschededTimes + 1
	newTaskGroup.LastReschedTime = time.Now().Unix()

	resources := task.BuildResources(version.AllResource())

	newTaskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, newTaskGroup)
	if newTaskGroupInfo == nil {
		blog.Error("transaction %s build taskgroupinfo(%s) failed", trans.ID, newTaskGroup.ID)
		s.DeleteTaskGroup(app, newTaskGroup, "create taskgroupinfo fail")
		s.DeclineResource(offer.Id.Value)
		return
	}

	if err = s.store.SaveTaskGroup(newTaskGroup); err != nil {
		blog.Error("transaction %s save taskgroup(%s) error %s", trans.ID, newTaskGroup.ID, err.Error())
		s.DeleteTaskGroup(app, newTaskGroup, "save taskgroup fail")
		s.DeclineResource(offer.Id.Value)
		return
	}

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

	//delete old taskgroup
	if err = s.DeleteTaskGroup(app, taskGroup, "do reschedule"); err != nil {
		blog.Error("transaction %s delete taskgroup(%s) failed: %s", trans.ID, taskGroup.ID, err.Error())
	}

	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s save application(%s.%s) error %s", trans.ID, app.RunAs, app.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	resp, err := s.LaunchTaskGroup(offer, newTaskGroupInfo, version)
	if err != nil {
		blog.Error("transaction %s launch taskgroup(%s) failed: %s", trans.ID, newTaskGroup.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s request mesos resp code: %s", trans.ID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}

	//launch taskgroup success, and metrics
	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, RescheduleTaskgroupType, started)
	trans.Status = types.OPERATION_STATUS_FINISH
	blog.Info("transaction %s reschedule new taskgroup(%s) succeed", trans.ID, newTaskGroupID)

	var alarmTimeval uint16 = 600
	s.SendHealthMsg(alarm.InfoKind, newTaskGroup.RunAs, "taskgroup restarted, id:"+taskGroupID+"->"+newTaskGroupID, newTaskGroup.RunAs+"."+newTaskGroup.Name+"-restart", &alarmTimeval)

	blog.Infof("do transaction %s end", trans.ID)
	return
}
