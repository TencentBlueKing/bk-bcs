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
	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"time"
)

// Default transaction max lifeperoid, 480 seconds
// the lifeperoid for all not specified transactions are set to this
const TRANSACTION_DEFAULT_LIFEPERIOD = 480

const TRANSACTION_APPLICATION_LAUNCH_LIFEPERIOD = 1800

// Max lifeperoid for every rolling transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_ROLLING_UP_LIFEPERIOD = 300

// Max lifeperoid for every rolling transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_ROLLING_DOWN_LIFEPERIOD = 7500

// Max lifeperoid for innder delete application transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_INNERDELETE_LIFEPERIOD = 7500

// Max lifePeriod for inner taskgroup-reschedule, 3600 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_INNER_RESCHEDULE_LIFEPERIOD = 86400

// If taskgroup running than 1800 seconds, the restart times will be reset to 0
const TRANSACTION_RESCHEDULE_RESET_INTERVAL = 1800

// Transaction
type Transaction struct {
	// transaction unique ID, created in CreateTransaction
	ID string
	// namepace
	RunAs string
	// application name
	AppID string
	// operation type: LAUNCH, DELETE, SCALE, UPDATE, RESCHEDULE ...
	OpType string
	// operation status: INIT, FINISH, FAIL, ERROR ...
	Status string
	// operation data
	OpData interface{}
	// the seconds before transaction timeout
	LifePeriod int64
	// the seconds before transaction really excute
	DelayTime int64
	// transaction create time
	CreateTime int64
}

// Launch application transaction data
type TransAPILaunchOpdata struct {
	// version definition for launch application
	Version *types.Version
	// already launched taskgroups number
	LaunchedNum int
	// resource for a taskgroup
	NeedResource *types.Resource
	// why do this operation
	Reason string
}

// Scale application transaction data
type TransAPIScaleOpdata struct {
	// version definition for application
	Version *types.Version
	// resource for a taskgroup
	NeedResource *types.Resource
	// the target count for application's taskgroups
	Instances uint64
	// scale down or up
	IsDown bool
	// already launched taskgroups number
	LaunchedNum int
}

// Update application transaction data
type TransAPIUpdateOpdata struct {
	// version definition for application
	Version *types.Version
	// already updated count
	LaunchedNum int
	// the count of taskgroups to be updated
	Instances int
	// resource for one taskgroup
	NeedResource *types.Resource
	// the taskgroups to be updated
	Taskgroups []*types.TaskGroup
}

// Delete application transaction data
type TransAPIDeleteOpdata struct {
	// if false, the operation will fail when some taskgroups cannot come to end status
	Enforce bool
}

// Reschedule taskgroup transaction data
type TransRescheduleOpData struct {
	// version definition for application
	Version *types.Version
	// the taskgroup to be rescheduled
	TaskGroupID string
	// if the taskgroup cannot come to end status, do the operation or not
	Force bool
	// the operation is created by schedulder( taskgroup fail or lost ) or not
	IsInner bool
	// resource for one taskgroup
	NeedResource *types.Resource
	// host retain time
	HostRetainTime int64
	// host retain
	HostRetain string
}

// Create a transaction, ID, createTime will be initialized
func CreateTransaction() *Transaction {
	transaction := new(Transaction)
	//tmp
	transaction.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	transaction.CreateTime = time.Now().Unix()
	transaction.LifePeriod = TRANSACTION_DEFAULT_LIFEPERIOD
	transaction.DelayTime = 0
	blog.Info("create transaction %s, current time(%d)", transaction.ID, transaction.CreateTime)
	return transaction
}

// Finish a transaction, set application status
func (s *Scheduler) FinishTransaction(transaction *Transaction) {

	blog.Info("transaction(%s)(runAs:%s, ID:%s) type(%s) status(%s) end",
		transaction.ID, transaction.RunAs, transaction.AppID, transaction.OpType, transaction.Status)

	runAs := transaction.RunAs
	appID := transaction.AppID

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	if transaction.OpType == types.OPERATION_DELETE {
		return
	}
	app, err := s.store.FetchApplication(runAs, appID)
	if err != nil || app == nil {
		blog.Warn("transaction %s finished, fetch application(%s.%s) err:%s", transaction.ID, runAs, appID, err.Error())
		return
	}

	if transaction.Status == types.OPERATION_STATUS_TIMEOUT {
		s.SendHealthMsg(alarm.WarnKind, app.RunAs, "transaction("+transaction.ID+") timeout: "+transaction.OpType+" "+runAs+"."+appID, "", nil)
	}

	if transaction.OpType == types.OPERATION_INNERSCALE {
		if transaction.Status != types.OPERATION_STATUS_FINISH {
			app.Message = "application " + transaction.OpType + " " + transaction.Status
			s.store.SaveApplication(app)
		}
		return
	}

	if app.Status == types.APP_STATUS_OPERATING {
		app.LastStatus = app.Status
		app.UpdateTime = time.Now().Unix()

		if app.Instances < app.DefineInstances {
			app.Status = types.APP_STATUS_ABNORMAL
			app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
			app.Message = "have not enough resources to launch application"
			blog.Warn("transaction %s end unnormal, set app(%s %s) to APP_STATUS_STAGING", transaction.ID, runAs, appID)
		} else {
			app.Status = types.APP_STATUS_RUNNING
			app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
			app.UpdateTime = time.Now().Unix()
			app.Message = "application " + transaction.OpType + " timeout"
		}

		err = s.store.SaveApplication(app)
		if err != nil {
			blog.Error("set application(%s.%s) status from OPERATING to RUNNING err:%s", app.RunAs, app.ID, err.Error())
		} else {
			blog.Info("set application(%s.%s) status from OPERATING to RUNNING succ!", app.RunAs, app.ID)
		}
	}

	return
}
