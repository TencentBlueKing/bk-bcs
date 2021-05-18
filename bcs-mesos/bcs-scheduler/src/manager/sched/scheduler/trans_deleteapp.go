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

// transaction for delete application

package scheduler

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
)

// RunDeleteApplication The goroutine function for delete application transaction
// You can create a transaction for delete application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
// this function may modify fields of transaction struct
func (s *Scheduler) RunDeleteApplication(transaction *types.Transaction) bool {
	blog.V(3).Infof("transaction %s delete application(%s.%s) try to do",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName)

	isContinue := s.doDeleteAppTrans(transaction)
	if !isContinue {
		blog.Infof("transaction %s delete application(%s.%s) finish, result(%s)",
			transaction.TransactionID, transaction.Namespace, transaction.ObjectName, transaction.Status)
		return false
	}
	blog.V(3).Infof("transaction %s delete application(%s.%s) not finish, waiting 3 seconds ...",
		transaction.TransactionID, transaction.Namespace, transaction.ObjectName)
	return true
}

// the return value indicates whether the transaction need to continue
func (s *Scheduler) doDeleteAppTrans(trans *types.Transaction) bool {
	appID := trans.ObjectName
	runAs := trans.Namespace

	blog.Infof("do transaction %s begin, to delete application(%s.%s)", trans.TransactionID, runAs, appID)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	opData := trans.CurOp.OpDeleteData
	enforce := opData.Enforce

	taskGroups, err := s.store.ListTaskGroups(runAs, appID)
	if err != nil {
		blog.Error("transaction %s list taskgroup(%s %s) err(%s)", trans.TransactionID, runAs, appID, err.Error())
		return true
	}

	for _, taskGroup := range taskGroups {
		isEnd := task.IsTaskGroupEnd(taskGroup)
		if isEnd == false {
			if task.CanTaskGroupShutdown(taskGroup) {
				blog.Info("transaction %s delete application: taskGroup(%s) not in end status, kill it",
					trans.TransactionID, taskGroup.ID)
				s.KillTaskGroup(taskGroup)
			} else {
				blog.Info("transaction %s delete application: taskGroup(%s) not in end status at current",
					trans.TransactionID, taskGroup.ID)
			}

			if enforce {
				blog.Warn("transaction %s delete application: taskGroup(%s) is not end, but enforce to delete",
					trans.TransactionID, taskGroup.ID)
			} else {
				blog.Infof("transaction %s delete application(%s.%s) pending", trans.TransactionID, runAs, appID)
				return true
			}
		} else {
			blog.Info("transaction %s delete application: taskGroup(%s) is end", trans.TransactionID, taskGroup.ID)
		}
	}

	for _, taskGroup := range taskGroups {
		blog.Info("transaction %s delete taskgroup(%s)", trans.TransactionID, taskGroup.ID)
		if err := s.DeleteTaskGroup(nil, taskGroup, "delete application"); err != nil {
			blog.Error("transaction %s delete taskgroup(%s) failed: %s",
				trans.TransactionID, taskGroup.ID, err.Error())
			return true
		}
	}

	versions, err := s.store.ListVersions(runAs, appID)
	if err != nil {
		blog.Error("transaction %s list version(%s.%s) failed: %s",
			trans.TransactionID, runAs, appID, err.Error())
	}

	if versions != nil {
		for _, version := range versions {
			if err1 := s.store.DeleteVersion(runAs, appID, version); err1 != nil {
				blog.Error("transaction %s delete version(%s.%s %s) failed: %s",
					trans.TransactionID, runAs, appID, version, err1.Error())
				return true
			}
		}
	}

	if err = s.store.DeleteVersionNode(runAs, appID); err != nil {
		blog.Error("delete app transaction %s: delete version node(%s.%s) err:%s",
			trans.TransactionID, runAs, appID, err.Error())
	}

	err = s.store.DeleteApplication(runAs, appID)
	if err != nil {
		blog.Error("transaction %s delete application(%s.%s) failed: %s",
			trans.TransactionID, runAs, appID, err.Error())
		return true
	}

	blog.Info("app transaction %s delete application(%s.%s) finish",
		trans.TransactionID, runAs, appID)
	trans.Status = types.OPERATION_STATUS_FINISH
	return false
}
