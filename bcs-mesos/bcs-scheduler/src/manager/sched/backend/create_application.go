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
	commontypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"time"
)

func (b *backend) CheckVersion(version *types.Version) error {
	return task.CheckVersion(version, b.store)
}

func (b *backend) LaunchApplication(version *types.Version) error {
	blog.V(3).Infof("launch application(%s.%s)", version.RunAs, version.ID)

	runAs := version.RunAs
	appID := version.ID
	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("launch application(%s.%s) err: %s", runAs, appID, err.Error())
		return err
	}
	if app == nil {
		blog.Error("launch application(%s.%s) err, application not found", runAs, appID)
		return errors.New("Application not found")
	}

	launchTrans := &types.Transaction{
		ObjectKind:    string(commontypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		TransactionID: types.GenerateTransactionID(string(commontypes.BcsDataType_APP)),
		CreateTime:    time.Now(),
		CheckInterval: time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeLaunch,
			OpLaunchData: &types.TransAPILaunchOpdata{
				Version:      version,
				LaunchedNum:  0,
				NeedResource: version.AllResource(),
				Reason:       "request launch",
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}

	if err := b.store.SaveTransaction(launchTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}

	b.sched.PushEventQueue(launchTrans)

	//app.RawJson = version.RawJson
	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in launching"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("save application(%s.%s) status(%s) into db err: %s",
			runAs, appID, app.Status, err.Error())
		return err
	}

	return nil
}

func (b *backend) RecoverApplication(version *types.Version) error {

	blog.Info("recover application(%s.%s) to instances(%d)",
		version.RunAs, version.ID, version.Instances)
	runAs := version.RunAs
	appID := version.ID
	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("recover application(%s.%s) err %s", runAs, appID, err.Error())
		return err
	}
	if app == nil {
		blog.Error("recover application(%s.%s) err, application not found", runAs, appID)
		return errors.New("Application not found")
	}

	blog.Info("recover application(%s.%s) from instance %d to  %d",
		runAs, appID, app.Instances, version.Instances)

	if app.Instances >= uint64(version.Instances) {
		app.LastStatus = app.Status
		app.Status = types.APP_STATUS_RUNNING
		app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		app.UpdateTime = time.Now().Unix()
		app.Message = types.APP_STATUS_RUNNING_STR
		if err := b.store.SaveApplication(app); err != nil {
			blog.Error("save application(%s.%s) status(%s) into db err:%s, recover fail",
				runAs, appID, app.Status, err.Error())
			return err
		}
		return nil
	}

	launchTrans := &types.Transaction{
		ObjectKind:    string(commontypes.BcsDataType_APP),
		ObjectName:    version.ID,
		Namespace:     version.RunAs,
		TransactionID: types.GenerateTransactionID(string(commontypes.BcsDataType_APP)),
		CreateTime:    time.Now(),
		CheckInterval: time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeLaunch,
			OpLaunchData: &types.TransAPILaunchOpdata{
				Version:      version,
				LaunchedNum:  int(app.Instances),
				NeedResource: version.AllResource(),
				Reason:       "recover",
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}

	if err := b.store.SaveTransaction(launchTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}

	b.sched.PushEventQueue(launchTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in recovering"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("save application(%s.%s) status(%s) into db err:%s, recover fail",
			runAs, appID, app.Status, err.Error())
		return err
	}

	return nil
}
