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
	"net/http"

	"bk-bcs/bcs-common/common/blog"
	commonTypes "bk-bcs/bcs-common/common/types"
	sched "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

const (
	DefaultRescheduleDelayTime = 5 // seconds
)

//RescheduleTaskgroup is used to reschedule taskgroup.
func (b *backend) RescheduleTaskgroup(taskgroupId string, hostRetainTime int64) error {
	blog.Infof("reschedule taskgroup(%s)", taskgroupId)

	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskgroupId)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("reschedule taskgroup(%s) fail, fetch application(%s.%s) err:%s", taskgroupId, runAs, appID, err.Error())
		return err
	}

	if app == nil {
		blog.Error("reschedule taskgroup(%s) fail, get application(%s.%s) return nil", taskgroupId, runAs, appID)
		return errors.New("Application not found")
	}
	/*if app.Status == types.APP_STATUS_OPERATING {
		blog.Warn("reschedule taskgroup(%s) fail, application(%s.%s) status(%s) err", taskgroupId, runAs, appID, app.Status)
		return errors.New("Operation Not Allowed")
	}
	if app.Status == types.APP_STATUS_ROLLINGUPDATE && app.SubStatus != types.APP_SUBSTATUS_ROLLINGUPDATE_UP {
		blog.Error("reschedule taskgroup(%s) fail, application(%s.%s) status(%s:%s) err",
			taskgroupId, runAs, appID, app.Status, app.SubStatus)
		return errors.New("operation Not Allowed")
	}*/

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	//versions, err := b.store.ListVersions(runAs, appID)
	//if err != nil {
	//	blog.Error("reschedule taskgroup(%s) fail, list version(%s.%s) err:%s", taskgroupId, runAs, appID, err.Error())
	//	return err
	//}
	//sort.Strings(versions)
	//newestVersion := versions[len(versions)-1]
	version, _ := b.store.GetVersion(runAs, appID)
	if version == nil {
		blog.Error("reschedule taskgroup(%s) fail, no version for application(%s.%s)", taskgroupId, runAs, appID)
		return errors.New("application version not exist")
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

	rescheduleTrans := sched.CreateTransaction()
	rescheduleTrans.RunAs = runAs
	rescheduleTrans.AppID = appID
	rescheduleTrans.OpType = types.OPERATION_RESCHEDULE
	rescheduleTrans.Status = types.OPERATION_STATUS_INIT
	rescheduleTrans.DelayTime = int64(DefaultRescheduleDelayTime)

	var rescheduleOpdata sched.TransRescheduleOpData

	rescheduleOpdata.TaskGroupID = taskgroup.ID
	rescheduleOpdata.Force = true
	rescheduleOpdata.IsInner = false
	rescheduleOpdata.HostRetainTime = hostRetainTime
	if rescheduleOpdata.HostRetainTime > 0 {
		blog.Info("taskgroup(%s) will reschedule retain host(%s) for %d seconds",
			taskgroup.ID, taskgroup.HostName, rescheduleOpdata.HostRetainTime)
		rescheduleOpdata.HostRetain = taskgroup.HostName
	} else {
		rescheduleOpdata.HostRetainTime = 0
	}

	rescheduleOpdata.NeedResource = version.AllResource()
	rescheduleOpdata.Version = version

	rescheduleTrans.OpData = &rescheduleOpdata

	go b.sched.RunRescheduleTaskgroup(rescheduleTrans)

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

	if app.Kind != commonTypes.BcsDataType_PROCESS {
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

	if app.Kind != commonTypes.BcsDataType_PROCESS {
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
