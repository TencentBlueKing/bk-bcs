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

// transaction for launch application

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

// RunLaunchApplication The goroutine function for launch application transaction
// You can create a transaction for launch application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunLaunchApplication(transaction *Transaction) {

	runAs := transaction.RunAs
	appID := transaction.AppID

	blog.Infof("transaction %s launch(%s.%s) run begin", transaction.ID, runAs, appID)

	startedTaskgroup := time.Now()
	startedApp := time.Now()
	var schedulerNumber int64
	for {
		blog.Infof("transaction %s launch(%s.%s) run check", transaction.ID, runAs, appID)
		//check begin
		if transaction.CreateTime+transaction.DelayTime > time.Now().Unix() {
			blog.V(3).Infof("transaction %s launch(%s.%s) delaytime(%d), cannot do at now",
				transaction.ID, runAs, appID, transaction.DelayTime)
			time.Sleep(3 * time.Second)
			continue
		}
		schedulerNumber++

		//check doing
		opData := transaction.OpData.(*TransAPILaunchOpdata)
		version := opData.Version

		offerOut := s.GetFirstOffer()
		for offerOut != nil {
			offerIdx := offerOut.Id
			offer := offerOut.Offer
			blog.V(3).Infof("transaction %s get offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

			curOffer := offerOut
			offerOut = s.GetNextOffer(offerOut)
			//isFit := s.IsResourceFit(opData.NeedResource, offer) && s.IsConstraintsFit(version, offer, "")
			isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) && s.IsConstraintsFit(version, offer, "") &&
				s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
			if isFit == true {
				//when build new taskgroup, schedulerNumber==0
				schedulerNumber = 0
				blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				if s.UseOffer(curOffer) == true {
					blog.Info("transaction %s launch(%s.%s) use offer(%d) %s||%s", transaction.ID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					launchedNum := opData.LaunchedNum
					s.doLaunchTrans(transaction, curOffer, startedTaskgroup)
					if transaction.Status == types.OPERATION_STATUS_FINISH || transaction.Status == types.OPERATION_STATUS_FAIL {
						blog.Infof("transaction %s launch(%s.%s) end", transaction.ID, runAs, appID)
						goto run_end
					}
					if launchedNum < opData.LaunchedNum {
						startedTaskgroup = time.Now()
					}

				} else {
					blog.Info("transaction %s use offer(%d) %s||%s fail", transaction.ID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				}
			}
		}

		// when scheduler taskgroup number>=10, then report resources insufficient message
		if schedulerNumber >= 10 {
			blog.Warn("transaction %s launch(%s.%s) timeout", transaction.ID, runAs, appID)
			transaction.Status = types.OPERATION_STATUS_TIMEOUT
			goto run_end
		}

		//check timeout
		if (transaction.CreateTime + transaction.LifePeriod) < time.Now().Unix() {
			blog.Warn("transaction %s launch(%s.%s) timeout", transaction.ID, runAs, appID)
			transaction.Status = types.OPERATION_STATUS_TIMEOUT
			goto run_end
		}

		time.Sleep(time.Second)
	}

run_end:
	s.FinishTransaction(transaction)
	reportOperateAppMetrics(transaction.RunAs, transaction.AppID, LaunchApplicationType, startedApp)
	blog.Infof("transaction %s launch(%s.%s) run end, result(%s)", transaction.ID, runAs, appID, transaction.Status)
}

func (s *Scheduler) doLaunchTrans(trans *Transaction, outOffer *offer.Offer, started time.Time) {

	blog.Infof("do transaction %s begin", trans.ID)

	offer := outOffer.Offer

	cpus, mem, disk := s.OfferedResources(offer)

	var taskGroupInfos []*mesos.TaskGroupInfo

	opData := trans.OpData.(*TransAPILaunchOpdata)
	version := opData.Version
	resources := task.BuildResources(version.AllResource())

	runAs := trans.RunAs
	appID := trans.AppID
	blog.Info("transaction %s launch application(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.ID, runAs, appID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Error("transaction %s fail: fetch application(%s.%s) return nil", trans.ID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	if app.Created > trans.CreateTime {
		blog.Warn("transaction %s fail: application(%s.%s) Created(%d) > transaction.CreateTime(%d)",
			trans.ID, runAs, appID, app.Created, trans.CreateTime)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeclineResource(offer.Id.Value)
		return
	}
	if app.Status != types.APP_STATUS_OPERATING {
		blog.Warn("transaction %s: application(%s %s) status is %s", trans.ID, runAs, appID, app.Status)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		return
	}

	var taskgroupName string
	var taskGroup *types.TaskGroup
	if opData.LaunchedNum < int(version.Instances) && s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) &&
		s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		//if opData.LaunchedNum < int(version.Instances) && version.IsResourceFit(types.Resource{Cpus: cpus, Mem: mem, Disk: disk}) {
		taskGroup, err := s.BuildTaskGroup(version, app, "", "launch application")
		if err != nil {
			blog.Error("transaction %s: build taskgroup err: %s", trans.ID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}
		taskgroupName = taskGroup.Name

		taskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, taskGroup)
		if taskGroupInfo == nil {
			blog.Warn("transaction %s: build taskgroupinfo fail", trans.ID)
			s.DeleteTaskGroup(app, taskGroup, "create taskgroupinfo fail")
			//trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}

		if err := s.store.SaveTaskGroup(taskGroup); err != nil {
			blog.Error("transaction %s: save taskgroup error %s", trans.ID, err.Error())
			s.DeleteTaskGroup(app, taskGroup, "save taskgroup fail")
			//trans.Status = types.OPERATION_STATUS_FAIL
			s.DeclineResource(offer.Id.Value)
			return
		}
		opData.LaunchedNum++
		taskGroupInfos = append(taskGroupInfos, taskGroupInfo)

		//lock agentsetting
		util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, taskGroup.GetAgentIp())
		//update agentsettings taskgroup index info
		agentsetting, _ := s.store.FetchAgentSetting(taskGroup.GetAgentIp())
		if agentsetting != nil {
			agentsetting.Pods = append(agentsetting.Pods, taskGroup.ID)
			err := s.store.SaveAgentSetting(agentsetting)
			if err != nil {
				blog.Errorf("save agentsetting %s pods error %s", agentsetting.InnerIP, err.Error())
			}
		} else {
			blog.Errorf("fetch agentsetting %s Not Found", taskGroup.GetAgentIp())
		}
		util.Lock.UnLock(commtypes.BcsClusterAgentSetting{}, taskGroup.GetAgentIp())
	}

	if len(taskGroupInfos) <= 0 {
		blog.Error("transaction %s: have no any taskgroup to launch for this offer", trans.ID)
		s.DeclineResource(offer.Id.Value)
		return
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s: launch taskgroups err: %s", trans.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeleteTaskGroup(app, taskGroup, err.Error())
		s.DeclineResource(offer.Id.Value)
		return
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s: resp status err code : %d", trans.ID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		s.DeleteTaskGroup(app, taskGroup, err.Error())
		s.DeclineResource(offer.Id.Value)
		return
	}

	//launch taskgroup success, and metrics
	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, LaunchTaskgroupType, started)

	if opData.LaunchedNum >= int(version.Instances) {
		blog.Info("transaction %s finish", trans.ID)
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_DEPLOYING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = "application in deploying"
		trans.Status = types.OPERATION_STATUS_FINISH
	}

	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s finish, set application(%s.%s) to APP_STATUS_DEPLOYING err:%s",
			trans.ID, app.RunAs, app.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		return
	}
	blog.Info("transaction %s finish, set application(%s.%s) to APP_STATUS_DEPLOYING", trans.ID, app.RunAs, app.ID)
	blog.Info("do transaction %s end", trans.ID)
	return
}
