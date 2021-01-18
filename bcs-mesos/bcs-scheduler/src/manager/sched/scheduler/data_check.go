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
	"time"

	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
)

// DataCheckMsg DataChecker control message
type DataCheckMsg struct {
	// opencheck
	// closecheck
	// stop
	MsgType string
}

// DataCheckMgr DataChecker
type DataCheckMgr struct {
	store store.Store

	// we need this to call some functions
	sched *Scheduler

	msgQueue chan *DataCheckMsg

	//appCheckStatusTime int64
	//appTimeoutTime     int64

	openCheck bool
}

// CreateDataCheckMgr Create DataChecker
func CreateDataCheckMgr(store store.Store, s *Scheduler) (*DataCheckMgr, error) {

	mgr := &DataCheckMgr{
		store:     store,
		sched:     s,
		openCheck: false,
	}

	// create msg queue for events
	mgr.msgQueue = make(chan *DataCheckMsg, 128)

	return mgr, nil
}

func (mgr *DataCheckMgr) stop() {
	blog.V(3).Infof("data checker stop...")
	close(mgr.msgQueue)
}

// SendMsg Send DataChecker control message
func (mgr *DataCheckMgr) SendMsg(msg *DataCheckMsg) error {

	blog.V(3).Infof("data checker: send an msg to datacheck manager")

	select {
	case mgr.msgQueue <- msg:
	default:
		blog.Error("data checker: send an msg to datacheck manager fail")
		return fmt.Errorf("data checker: mgr queue is full now")
	}

	return nil
}

// DataCheckManage DataChecker main function
func DataCheckManage(mgr *DataCheckMgr, doRecover bool) {
	blog.Info("data checker: goroutine start ...")

	/*if doRecover {
		time.Sleep(180 * time.Second)
		blog.Warn("data checker will set all LOST&FINISH taskgroup to RUNNING !!!!!!!")
		mgr.doRecover()
	}*/

	for {
		select {
		case req := <-mgr.msgQueue:
			blog.Info("data checker: receive msg: %s", req.MsgType)
			if req.MsgType == "opencheck" {
				mgr.openCheck = true
			} else if req.MsgType == "closecheck" {
				mgr.openCheck = false
			} else if req.MsgType == "stop" {
				mgr.stop()
				blog.Info("data checker: goroutine finish!")
			}

		case <-time.After(time.Second * time.Duration(DATA_CHECK_INTERVAL)):
			blog.Info("data checker: to do data check")
			mgr.doCheck()
		}
	}
}

func (mgr *DataCheckMgr) doCheck() {

	if mgr.openCheck == false {
		blog.Info("data checker: opencheck is false, do nothing")
		return
	}

	// add mesos heartbeat checking
	mgr.sched.lockService()
	heartbeat := mgr.sched.mesosHeartBeatTime
	now := time.Now().Unix()
	mgr.sched.unlockService()

	blog.Info("data checker: doCheck begin.... now(%d)", now)

	if now-heartbeat > MESOS_HEARTBEAT_TIMEOUT {
		blog.Warn("data checker: mesos heartbeat timeout, now(%d), heatbeat(%d), so skip checking",
			now, heartbeat)
		return
	}

	defer func() {
		now = time.Now().Unix()
		blog.Info("data checker: doCheck end.... now(%d)", now)
	}()

	runAses, err := mgr.store.ListRunAs()
	if err != nil {
		blog.Error("data checker: fail to list runAses, err:%s", err.Error())
		return
	}

	//check application taskgroup whether lost
	for _, runAs := range runAses {
		blog.Info("data checker: to check runAs(%s)", runAs)
		appIDs, err := mgr.store.ListApplicationNodes(runAs)
		if err != nil {
			blog.Error("data checker: fail to list %s, err:%s", runAs, err.Error())
			continue
		}
		if nil == appIDs {
			blog.Warn("data checker: no application nodes under runAs:%s", runAs)
			continue
		}
		for _, appID := range appIDs {
			blog.Info("data checker: to check application:%s.%s ", runAs, appID)
			//mgr.checkApplication(runAs, appID)
			mgr.checkTaskgroup(runAs, appID)
		}
	}
	//check daemonsets taskgroup whether lost
	daemonsets, err := mgr.store.ListAllDaemonset()
	if err != nil {
		blog.Error("data checker ListAllDaemonset failed: %s", err.Error())
		return
	}
	for _, daemon := range daemonsets {
		blog.Info("data checker: to check daemonest(%s) taskgroup status", daemon.GetUuid())
		//lock
		util.Lock.Lock(types.BcsDaemonset{}, daemon.GetUuid())
		taskgroups, _ := mgr.store.ListDaemonsetTaskGroups(daemon.NameSpace, daemon.Name)
		lostNum := mgr.checkTaskgroupWhetherLost(taskgroups, false)
		if lostNum == 0 {
			util.Lock.UnLock(types.BcsDaemonset{}, daemon.GetUuid())
			continue
		}
		//if have new lost taskgroup, then update daemonset status
		mgr.sched.updateDaemonsetStatus(daemon.NameSpace, daemon.Name)
		util.Lock.UnLock(types.BcsDaemonset{}, daemon.GetUuid())
	}

	return
}

func (mgr *DataCheckMgr) checkTaskgroup(runAs, appID string) {
	mgr.sched.store.LockApplication(runAs + "." + appID)
	defer mgr.sched.store.UnLockApplication(runAs + "." + appID)

	blog.Info("data checker: to check taskgroups:%s.%s ", runAs, appID)
	taskGroups, _ := mgr.store.ListTaskGroups(runAs, appID)
	//check taskgroup whether lost
	lostNum := mgr.checkTaskgroupWhetherLost(taskGroups, true)
	if lostNum == 0 {
		return
	}
	//  get application data from ZK
	app, err := mgr.sched.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("data checker: fetch application(%s.%s) failed, err:%s", runAs, appID, err.Error())
		return
	}
	appStatus := app.Status
	isUpdated, _ := mgr.sched.updateApplicationStatus(app)
	if isUpdated == true {
		if err = mgr.sched.store.SaveApplication(app); err != nil {
			blog.Error("data checker: save application(%s.%s) into db failed! err:%s",
				app.RunAs, app.ID, err.Error())
			return
		}
	}
	mgr.sched.applicationStatusUpdated(app, appStatus)
}

//if taskgroup status long time no update, then show taskgroup lost
//application't taskgroup trigger reschedule, and daemonset't taskgroup don't trigger reschedule
func (mgr *DataCheckMgr) checkTaskgroupWhetherLost(taskGroups []*types.TaskGroup, reschedule bool) int {
	now := time.Now().Unix()
	lostnum := 0
	for _, taskGroup := range taskGroups {
		if taskGroup.LastUpdateTime == 0 ||
			(taskGroup.Status != types.TASKGROUP_STATUS_RUNNING &&
				taskGroup.Status != types.TASKGROUP_STATUS_STAGING &&
				taskGroup.Status != types.TASKGROUP_STATUS_STARTING) {
			continue
		}

		var updateInterval int64

		switch taskGroup.Status {
		case types.TASKGROUP_STATUS_RUNNING, types.TASKGROUP_STATUS_STARTING:
			updateInterval = 4 * MAX_DATA_UPDATE_INTERVAL
			/*case types.TASKGROUP_STATUS_STAGING:
			updateInterval = 4 * MAX_STAGING_UPDATE_INTERVAL*/
		}

		if taskGroup.LastUpdateTime+updateInterval < now {
			for _, task := range taskGroup.Taskgroup {
				if task.Status != types.TASK_STATUS_RUNNING && task.Status != types.TASK_STATUS_STAGING && task.Status != types.TASK_STATUS_STARTING {
					continue
				}
				if task.LastUpdateTime+updateInterval < now {
					blog.Warn("data checker: task(%s) lost for long time not report status under status(%s)",
						task.ID, task.Status)
					task.LastStatus = task.Status
					task.Status = types.TASK_STATUS_LOST
					task.LastUpdateTime = now
					mgr.sched.SendHealthMsg(alarm.WarnKind, taskGroup.RunAs, task.ID+"("+taskGroup.HostName+")"+"long time not report status, set status to lost", "", nil)
				}
			}

			lostnum++
			blog.Warn("data checker: taskgroup(%s) lost for long time not report status under status(%s)",
				taskGroup.ID, taskGroup.Status)
			taskGroupStatus := taskGroup.Status
			taskGroup.LastStatus = taskGroup.Status
			taskGroup.Status = types.TASKGROUP_STATUS_LOST
			taskGroup.LastUpdateTime = now
			//if reschedule==true, then trigger reschedule function
			if reschedule {
				mgr.sched.taskGroupStatusUpdated(taskGroup, taskGroupStatus)
				mgr.sched.ServiceMgr.TaskgroupUpdate(taskGroup)
			}
			if err := mgr.sched.store.SaveTaskGroup(taskGroup); err != nil {
				blog.Errorf("data checker: save taskgroup(%s) into db failed! err:%s", taskGroup.ID, err.Error())
				return 0
			}
		}
	}

	return lostnum
}

//donot call this function !!!!!!!!!!!
func (mgr *DataCheckMgr) doRecover() {

	now := time.Now().Unix()
	blog.Info("data checker: doRecover begin.... now(%d)", now)
	defer func() {
		now = time.Now().Unix()
		blog.Info("data checker: doRecover end.... now(%d)", now)
	}()

	runAses, err := mgr.store.ListRunAs()
	if err != nil {
		blog.Error("data checker: fail to list runAses, err:%s", err.Error())
		return
	}

	for _, runAs := range runAses {
		appIDs, err := mgr.store.ListApplicationNodes(runAs)
		if err != nil {
			blog.Error("data checker: fail to list %s, err:%s", runAs, err.Error())
			continue
		}
		if nil == appIDs {
			blog.Warn("data checker: no application nodes under runAs:%s", runAs)
			continue
		}
		for _, appID := range appIDs {
			blog.Info("data checker: to recover application:%s.%s ", runAs, appID)
			mgr.recoverTaskgroup(runAs, appID)
		}
	}

	return
}

func (mgr *DataCheckMgr) recoverTaskgroup(runAs, appID string) {
	now := time.Now().Unix()
	mgr.sched.store.LockApplication(runAs + "." + appID)
	defer mgr.sched.store.UnLockApplication(runAs + "." + appID)

	taskGroups, err := mgr.store.ListTaskGroups(runAs, appID)
	if err != nil {
		blog.Warn("data checker: list taskgroup(%s.%s) return err:%s", runAs, appID, err.Error())
		return
	}
	blog.Info("data checker: to recover taskgroups:%s.%s ", runAs, appID)

	recoverNum := 0
	for _, taskGroup := range taskGroups {
		if taskGroup.Status == types.TASKGROUP_STATUS_LOST || taskGroup.Status == types.TASKGROUP_STATUS_FINISH {
			for _, task := range taskGroup.Taskgroup {
				if task.Status == types.TASK_STATUS_LOST {
					blog.Warn("data checker: recover task %s from %s to running ", task.ID, task.Status)
					task.LastStatus = task.Status
					task.Status = types.TASK_STATUS_RUNNING
					task.LastUpdateTime = now
				}
			}

			recoverNum++
			blog.Warn("data checker: recover taskgroup %s from %s to running", taskGroup.ID, taskGroup.Status)
			taskGroupStatus := taskGroup.Status
			taskGroup.LastStatus = taskGroup.Status
			taskGroup.Status = types.TASKGROUP_STATUS_RUNNING
			taskGroup.LastUpdateTime = now
			mgr.sched.taskGroupStatusUpdated(taskGroup, taskGroupStatus)
			mgr.sched.ServiceMgr.TaskgroupUpdate(taskGroup)
			if err = mgr.sched.store.SaveTaskGroup(taskGroup); err != nil {
				blog.Error("data checker: save taskgroup(%s) into db failed! err:%s", taskGroup.ID, err.Error())
				return
			}
		}
	}

	if recoverNum > 0 {
		app, err := mgr.sched.store.FetchApplication(runAs, appID)
		if err != nil {
			blog.Error("data checker: fetch application(%s.%s) failed, err:%s", runAs, appID, err.Error())
			return
		}
		appStatus := app.Status
		isUpdated, err := mgr.sched.updateApplicationStatus(app)
		if isUpdated == true {
			if err = mgr.sched.store.SaveApplication(app); err != nil {
				blog.Error("data checker: save application(%s.%s) into db failed! err:%s",
					app.RunAs, app.ID, err.Error())
				return
			}
		}

		mgr.sched.applicationStatusUpdated(app, appStatus)
	}
}
