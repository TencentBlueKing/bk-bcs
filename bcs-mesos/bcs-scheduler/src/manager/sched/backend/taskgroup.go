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

package backend

import (
	"errors"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

const (
	DefaultRescheduleDelayTime = 5 // seconds
)

//RescheduleTaskgroup is used to reschedule taskgroup.
func (b *backend) RescheduleTaskgroup(taskgroupId string, hostRetainTime int64) error {
	blog.Infof("reschedule taskgroup(%s)", taskgroupId)
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskgroupId)
	//check taskgroup whether belongs to daemonset
	if b.sched.CheckPodBelongDaemonset(taskgroupId) {
		util.Lock.Lock(types.BcsDaemonset{}, runAs+"."+appID)
		defer util.Lock.UnLock(types.BcsDaemonset{}, runAs+"."+appID)
	} else {
		b.store.LockApplication(runAs + "." + appID)
		defer b.store.UnLockApplication(runAs + "." + appID)
	}
	taskgroup, err := b.store.FetchTaskGroup(taskgroupId)
	if err != nil {
		blog.Errorf("reschedule taskgroup(%s) fail, fetch taskgroup err: %s", taskgroupId, err.Error())
		return err
	}
	// here kill taskGroup
	resp, err := b.sched.KillTaskGroup(taskgroup)
	if err != nil {
		blog.Warn("taskgroup(%s) reschedule under status(%s) but do kill failed: %s", taskgroup.ID, taskgroup.Status, err.Error())
	}
	if resp == nil {
		blog.Warn("taskgroup(%s) reschedule under status(%s) but kill resp == nil", taskgroup.ID, taskgroup.Status)
	} else if resp.StatusCode != http.StatusAccepted {
		blog.Warn("taskgroup(%s) reschedule under status(%s) but kill return code %d", taskgroup.ID, taskgroup.Status, resp.StatusCode)
	}

	// set LOST to FAIL to release resource for new taskgroup
	if taskgroup.Status == types.TASKGROUP_STATUS_LOST {
		blog.Warn("taskgroup(%s) force to reschedule under LOST, set status to FAIL", taskgroup.ID)
		taskgroup.Status = types.TASKGROUP_STATUS_FAIL
		err := b.store.SaveTaskGroup(taskgroup)
		if err != nil {
			blog.Errorf("save taskgroup(%s) error %s", taskgroup.ID, err.Error())
			return err
		}
	}
	//if taskgroup belongs to daemonsets, then don't trigger reschedule transaction
	if b.sched.CheckPodBelongDaemonset(taskgroupId) {
		return nil
	}

	//trigger taskgroup reschedule
	version, _ := b.store.GetVersion(runAs, appID)
	if version == nil {
		blog.Error("reschedule taskgroup(%s) fail, no version for application(%s.%s)", taskgroupId, runAs, appID)
		return errors.New("application version not exist")
	}

	rescheduleTrans := &types.Transaction{
		ObjectKind:    string(commtypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_APP)),
		CreateTime:    time.Now(),
		CheckInterval: 3 * time.Second,
		Status:        types.OPERATION_STATUS_INIT,
	}
	rescheduleOpdata := &types.TransactionOperartion{
		OpType: types.TransactionOpTypeReschedule,
		OpRescheduleData: &types.TransRescheduleOpData{
			TaskGroupID:    taskgroup.ID,
			Force:          true,
			IsInner:        false,
			HostRetainTime: hostRetainTime,
		},
	}
	if rescheduleOpdata.OpRescheduleData.HostRetainTime > 0 {
		blog.Info("taskgroup(%s) will rescheduled retain host(%s) for %d seconds",
			taskgroup.ID, taskgroup.HostName, rescheduleOpdata.OpRescheduleData.HostRetainTime)
		rescheduleOpdata.OpRescheduleData.HostRetain = taskgroup.HostName
	} else {
		rescheduleOpdata.OpRescheduleData.HostRetainTime = 0
	}

	rescheduleOpdata.OpRescheduleData.NeedResource = version.AllResource()
	rescheduleOpdata.OpRescheduleData.Version = version
	rescheduleTrans.CurOp = rescheduleOpdata

	if err := b.store.SaveTransaction(rescheduleTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}
	b.sched.PushEventQueue(rescheduleTrans)

	return nil
}

// RestartTaskGroup is used to restart process taskGroup. If the taskGroup type is container, then return error.
func (b *backend) RestartTaskGroup(taskGroupID string) (*types.BcsMessage, error) {
	blog.V(3).Infof("to restart taskgroup(%s)", taskGroupID)
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("restart taskgroup(%s), fetch application(%s.%s) err:%s", taskGroupID, runAs, appID, err.Error())
		return nil, err
	}

	if app == nil {
		blog.Error("restart taskgroup(%s), get application(%s.%s) return nil", taskGroupID, runAs, appID)
		return nil, errors.New("application not found")
	}

	if app.Kind != commtypes.BcsDataType_PROCESS {
		blog.Errorf("restart taskgroup(%s), application(%s.%s), type is %s, restart only for process", taskGroupID, app.RunAs, app.Name, app.Kind)
		return nil, errors.New("application type is not process")
	}

	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Error("restart taskgroup(%s),  application(%s.%s) status(%s) err", taskGroupID, runAs, appID, app.Status)
		return nil, errors.New("operation Not Allowed")
	}

	taskGroup, err := b.store.FetchTaskGroup(taskGroupID)
	if err != nil {
		blog.Errorf("restart taskgroup(%s), fetch error:%s", taskGroupID, err.Error())
		return nil, err
	}

	bcsMsg := &types.BcsMessage{
		Type:       types.Msg_RESTART_TASK.Enum(),
		CommitTask: &types.Msg_CommitTask{},
	}

	return b.sched.SendBcsMessage(taskGroup, bcsMsg)
}

// ReloadTaskGroup is used to reload process taskGroup. If the taskGroup type is container, then return error.
func (b *backend) ReloadTaskGroup(taskGroupID string) (*types.BcsMessage, error) {
	blog.V(3).Infof("to reload taskgroup(%s)", taskGroupID)
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("reload taskgroup(%s), fetch application(%s.%s) err:%s", taskGroupID, runAs, appID, err.Error())
		return nil, err
	}

	if app == nil {
		blog.Error("reload taskgroup(%s), get application(%s.%s) return nil", taskGroupID, runAs, appID)
		return nil, errors.New("application not found")
	}

	if app.Kind != commtypes.BcsDataType_PROCESS {
		blog.Errorf("reload taskgroup(%s), application(%s.%s), type is %s, restart only for process", taskGroupID, app.RunAs, app.Name, app.Kind)
		return nil, errors.New("application type is not process")
	}

	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Error("reload taskgroup(%s),  application(%s.%s) status(%s) err", taskGroupID, runAs, appID, app.Status)
		return nil, errors.New("operation Not Allowed")
	}

	taskGroup, err := b.store.FetchTaskGroup(taskGroupID)
	if err != nil {
		blog.Errorf("reload taskgroup(%s), fetch taskGroup error: %s", taskGroupID, err.Error())
		return nil, err
	}

	bcsMsg := &types.BcsMessage{
		Type:       types.Msg_RELOAD_TASK.Enum(),
		CommitTask: &types.Msg_CommitTask{},
	}

	return b.sched.SendBcsMessage(taskGroup, bcsMsg)
}
