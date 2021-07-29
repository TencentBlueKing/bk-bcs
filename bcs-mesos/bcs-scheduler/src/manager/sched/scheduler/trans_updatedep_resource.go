/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
)

// RunUpdateDeploymentResource do update deployment resource transaction
func (s *Scheduler) RunUpdateDeploymentResource(transaction *types.Transaction) bool {
	name := transaction.ObjectName
	ns := transaction.Namespace
	if transaction.ObjectKind != types.TransactionObjectKindDeployment {
		transaction.Status = types.OPERATION_STATUS_FAIL
		transaction.Message = fmt.Sprintf("do UpdateDeploymentResource but object (%s.%s) is not %s",
			ns, name, types.TransactionObjectKindDeployment)
		blog.Errorf(transaction.Message)
		return false
	}

	// trigger transaction about application related to this deployment
	if len(transaction.ChildTransacationID) == 0 {
		currDeployment, err := s.store.FetchDeployment(ns, name)
		if err != nil {
			transaction.Status = types.OPERATION_STATUS_FAIL
			transaction.Message = fmt.Sprintf("get deployment(%s.%s) failed when UpdateDeploymentResource, err %s",
				ns, name, err.Error())
			blog.Errorf(transaction.Message)
			return false
		}
		appName := currDeployment.Application.ApplicationName
		s.store.LockApplication(ns + "." + appName)
		defer s.store.UnLockApplication(ns + "." + appName)
		currApp, err := s.store.FetchApplication(ns, appName)
		if err != nil {
			transaction.Status = types.OPERATION_STATUS_FAIL
			transaction.Message = fmt.Sprintf("get application(%s.%s) failed when UpdateDeploymentResource, err %s",
				ns, appName, err.Error())
			blog.Errorf(transaction.Message)
			return false
		}
		updateAppTrans := &types.Transaction{
			TransactionID: types.GenerateTransactionID(string(commtypes.BcsDataType_APP)),
			ObjectKind:    string(commtypes.BcsDataType_APP),
			ObjectName:    appName,
			Namespace:     ns,
			CreateTime:    time.Now(),
			CheckInterval: 3 * time.Second,
			CurOp: &types.TransactionOperartion{
				OpType: types.TransactionOpTypeUpdate,
			},
			Status: types.OPERATION_STATUS_INIT,
		}
		updateOpdata := types.TransAPIUpdateOpdata{}
		updateOpdata.Version = transaction.CurOp.OpDepUpdateData.Version
		updateOpdata.LaunchedNum = 0
		updateOpdata.NeedResource = transaction.CurOp.OpDepUpdateData.Version.AllResource()
		updateOpdata.Instances = int(currApp.Instances)
		updateOpdata.IsUpdateResource = true
		updateOpdata.Taskgroups, err = s.store.ListTaskGroups(ns, appName)
		if err != nil {
			transaction.Status = types.OPERATION_STATUS_FAIL
			transaction.Message = fmt.Sprintf(
				"get taskgroups of application(%s.%s) failed when UpdateDeploymentResource, err %s",
				ns, appName, err.Error())
			blog.Errorf(transaction.Message)
			return false
		}
		updateAppTrans.CurOp.OpUpdateData = &updateOpdata

		// set application status
		currApp.LastStatus = currApp.Status
		currApp.Status = types.APP_STATUS_OPERATING
		currApp.SubStatus = types.APP_SUBSTATUS_UNKNOWN
		currApp.UpdateTime = time.Now().Unix()
		currApp.Message = "application in updating"
		if err := s.store.SaveApplication(currApp); err != nil {
			blog.Error("update application(%s.%s) status(%s), save application err:%s",
				currApp.RunAs, currApp.ID, currApp.Status, err.Error())
			return false
		}

		// trigger application transaction
		if err := s.store.SaveTransaction(updateAppTrans); err != nil {
			blog.Errorf("save transaction(%s,%s) into db failed, err %s", ns, appName, err.Error())
			return true
		}
		s.PushEventQueue(updateAppTrans)

		// save child transactionID into parent transaction
		transaction.ChildTransacationID = updateAppTrans.TransactionID
		return true
	}

	// check child transaction status
	_, err := s.store.FetchTransaction(ns, transaction.ChildTransacationID)
	if err == store.ErrNoFound {
		blog.Infof("child transaction %s not found, think it is finished", transaction.ChildTransacationID)
		transaction.Status = types.OPERATION_STATUS_FINISH
		return false
	}
	blog.Infof("child transaction %s found, think it is not finished", transaction.ChildTransacationID)
	return true
}
