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
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

//UpdateApplication is used for application rolling-update.
func (b *backend) UpdateApplication(runAs, appID string, args string, instances int, version *types.Version) error {

	blog.V(3).Infof("update application(%s.%s): args(%s), instances(%d)", runAs, appID, args, instances)

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Error("get application(%s.%s) to do update err %s", runAs, appID, err.Error())
		return err
	}

	if app == nil {
		blog.Error("get application(%s.%s) to do update return nil", runAs, appID)
		return errors.New("Application not found")
	}

	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Warn("application(%s.%s) cannot do update under status(%s)", runAs, appID, app.Status)
		return errors.New("Operation Not Allowed")
	}

	updateTrans := &types.Transaction{
		TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_APP)),
		ObjectKind:    string(commtypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		CreateTime:    time.Now(),
		CheckInterval: time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeUpdate,
		},
		Status: types.OPERATION_STATUS_INIT,
	}

	updateOpdata := types.TransAPIUpdateOpdata{}
	updateOpdata.Version = version
	updateOpdata.LaunchedNum = 0
	updateOpdata.NeedResource = version.AllResource()
	updateOpdata.Instances = instances

	updateOpdata.Taskgroups, err = b.store.ListTaskGroups(runAs, appID)
	if err != nil {
		blog.Error("list taskgroups(%s.%s) to do update err: %s", runAs, appID, err.Error())
		return err
	}

	//add taskgroup number check
	if len(updateOpdata.Taskgroups) == 0 {
		blog.Error("list taskgroups(%s.%s) return empty", runAs, appID)
		return errors.New("no taskgroups to update")
	}
	//check end

	blog.Info("taskgroups before sort:")

	for _, taskGroup := range updateOpdata.Taskgroups {
		blog.Info("taskgroup: %s", taskGroup.ID)
	}
	sort.Sort(TaskSorter(updateOpdata.Taskgroups))

	blog.Info("taskgroups after sort:")
	for _, taskGroup := range updateOpdata.Taskgroups {
		blog.Info("taskgroup: %s", taskGroup.ID)
	}

	if args == "resource" {
		updateOpdata.IsUpdateResource = true
		updateOpdata.Instances = len(updateOpdata.Taskgroups)
		updateTrans.CurOp.OpUpdateData = &updateOpdata
	} else {
		//correct the instances for update
		if updateOpdata.Instances > len(updateOpdata.Taskgroups) {
			blog.Warn("request update instances(%d) > taskgroups num(%d)",
				updateOpdata.Instances, len(updateOpdata.Taskgroups))
			updateOpdata.Instances = len(updateOpdata.Taskgroups)
		}
		if updateOpdata.Instances <= 0 {
			updateOpdata.Instances = len(updateOpdata.Taskgroups)
		}
		updateTrans.CurOp.OpUpdateData = &updateOpdata
		// kill old taskgroup
		index := 0
		for index < updateOpdata.Instances {
			taskGroup := updateOpdata.Taskgroups[index]
			blog.Info("kill taskgroup(%s) for update", taskGroup.ID)
			resp, err := b.sched.KillTaskGroup(taskGroup)
			if err != nil {
				blog.Warn("kill taskgroup(%s) err: %s", taskGroup.ID, err.Error())
			}
			if resp == nil {
				blog.Warn("kill taskgroup(%s), resp == nil")
			} else if resp.StatusCode != http.StatusAccepted {
				blog.Warn("kill taskgroup(%s), return code %d", resp.StatusCode)
			}
			index++
		}
	}
	if err := b.store.SaveTransaction(updateTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}
	b.sched.PushEventQueue(updateTrans)

	//app.RawJson = version.RawJson
	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in updating"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("update application(%s.%s) status(%s), save application err:%s",
			app.RunAs, app.ID, app.Status, err.Error())
		return err
	}
	return nil
}

func (b *backend) HealthyReport(healthyResult *commtypes.HealthCheckResult) {
	b.sched.HealthyReport(healthyResult)
	return
}
