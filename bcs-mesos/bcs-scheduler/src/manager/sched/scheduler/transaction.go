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
	"errors"
	"fmt"
	"time"

	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

const (
	defaultTransactionCheckInterval = 3 * time.Second
)

// TRANSACTION_DEFAULT_LIFEPERIOD Default transaction max lifeperoid, 480 seconds
// the lifeperoid for all not specified transactions are set to this
const TRANSACTION_DEFAULT_LIFEPERIOD = 480

const TRANSACTION_APPLICATION_LAUNCH_LIFEPERIOD = 1800

// TRANSACTION_DEPLOYMENT_ROLLING_UP_LIFEPERIOD Max lifeperoid for every rolling transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_ROLLING_UP_LIFEPERIOD = 300

// TRANSACTION_DEPLOYMENT_ROLLING_DOWN_LIFEPERIOD Max lifeperoid for every rolling transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_ROLLING_DOWN_LIFEPERIOD = 7500

// TRANSACTION_DEPLOYMENT_INNERDELETE_LIFEPERIOD Max lifeperoid for innder delete application transaction, 300 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_DEPLOYMENT_INNERDELETE_LIFEPERIOD = 7500

// TRANSACTION_INNER_RESCHEDULE_LIFEPERIOD Max lifePeriod for inner taskgroup-reschedule, 3600 seconds
// If a transaction dosen't finish in its lifePeriod, it will be timeout
const TRANSACTION_INNER_RESCHEDULE_LIFEPERIOD = 86400

// TRANSACTION_RESCHEDULE_RESET_INTERVAL If taskgroup running than 1800 seconds, the restart times will be reset to 0
const TRANSACTION_RESCHEDULE_RESET_INTERVAL = 1800

// updateTransactionApp Finish a transaction, set application status
func (s *Scheduler) updateTransactionApp(transaction *types.Transaction) error {
	blog.Infof("transaction %s type %s ns %s objectName %s status %s end",
		transaction.TransactionID, transaction.CurOp.OpType, transaction.Namespace,
		transaction.ObjectName, transaction.Status)
	s.store.LockApplication(transaction.Namespace + "." + transaction.ObjectName)
	defer s.store.UnLockApplication(transaction.Namespace + "." + transaction.ObjectName)

	if transaction.CurOp.OpType == types.TransactionOpTypeDelete {
		return nil
	}
	app, err := s.store.FetchApplication(transaction.Namespace, transaction.ObjectName)
	if err != nil {
		if errors.Is(err, store.ErrNoFound) {
			blog.Infof("transaction %s finished, application(%s.%s) not found, no need update app",
				transaction.TransactionID, transaction.Namespace, transaction.ObjectName)
			return nil
		}
		return fmt.Errorf("transaction %s finished, fetch application(%s.%s), err %s",
			transaction.TransactionID, transaction.Namespace, transaction.ObjectName, err.Error())
	}

	if transaction.Status == types.OPERATION_STATUS_TIMEOUT {
		s.SendHealthMsg(alarm.WarnKind, app.RunAs,
			"transaction("+transaction.TransactionID+") timeout: "+
				transaction.CurOp.OpType+" "+transaction.Namespace+"."+transaction.ObjectName, "", nil)
	}

	if transaction.CurOp.OpType == types.TransactionOpTypeInnerScale {
		if transaction.Status != types.OPERATION_STATUS_FINISH {
			app.Message = "application " + transaction.CurOp.OpType + " " + transaction.Status
			return s.store.SaveApplication(app)
		}
	}

	if app.Status == types.APP_STATUS_OPERATING {
		app.LastStatus = app.Status
		app.UpdateTime = time.Now().Unix()

		if app.Instances < app.DefineInstances {
			app.Status = types.APP_STATUS_ABNORMAL
			app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
			if transaction.Status == types.OPERATION_STATUS_FAIL {
				app.Message = transaction.Message
			} else {
				app.Message = "have not enough resources to launch application"
			}
			blog.Warnf("transaction %s end abnormal, set app(%s %s) to APP_STATUS_ABNORMAL",
				transaction.TransactionID, transaction.Namespace, transaction.ObjectName)
		} else {
			app.Status = types.APP_STATUS_RUNNING
			app.SubStatus = types.APP_SUBSTATUS_UNKNOWN
			app.UpdateTime = time.Now().Unix()
			if transaction.Status == types.OPERATION_STATUS_FAIL {
				app.Message = transaction.Message
			} else if transaction.Status == types.OPERATION_STATUS_TIMEOUT {
				app.Message = "application " + transaction.CurOp.OpType + " timeout"
			} else {
				app.Message = types.APP_STATUS_RUNNING_STR
			}
		}

		err = s.store.SaveApplication(app)
		if err != nil {
			return fmt.Errorf("set application(%s.%s) status from OPERATING to RUNNING err:%s",
				app.RunAs, app.ID, err.Error())
		} else {
			blog.Info("set application(%s.%s) status from OPERATING to RUNNING succ!", app.RunAs, app.ID)
		}
	}
	return nil
}

func (s *Scheduler) updateTransactionDeployment(transaction *types.Transaction) error {
	blog.Infof("transaction %s type %s ns %s objectKind %s objectName %s status %s end",
		transaction.TransactionID, transaction.CurOp.OpType, transaction.Namespace,
		transaction.ObjectKind, transaction.ObjectName, transaction.Status)
	s.store.LockDeployment(transaction.Namespace + "." + transaction.ObjectName)
	defer s.store.UnLockDeployment(transaction.Namespace + "." + transaction.ObjectName)

	if transaction.CurOp.OpType == types.TransactionOpTypeDepUpdateResource {
		name := transaction.ObjectName
		ns := transaction.Namespace
		currDeployment, err := s.store.FetchDeployment(ns, name)
		if err != nil {
			return fmt.Errorf("get deployment(%s.%s) failed when finish transaction, err %s",
				ns, name, err.Error())
		}
		appName := currDeployment.Application.ApplicationName
		currApp, err := s.store.FetchApplication(ns, appName)
		if err != nil {
			return fmt.Errorf("get application(%s.%s) failed when UpdateDeploymentResource, err %s",
				ns, appName, err.Error())
		}
		if currApp.Message != types.APP_STATUS_RUNNING_STR {
			currDeployment.Message = currApp.Message
		}
		currDeployment.Status = types.DEPLOYMENT_STATUS_RUNNING
		if err := s.store.SaveDeployment(currDeployment); err != nil {
			return fmt.Errorf("update deployment(%s.%s) status failed when finish transaction, err %s",
				ns, name, err.Error())
		}
	}
	return nil
}

func (s *Scheduler) transactionLoop() {
	blog.Infof("enter transaction loop")
	for s.handleTransaction() {
	}
	blog.Infof("exit transaction loop")
}

func (s *Scheduler) handleLaunchTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpLaunchData == nil {
		blog.Warnf("handle launch transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}
	isContinue := s.RunLaunchApplication(transaction)
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	if err := s.updateTransactionApp(transaction); err != nil {
		blog.Warnf("update transaction application failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	reportOperateAppMetrics(transaction.Namespace, transaction.ObjectName,
		LaunchApplicationType, transaction.CreateTime)
	blog.Infof("transaction %s launch application(%s.%s) run end, result(%s)",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
}

func (s *Scheduler) handleDeleteTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpDeleteData == nil {
		blog.Warnf("handle delete transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}

	isContinue := s.RunDeleteApplication(transaction)
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	if err := s.updateTransactionApp(transaction); err != nil {
		blog.Warnf("update transaction application failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	reportOperateAppMetrics(transaction.Namespace, transaction.ObjectName,
		DeleteApplicationType, transaction.CreateTime)
	blog.Infof("transaction %s delete application(%s.%s) run end, result(%s)",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
}

func (s *Scheduler) handleScaleTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpScaleData == nil {
		blog.Warnf("handle scale transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}

	var isContinue bool
	if transaction.CurOp.OpScaleData.IsInner {
		isContinue = s.RunInnerScaleApplication(transaction)
	} else {
		isContinue = s.RunScaleApplication(transaction)
	}
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	if err := s.updateTransactionApp(transaction); err != nil {
		blog.Warnf("update transaction application failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// only report metric when the scale operation is not inner
	if !transaction.CurOp.OpScaleData.IsInner {
		reportOperateAppMetrics(transaction.Namespace, transaction.ObjectName,
			ScaleApplicationType, transaction.CreateTime)
	}
	blog.Infof("transaction %s scale application(%s.%s) run end, result(%s)",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
}

func (s *Scheduler) handleUpdateTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpUpdateData == nil {
		blog.Warnf("handle update transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}

	var isContinue bool
	if transaction.CurOp.OpUpdateData.IsUpdateResource {
		isContinue = s.RunUpdateApplicationResource(transaction)
	} else {
		isContinue = s.RunUpdateApplication(transaction)
	}
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	if err := s.updateTransactionApp(transaction); err != nil {
		blog.Warnf("update transaction application failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	reportOperateAppMetrics(transaction.Namespace, transaction.ObjectName,
		UpdateApplicationType, transaction.CreateTime)
	blog.Infof("transaction %s update(%s.%s) run end, result(%s)",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
}

func (s *Scheduler) handleRescheduleTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpRescheduleData == nil {
		blog.Warnf("handle reschedule transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}

	isContinue := s.RunRescheduleTaskgroup(transaction)
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	if err := s.updateTransactionApp(transaction); err != nil {
		blog.Warnf("update transaction application failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	blog.Infof("transaction %s reschedule(%s) run end, result(%s)",
		transaction.TransactionID, transaction.CurOp.OpRescheduleData.TaskGroupID, transaction.Status)
}

func (s *Scheduler) handleDepUpdateResourceTransaction(obj k8stypes.NamespacedName, transaction *types.Transaction) {
	if transaction.CurOp == nil || transaction.CurOp.OpDepUpdateData == nil {
		blog.Warnf("handle deployment update resource transaction %s lost Op data", obj.String())
		s.transactionQueue.Forget(obj)
		// delete transaction
		if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
			blog.Warnf("delete transaction failed, err %s", err.Error())
			s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
			return
		}
		return
	}

	isContinue := s.RunUpdateDeploymentResource(transaction)
	if isContinue {
		s.transactionQueue.Forget(obj)
		// update transaction
		if err := s.store.SaveTransaction(transaction); err != nil {
			blog.Warnf("update transaction store failed, err %s", err.Error())
		}
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	s.transactionQueue.Forget(obj)
	if err := s.updateTransactionDeployment(transaction); err != nil {
		blog.Warnf("update transaction deployment failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	// delete transaction
	if err := s.store.DeleteTransaction(transaction.Namespace, transaction.TransactionID); err != nil {
		blog.Warnf("delete transaction failed, err %s", err.Error())
		s.transactionQueue.AddAfter(obj, transaction.CheckInterval)
		return
	}
	blog.Infof("transaction %s update resource of deployment(%s.%s) run end, result(%s)",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
}

func (s *Scheduler) handleTransaction() bool {
	obj, shutDown := s.transactionQueue.Get()
	if shutDown {
		blog.Warnf("transaction queue was shut down")
		return false
	}
	defer s.transactionQueue.Done(obj)

	nsName, ok := obj.(k8stypes.NamespacedName)
	if !ok {
		blog.Warnf("invalid queue item, type %T, value %v, forget it", obj, obj)
		s.transactionQueue.Forget(obj)
		return true
	}
	transaction, err := s.store.FetchTransaction(nsName.Namespace, nsName.Name)
	if err != nil {
		if errors.Is(err, store.ErrNoFound) {
			blog.V(4).Infof("transaction %s not found in store, think it has been completed", nsName.String())
			s.transactionQueue.Forget(obj)
			return true
		}
		blog.Errorf("get transaction %s from store failed, err %s", nsName.String(), err.Error())
		s.transactionQueue.Forget(obj)
		s.transactionQueue.AddAfter(obj, defaultTransactionCheckInterval)
		return true
	}

	if transaction.CurOp == nil {
		blog.Errorf("transaction missing current operation field, %v", transaction)
		s.transactionQueue.Forget(obj)
		return true
	}

	blog.Infof("transaction %s trigger, object kind %s, object name %s, namespaces %s",
		transaction.TransactionID, transaction.ObjectKind, transaction.ObjectName, transaction.Namespace)

	switch transaction.CurOp.OpType {
	case types.TransactionOpTypeLaunch:
		go s.handleLaunchTransaction(nsName, transaction)
		return true
	case types.TransactionOpTypeDelete:
		go s.handleDeleteTransaction(nsName, transaction)
		return true
	case types.TransactionOpTypeScale, types.TransactionOpTypeInnerScale:
		go s.handleScaleTransaction(nsName, transaction)
		return true
	case types.TransactionOpTypeUpdate:
		go s.handleUpdateTransaction(nsName, transaction)
		return true
	case types.TransactionOpTypeReschedule:
		go s.handleRescheduleTransaction(nsName, transaction)
		return true
	case types.TransactionOpTypeDepUpdateResource:
		go s.handleDepUpdateResourceTransaction(nsName, transaction)
		return true
	default:
		blog.Errorf("invalide transaction op type %s, transaction info %v", transaction.CurOp.OpType, transaction)
		s.transactionQueue.Forget(obj)
		return true
	}
}

func (s *Scheduler) startTransactionLoop() error {
	blog.Infof("start transaction loop")
	transList, err := s.store.ListAllTransaction()
	if err != nil {
		return err
	}
	s.transactionQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	go s.transactionLoop()

	for _, trans := range transList {
		blog.V(4).Infof("push online transaction %s, object kind %s, object name %s, ns %s",
			trans.TransactionID, trans.ObjectKind, trans.ObjectName, trans.Namespace)
		s.PushEventQueue(trans)
	}
	return nil
}

func (s *Scheduler) stopTransactionLoop() {
	if s.transactionQueue != nil {
		blog.Infof("stop transaction loop")
		s.transactionQueue.ShutDown()
		s.transactionQueue = nil
	}

}
