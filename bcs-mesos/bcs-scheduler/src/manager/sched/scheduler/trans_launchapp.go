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

// RunLaunchApplication The goroutine function for launch application transaction
// You can create a transaction for launch application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunLaunchApplication(transaction *types.Transaction) bool {

	runAs := transaction.Namespace
	appID := transaction.ObjectName

	blog.Infof("transaction %s launch application (%s.%s) trigger", transaction.TransactionID, runAs, appID)

	startedTaskgroup := time.Now()
	transaction.CurOp.OpLaunchData.SchedulerNum = transaction.CurOp.OpLaunchData.SchedulerNum + 1
	// check doing
	opData := transaction.CurOp.OpLaunchData
	version := opData.Version

	offerOut := s.GetFirstOffer()
	for offerOut != nil {
		offerIdx := offerOut.Id
		offer := offerOut.Offer
		blog.V(3).Infof("transaction %s launch(%s.%s) get offer(%d) %s||%s ",
			transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))

		curOffer := offerOut
		offerOut = s.GetNextOffer(offerOut)
		isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) &&
			s.IsConstraintsFit(version, offer, "") &&
			s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
		if isFit == true {
			// when build new taskgroup, schedulerNumber==0
			transaction.CurOp.OpLaunchData.SchedulerNum = 0
			blog.V(3).Infof("transaction %s launch(%s.%s) fit offer(%d) %s||%s ",
				transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
			if s.UseOffer(curOffer) == true {
				blog.Info("transaction %s launch(%s.%s) use offer(%d) %s||%s",
					transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				launchedNum := opData.LaunchedNum
				isContinue := s.doLaunchTrans(transaction, curOffer, startedTaskgroup)
				if !isContinue {
					blog.Infof("transaction %s launch(%s.%s) end", transaction.TransactionID, runAs, appID)
					return false
				}
				if launchedNum < opData.LaunchedNum {
					startedTaskgroup = time.Now()
				}

			} else {
				blog.Info("transaction %s launch(%s.%s) use offer(%d) %s||%s fail",
					transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
			}
		}
	}

	// when scheduler taskgroup number>=10, then report resources insufficient message
	if transaction.CurOp.OpLaunchData.SchedulerNum >= 10 {
		blog.Warn("transaction %s launch(%s.%s) timeout", transaction.TransactionID, runAs, appID)
		transaction.Status = types.OPERATION_STATUS_TIMEOUT
		return false
	}
	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doLaunchTrans(trans *types.Transaction, outOffer *offer.Offer, started time.Time) bool {

	blog.Infof("do transaction %s launch(%s.%s) begin", trans.TransactionID, trans.ObjectName, trans.Namespace)
	offer := outOffer.Offer
	cpus, mem, disk := s.OfferedResources(offer)

	opData := trans.CurOp.OpLaunchData
	version := opData.Version
	resources := task.BuildResources(version.AllResource())

	var taskGroupInfos []*mesos.TaskGroupInfo

	runAs := trans.Namespace
	appID := trans.ObjectName
	blog.Infof("transaction %s launch application(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.TransactionID, runAs, appID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if app == nil {
		blog.Errorf("transaction %s launch application(%s.%s) fail: fetch application return nil",
			trans.TransactionID, runAs, appID)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "fetch application return nil"
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if app.Created > trans.CreateTime.Unix() {
		blog.Warnf(
			"transaction %s launch application fail: application(%s.%s) Created(%d) > transaction.CreateTime(%d)",
			trans.TransactionID, runAs, appID, app.Created, trans.CreateTime.Unix())
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "application.Created > transaction.CreateTime"
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if app.Status != types.APP_STATUS_OPERATING {
		blog.Warnf("transaction %s launch application(%s %s) status is %s",
			trans.TransactionID, runAs, appID, app.Status)
		trans.Status = types.OPERATION_STATUS_FINISH
		trans.Message = fmt.Sprintf("application status is %s", app.Status)
		s.DeclineResource(offer.Id.Value)
		return false
	}

	var taskgroupName string
	var taskGroup *types.TaskGroup
	if opData.LaunchedNum < int(version.Instances) && s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) &&
		s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		taskGroup, err := s.BuildTaskGroup(version, app, "", "launch application")
		if err != nil {
			blog.Error("transaction %s launch application(%s %s): build taskgroup err: %s",
				trans.TransactionID, runAs, appID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			trans.Message = fmt.Sprintf("build taskgroup err %s", err.Error())
			s.DeclineResource(offer.Id.Value)
			return false
		}
		taskgroupName = taskGroup.Name

		taskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, taskGroup)
		if taskGroupInfo == nil {
			blog.Warn("transaction %s launch application(%s %s): build taskgroupinfo fail",
				trans.TransactionID, runAs, appID)
			s.DeleteTaskGroup(app, taskGroup, "create taskgroupinfo fail")
			s.DeclineResource(offer.Id.Value)
			return true
		}

		if err := s.store.SaveTaskGroup(taskGroup); err != nil {
			blog.Error("transaction %s launch application(%s %s): save taskgroup error %s",
				trans.TransactionID, runAs, appID, err.Error())
			s.DeleteTaskGroup(app, taskGroup, "save taskgroup fail")
			s.DeclineResource(offer.Id.Value)
			return true
		}
		opData.LaunchedNum++
		taskGroupInfos = append(taskGroupInfos, taskGroupInfo)

		// lock agentsetting
		util.Lock.Lock(commtypes.BcsClusterAgentSetting{}, taskGroup.GetAgentIp())
		// update agentsettings taskgroup index info
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
		blog.Error("transaction %s launch application(%s %s): have no any taskgroup to launch for this offer",
			trans.TransactionID, runAs, appID)
		s.DeclineResource(offer.Id.Value)
		return true
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s launch application(%s %s): launch taskgroups err: %s",
			trans.TransactionID, runAs, appID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("launch taskgroups err %s", err.Error())
		s.DeleteTaskGroup(app, taskGroup, err.Error())
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s launch application(%s %s): resp status err code : %d",
			trans.TransactionID, runAs, appID, resp.StatusCode)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "launch taskgroups response err"
		s.DeleteTaskGroup(app, taskGroup, err.Error())
		s.DeclineResource(offer.Id.Value)
		return false
	}

	// launch taskgroup success, and metrics
	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, LaunchTaskgroupType, started)

	if opData.LaunchedNum >= int(version.Instances) {
		blog.Info("transaction %s launch application(%s %s) finish", trans.TransactionID, runAs, appID)
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_DEPLOYING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = "application in deploying"
		trans.Status = types.OPERATION_STATUS_FINISH
		err = s.store.SaveApplication(app)
		if err != nil {
			blog.Error(
				"transaction %s launch application(%s %s) finish, set application to APP_STATUS_DEPLOYING err:%s",
				trans.TransactionID, app.RunAs, app.ID, err.Error())
			// when save application failed, try again in later transaction check, so do not set OPERATION_STATUS_FAIL
			return true
		}
		blog.Info("transaction %s launch application(%s %s) finish, set application to APP_STATUS_DEPLOYING",
			trans.TransactionID, app.RunAs, app.ID)
		blog.Info("do transaction %s launch application(%s %s) end", trans.TransactionID, app.RunAs, app.ID)
		return false
	}
	// BuildTaskgroup will change instances field of application, so save application here
	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error(
			"transaction %s launch application(%s %s) contine, launched %d/%d, save application err %s",
			trans.TransactionID, app.RunAs, app.ID, opData.LaunchedNum, int(version.Instances), err.Error())
		// when save application failed, try again in later transaction check, so do not set OPERATION_STATUS_FAIL
		return true
	}
	blog.Info("transaction %s launch application(%s %s) continue, launched %d/%d",
		trans.TransactionID, app.RunAs, app.ID, opData.LaunchedNum, int(version.Instances))
	return true
}
