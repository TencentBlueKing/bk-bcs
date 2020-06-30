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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	sched "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"time"
)

func (b *backend) CheckVersion(version *types.Version) error {
	return task.CheckVersion(version, b.store)
}

func (b *backend) LaunchApplication(version *types.Version) error {
	blog.V(3).Infof("launch application(%s.%s)", version.RunAs, version.ID)

	runAs := version.RunAs
	appId := version.ID
	b.store.LockApplication(runAs + "." + appId)
	defer b.store.UnLockApplication(runAs + "." + appId)

	app, err := b.store.FetchApplication(runAs, appId)
	if err != nil {
		blog.Error("launch application(%s.%s) err: %s", runAs, appId, err.Error())
		return err
	}
	if app == nil {
		blog.Error("launch application(%s.%s) err, application not found", runAs, appId)
		return errors.New("Application not found")
	}

	launchTrans := sched.CreateTransaction()
	launchTrans.LifePeriod = sched.TRANSACTION_APPLICATION_LAUNCH_LIFEPERIOD
	launchTrans.RunAs = version.RunAs
	launchTrans.AppID = version.ID
	launchTrans.OpType = types.OPERATION_LAUNCH
	launchTrans.Status = types.OPERATION_STATUS_INIT
	var launchOpdata sched.TransAPILaunchOpdata
	launchOpdata.Version = version
	launchOpdata.LaunchedNum = 0
	launchOpdata.NeedResource = version.AllResource()
	launchOpdata.Reason = "request launch"
	launchTrans.OpData = &launchOpdata

	go b.sched.RunLaunchApplication(launchTrans)

	//app.RawJson = version.RawJson
	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in launching"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("save application(%s.%s) status(%s) into db err: %s",
			runAs, appId, app.Status, err.Error())
		return err
	}

	return nil
}

func (b *backend) RecoverApplication(version *types.Version) error {

	blog.Info("recover application(%s.%s) to instances(%d)",
		version.RunAs, version.ID, version.Instances)
	runAs := version.RunAs
	appId := version.ID
	b.store.LockApplication(runAs + "." + appId)
	defer b.store.UnLockApplication(runAs + "." + appId)

	app, err := b.store.FetchApplication(runAs, appId)
	if err != nil {
		blog.Error("recover application(%s.%s) err %s", runAs, appId, err.Error())
		return err
	}
	if app == nil {
		blog.Error("recover application(%s.%s) err, application not found", runAs, appId)
		return errors.New("Application not found")
	}

	blog.Info("recover application(%s.%s) from instance %d to  %d",
		runAs, appId, app.Instances, version.Instances)

	if app.Instances >= uint64(version.Instances) {
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		if err := b.store.SaveApplication(app); err != nil {
			blog.Error("save application(%s.%s) status(%s) into db err:%s, recover fail",
				runAs, appId, app.Status, err.Error())
			return err
		}
		return nil
	}

	launchTrans := sched.CreateTransaction()
	launchTrans.LifePeriod = sched.TRANSACTION_APPLICATION_LAUNCH_LIFEPERIOD
	launchTrans.RunAs = version.RunAs
	launchTrans.AppID = version.ID
	launchTrans.OpType = types.OPERATION_LAUNCH
	launchTrans.Status = types.OPERATION_STATUS_INIT
	launchTrans.DelayTime = 8
	var launchOpdata sched.TransAPILaunchOpdata
	launchOpdata.Version = version
	launchOpdata.LaunchedNum = int(app.Instances)
	launchOpdata.NeedResource = version.AllResource()
	launchOpdata.Reason = "recover"
	launchTrans.OpData = &launchOpdata

	go b.sched.RunLaunchApplication(launchTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in recovering"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("save application(%s.%s) status(%s) into db err:%s, recover fail",
			runAs, appId, app.Status, err.Error())
		return err
	}

	return nil
}
