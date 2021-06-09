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

//transaction for scaleup application

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

// RunScaleApplication The goroutine function for scale application transaction
// You can create a transaction for scale application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunScaleApplication(transaction *types.Transaction) bool {

	runAs := transaction.Namespace
	appID := transaction.ObjectName
	startedTaskgroup := time.Now()
	blog.Infof("transaction %s scale(%s.%s) run check", transaction.TransactionID, runAs, appID)
	// check begin
	if transaction.CreateTime.Add(transaction.DelayTime).After(time.Now()) {
		blog.Infof("transaction %s scale(%s.%s) delaytime(%d), cannot do now",
			transaction.TransactionID, runAs, appID, transaction.DelayTime)
		time.Sleep(3 * time.Second)
		return true
	}
	opData := transaction.CurOp.OpScaleData
	version := opData.Version

	if opData.IsDown {
		isContinue := s.doScaleDownAppTrans(transaction, false)
		if !isContinue {
			blog.Infof("transaction %s scaledown(%s.%s) finish", transaction.TransactionID, runAs, appID)
			return false
		}
	} else {
		transaction.CurOp.OpScaleData.SchedulerNum++
		offerOut := s.GetFirstOffer()
		for offerOut != nil {
			offer := offerOut.Offer

			curOffer := offerOut
			offerOut = s.GetNextOffer(offerOut)
			blog.V(3).Infof("transaction %s get offer %s||%s ",
				transaction.TransactionID, offer.GetHostname(), *(offer.Id.Value))
			isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) &&
				s.IsConstraintsFit(version, offer, "") &&
				s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
			if isFit == true {
				// when build new taskgroup, schedulerNumber==0
				transaction.CurOp.OpScaleData.SchedulerNum = 0
				blog.V(3).Infof("transaction %s fit offer %s||%s ",
					transaction.TransactionID, offer.GetHostname(), *(offer.Id.Value))
				if s.UseOffer(curOffer) == true {
					blog.Info("transaction %s scale(%s.%s) use offer %s||%s",
						transaction.TransactionID, runAs, appID, offer.GetHostname(), *(offer.Id.Value))
					launchedNum := opData.LaunchedNum
					isContinue := s.doScaleUpAppTrans(transaction, curOffer, false, startedTaskgroup)
					if !isContinue {
						blog.Infof("transaction %s scaleup(%s.%s) finish", transaction.TransactionID, runAs, appID)
						return false
					}
					if launchedNum < opData.LaunchedNum {
						startedTaskgroup = time.Now()
					}

				} else {
					blog.Info("transaction %s use offer %s||%s fail",
						transaction.TransactionID, offer.GetHostname(), *(offer.Id.Value))
				}
			}

		}
	}

	// when scheduler taskgroup number>=10, then report resources insufficient message
	if transaction.CurOp.OpScaleData.SchedulerNum >= 10 {
		blog.Warn("transaction %s scale(%s.%s) timeout", transaction.TransactionID, runAs, appID)
		transaction.Status = types.OPERATION_STATUS_TIMEOUT
		return false
	}
	return true
}

// RunInnerScaleApplication The goroutine function for inner scale application transaction
// You can create a transaction for scale application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunInnerScaleApplication(transaction *types.Transaction) bool {
	runAs := transaction.Namespace
	appID := transaction.ObjectName
	blog.Infof("transaction %s innerscale(%s.%s) run check", transaction.TransactionID, runAs, appID)

	// check begin
	if transaction.CreateTime.Add(transaction.DelayTime).After(time.Now()) {
		blog.Infof("transaction %s innerscale(%s.%s) delaytime(%d), cannot do at now",
			transaction.TransactionID, runAs, appID, transaction.DelayTime)
		time.Sleep(3 * time.Second)
		return true
	}

	// add  20181207, different pods may need differen resource, for example: requestIP
	opData := transaction.CurOp.OpScaleData
	version := opData.Version
	startedTaskgroup := time.Now()
	if opData.IsDown {
		s.doScaleDownAppTrans(transaction, true)
		if transaction.Status == types.OPERATION_STATUS_FINISH ||
			transaction.Status == types.OPERATION_STATUS_FAIL {
			blog.Infof("transaction %s innerscaledown(%s.%s) end", transaction.TransactionID, runAs, appID)
			return false
		}
	} else {
		offerOut := s.GetFirstOffer()
		for offerOut != nil {
			offerIdx := offerOut.Id
			offer := offerOut.Offer
			curOffer := offerOut
			offerOut = s.GetNextOffer(offerOut)
			blog.V(3).Infof("transaction %s get offer(%d) %s||%s ",
				transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
			isFit := s.IsOfferResourceFitLaunch(opData.NeedResource, curOffer) &&
				s.IsConstraintsFit(version, offer, "") &&
				s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), curOffer)
			if isFit == true {
				blog.V(3).Infof("transaction %s fit offer(%d) %s||%s ",
					transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				if s.UseOffer(curOffer) == true {
					blog.Info("transaction %s innerscale(%s.%s) use offer(%d) %s||%s",
						transaction.TransactionID, runAs, appID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
					launchedNum := opData.LaunchedNum
					s.doScaleUpAppTrans(transaction, curOffer, true, startedTaskgroup)
					if transaction.Status == types.OPERATION_STATUS_FINISH ||
						transaction.Status == types.OPERATION_STATUS_FAIL {
						blog.Infof("transaction %s innerscaleup(%s.%s) end", transaction.TransactionID, runAs, appID)
						return false
					}
					if launchedNum < opData.LaunchedNum {
						startedTaskgroup = time.Now()
					}
				} else {
					blog.Info("transaction %s use offer(%d) %s||%s fail",
						transaction.TransactionID, offerIdx, offer.GetHostname(), *(offer.Id.Value))
				}
			}
		}
	}
	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doScaleUpAppTrans(
	trans *types.Transaction, outOffer *offer.Offer, isInner bool, started time.Time) bool {

	runAs := trans.Namespace
	appID := trans.ObjectName

	offer := outOffer.Offer

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("transaction %s fetch application(%s.%s) error %s", trans.TransactionID, runAs, appID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "fetch application failed"
		s.DeclineResource(offer.Id.Value)
		return false
	}
	if app.Created > trans.CreateTime.Unix() {
		blog.Warn("transaction %s fail: application(%s.%s) CreateTime(%d) > transaction.CreateTime(%d)",
			trans.TransactionID, runAs, appID, app.Created, trans.CreateTime)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "application CreateTime > transaction CreateTime"
		s.DeclineResource(offer.Id.Value)
		return false
	}

	opData := trans.CurOp.OpScaleData
	if app.Instances >= opData.Instances {
		blog.Warnf("transaction %s already finish", trans.TransactionID)
		trans.Status = types.OPERATION_STATUS_FINISH
		s.DeclineResource(offer.Id.Value)
		return false
	}

	cpus, mem, disk := s.OfferedResources(offer)
	blog.Info("transaction %s scale application(%s.%s) with offer:%s||%s, cpu:%f, mem:%f, disk:%f",
		trans.TransactionID, app.RunAs, app.ID, offer.GetHostname(), *(offer.Id.Value), cpus, mem, disk)

	version := opData.Version
	resources := task.BuildResources(version.AllResource())
	taskGroupInfos := make([]*mesos.TaskGroupInfo, 0)

	var taskgroupName string
	if app.Instances < opData.Instances &&
		s.IsOfferResourceFitLaunch(version.AllResource(), outOffer) &&
		s.IsOfferExtendedResourcesFitLaunch(version.GetExtendedResources(), outOffer) {
		taskGroup, err := s.BuildTaskGroup(version, app, "", "scale application")
		if err != nil {
			blog.Error("transaction %s build taskgroup fail, err %s", trans.TransactionID, err.Error())
			trans.Status = types.OPERATION_STATUS_FAIL
			trans.Message = fmt.Sprintf("build taskgroup failed, err %s", err.Error())
			s.DeclineResource(offer.Id.Value)
			return false
		}
		taskgroupName = taskGroup.Name

		taskGroupInfo := task.CreateTaskGroupInfo(offer, version, resources, taskGroup)
		if taskGroupInfo == nil {
			blog.Error("transaction %s build taskgroupinfo(%s) fail", trans.TransactionID, taskGroup.ID)
			s.DeleteTaskGroup(app, taskGroup, "create taskgroupinfo fail")
			s.DeclineResource(offer.Id.Value)
			return true
		}

		if err := s.store.SaveTaskGroup(taskGroup); err != nil {
			blog.Error("transaction %s save taskgroup(%s) error %s", trans.TransactionID, taskGroup.ID, err.Error())
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
		blog.Error("transaction %s have no taskgroup to launch", trans.TransactionID)
		s.DeclineResource(offer.Id.Value)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = "has no taskgroup to launch"
		return false
	}

	resp, err := s.LaunchTaskGroups(offer, taskGroupInfos, version)
	if err != nil {
		blog.Error("transaction %s launch taskgroup fail", trans.TransactionID)
		s.DeclineResource(offer.Id.Value)
	}

	if resp != nil && resp.StatusCode != http.StatusAccepted {
		blog.Error("transaction %s request mesos err code : %d", trans.TransactionID, resp.StatusCode)
		s.DeclineResource(offer.Id.Value)
	}

	reportScheduleTaskgroupMetrics(app.RunAs, app.Name, taskgroupName, ScaleTaskgroupType, started)
	if app.Instances >= opData.Instances {
		if isInner == false {
			app.LastStatus = app.Status
			app.Status = types.APP_STATUS_DEPLOYING
			app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
			app.Message = "application in deploying"
			app.UpdateTime = time.Now().Unix()
		}
		trans.Status = types.OPERATION_STATUS_FINISH
	}

	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("scale transaction %s finish, save application(%s.%s) info into db failed! err:%s",
			trans.TransactionID, app.RunAs, app.ID, err.Error())
	} else {
		blog.Info("scale transaction %s finish, save application(%s.%s) info into db succ!",
			trans.TransactionID, app.RunAs, app.ID)
	}

	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doScaleDownAppTrans(trans *types.Transaction, isInner bool) bool {

	runAs := trans.Namespace
	appID := trans.ObjectName

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	app, err := s.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("transaction %s fetch application(%s.%s) error %s", trans.TransactionID, runAs, appID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("fetch application failed, err %s", err.Error())
		return false
	}
	if app.Created > trans.CreateTime.Unix() {
		blog.Warn("transaction %s fail: application(%s.%s) CreateTime(%d) > transaction.CreateTime(%d)",
			trans.TransactionID, runAs, appID, app.Created, trans.CreateTime)
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("application.CreateTime > transaction.CreateTime")
		return false
	}

	taskGroups, err := s.store.ListTaskGroups(app.RunAs, app.ID)
	if err != nil {
		blog.Error("transaction %s list taskgroup(%s.%s) err(%s)", trans.TransactionID, app.RunAs, app.ID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		trans.Message = fmt.Sprintf("list taskgroup failed, err %s", err.Error())
		return false
	}

	opData := trans.CurOp.OpScaleData
	// all taskgroups are end
	isEnd := true
	for _, taskGroup := range taskGroups {
		if taskGroup.InstanceID >= opData.Instances {
			if !task.IsTaskGroupEnd(taskGroup) {
				isEnd = false
				if task.CanTaskGroupShutdown(taskGroup) {
					blog.Info("transaction %s scaledown taskgroup(%s) not in end status, kill",
						trans.TransactionID, taskGroup.ID)
					s.KillTaskGroup(taskGroup)
					continue
				}
				blog.Info("transaction %s scaledown taskgroup(%s) not in end status at current",
					trans.TransactionID, taskGroup.ID)
			} else {
				blog.Info("transaction %s scaledown taskgroup(%s) in end status current",
					trans.TransactionID, taskGroup.ID)
			}
		}
	}

	if isEnd == false {
		blog.Info("transaction %s scaledown(%s.%s) not end currently", trans.TransactionID, app.RunAs, app.ID)
		return true
	}

	for _, taskGroup := range taskGroups {
		if taskGroup.InstanceID >= opData.Instances {
			if err = s.DeleteTaskGroup(app, taskGroup, "scale down application"); err != nil {
				blog.Error("transaction %s delete taskgroup(%s) failed: %s",
					trans.TransactionID, taskGroup.ID, err.Error())
				// TODO: do next tick when delete failed
				// but it may cause the shrinking operation to slow down
				// so the original logic is maintained for now
			} else {
				blog.Infof("transaction %s delete taskgroup(%s) success",
					trans.TransactionID, taskGroup.ID)
			}
		}
	}
	app.Instances = opData.Instances

	if isInner == false {
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.Message = types.APP_STATUS_RUNNING_STR
		app.UpdateTime = time.Now().Unix()
	}
	trans.Status = types.OPERATION_STATUS_FINISH
	err = s.store.SaveApplication(app)
	if err != nil {
		blog.Error("transaction %s save application(%s.%s) err:%s",
			trans.TransactionID, app.RunAs, app.ID, err.Error())
	} else {
		blog.V(3).Infof("transaction %s save application(%s.%s) succ!", trans.TransactionID, app.RunAs, app.ID)
	}

	blog.Info("transaction %s scaledown application(%s.%s) instances(%d)",
		trans.TransactionID, app.RunAs, app.ID, app.Instances)
	return false
}
