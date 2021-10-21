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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commontypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"net/http"
	"time"
)

// DeleteApplication will delete all data associated with application.
func (b *backend) DeleteApplication(runAs, appID string, enforce bool, kind commontypes.BcsDataType) error {
	blog.V(3).Infof("to delete application(%s.%s)", runAs, appID)

	b.store.LockApplication(runAs + "." + appID)
	defer b.store.UnLockApplication(runAs + "." + appID)

	app, err := b.store.FetchApplication(runAs, appID)
	if err != nil {
		blog.Warn("delete application(%s.%s) err %s", runAs, appID, err.Error())
		return err
	}
	if app == nil {
		blog.Warn("delete application(%s.%s), application not found", runAs, appID)
		return errors.New("Application not found")
	}

	// added  20181011, app.Kind is a new field and the current running app kind maybe empty.
	// if app.Kind is empty and request kind is PROCESS, the deletion will not be allowed.
	// if app.Kind is not empty and it is different from request kind, also the deletion will not be allowed.
	currentKind := app.Kind
	if currentKind == "" {
		currentKind = commontypes.BcsDataType_APP
	}
	if currentKind != kind {
		blog.Warnf("delete application(%s.%s) err, currentKind(%s) != kind(%s)", runAs, appID, currentKind, kind)
		return fmt.Errorf("target is not %s, can not be delete", kind)
	}
	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Warn("delete application(%s.%s) in status(%s)", runAs, appID, app.Status)
	}
	if app.Status == types.APP_STATUS_ROLLINGUPDATE && enforce == false {
		blog.Warn("delete application(%s.%s), in status(%s), cannot delete directory",
			runAs, appID, app.Status)
		return errors.New("Application in rollingupdate, cannot delete directory")
	}

	taskGroups, err := b.store.ListTaskGroups(runAs, appID)
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		blog.Info("delete application(%s.%s), kill taskgroup(%s)", runAs, appID, taskGroup.ID)
		resp, err := b.sched.KillTaskGroup(taskGroup)
		if err != nil {
			blog.Error("delete application(%s.%s), kill taskgroup(%s) failed: %s", runAs, appID, taskGroup.ID, err.Error())
			continue
		}
		if resp == nil {
			blog.Error("delete application(%s.%s), kill taskGroup(%s) resp nil", runAs, appID, taskGroup.ID)
			continue
		} else if resp.StatusCode != http.StatusAccepted {
			blog.Error("delete application(%s.%s), kill taskGroup(%s) resp code %d", runAs, appID, taskGroup.ID, resp.StatusCode)
			continue
		}
	}

	deleteTrans := &types.Transaction{
		ObjectKind:    string(commontypes.BcsDataType_APP),
		ObjectName:    appID,
		Namespace:     runAs,
		TransactionID: types.GenerateTransactionID(string(commontypes.BcsDataType_APP)),
		CreateTime:    time.Now(),
		CheckInterval: 3 * time.Second,
		CurOp: &types.TransactionOperartion{
			OpType: types.TransactionOpTypeDelete,
			OpDeleteData: &types.TransAPIDeleteOpdata{
				Enforce: enforce,
			},
		},
		Status: types.OPERATION_STATUS_INIT,
	}
	if err := b.store.SaveTransaction(deleteTrans); err != nil {
		blog.Errorf("save transaction(%s,%s) into db failed, err %s", runAs, appID, err.Error())
		return err
	}
	blog.Infof("transaction %s delete application(%s.%s) run begin",
		deleteTrans.TransactionID, deleteTrans.Namespace, deleteTrans.ObjectName)

	b.sched.PushEventQueue(deleteTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in deleting"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("delete application(%s.%s), save status(%s) err:%s", runAs, appID, app.Status, err.Error())
		return err
	}

	return nil
}
