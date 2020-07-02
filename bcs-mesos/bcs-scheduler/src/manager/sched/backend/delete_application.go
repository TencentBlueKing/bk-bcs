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
	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	sched "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"net/http"
	"time"
)

// DeleteApplication will delete all data associated with application.
func (b *backend) DeleteApplication(runAs, appId string, enforce bool, kind commonTypes.BcsDataType) error {
	blog.V(3).Infof("to delete application(%s.%s)", runAs, appId)

	b.store.LockApplication(runAs + "." + appId)
	defer b.store.UnLockApplication(runAs + "." + appId)

	app, err := b.store.FetchApplication(runAs, appId)
	if err != nil {
		blog.Warn("delete application(%s.%s) err %s", runAs, appId, err.Error())
		return err
	}
	if app == nil {
		blog.Warn("delete application(%s.%s), application not found", runAs, appId)
		return errors.New("Application not found")
	}

	// added  20181011, app.Kind is a new field and the current running app kind maybe empty.
	// if app.Kind is empty and request kind is PROCESS, the deletion will not be allowed.
	// if app.Kind is not empty and it is different from request kind, also the deletion will not be allowed.
	currentKind := app.Kind
	if currentKind == "" {
		currentKind = commonTypes.BcsDataType_APP
	}
	if currentKind != kind {
		blog.Warnf("delete application(%s.%s) err, currentKind(%s) != kind(%s)", runAs, appId, currentKind, kind)
		return fmt.Errorf("target is not %s, can not be delete", kind)
	}
	if app.Status == types.APP_STATUS_OPERATING || app.Status == types.APP_STATUS_ROLLINGUPDATE {
		blog.Warn("delete application(%s.%s) in status(%s)", runAs, appId, app.Status)
	}
	if app.Status == types.APP_STATUS_ROLLINGUPDATE && enforce == false {
		blog.Warn("delete application(%s.%s), in status(%s), cannot delete directory",
			runAs, appId, app.Status)
		return errors.New("Application in rollingupdate, cannot delete directory")
	}

	taskGroups, err := b.store.ListTaskGroups(runAs, appId)
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		blog.Info("delete application(%s.%s), kill taskgroup(%s)", runAs, appId, taskGroup.ID)
		resp, err := b.sched.KillTaskGroup(taskGroup)
		if err != nil {
			blog.Error("delete application(%s.%s), kill taskgroup(%s) failed: %s", runAs, appId, taskGroup.ID, err.Error())
			continue
		}
		if resp == nil {
			blog.Error("delete application(%s.%s), kill taskGroup(%s) resp nil", runAs, appId, taskGroup.ID)
			continue
		} else if resp.StatusCode != http.StatusAccepted {
			blog.Error("delete application(%s.%s), kill taskGroup(%s) resp code %d", runAs, appId, taskGroup.ID, resp.StatusCode)
			continue
		}
	}

	deleteTrans := sched.CreateTransaction()
	deleteTrans.RunAs = runAs
	deleteTrans.AppID = appId
	deleteTrans.OpType = types.OPERATION_DELETE
	deleteTrans.Status = types.OPERATION_STATUS_INIT
	deleteTrans.DelayTime = 0
	deleteTrans.LifePeriod = 7500

	var deleteOpdata sched.TransAPIDeleteOpdata
	deleteOpdata.Enforce = enforce
	deleteTrans.OpData = &deleteOpdata

	go b.sched.RunDeleteApplication(deleteTrans)

	app.LastStatus = app.Status
	app.Status = types.APP_STATUS_OPERATING
	app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
	app.UpdateTime = time.Now().Unix()
	app.Message = "application in deleting"
	if err := b.store.SaveApplication(app); err != nil {
		blog.Error("delete application(%s.%s), save status(%s) err:%s", runAs, appId, app.Status, err.Error())
		return err
	}

	return nil
}
