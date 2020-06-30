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
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/task"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"time"
)

// The goroutine function for delete application transaction
// You can create a transaction for delete application, then call this function to do it
// This function will come to end as soon as the transaction is done, fail or timeout(as defined by transaction.LifePeriod)
func (s *Scheduler) RunDeleteApplication(transaction *Transaction) {

	blog.Infof("transaction %s delete application(%s.%s) run begin", transaction.ID, transaction.RunAs, transaction.AppID)
	started := time.Now()
	for {

		if transaction.CreateTime+transaction.DelayTime > time.Now().Unix() {
			blog.Infof("transaction %s delete application(%s.%s) delaytime(%d), cannot do at now",
				transaction.ID, transaction.RunAs, transaction.AppID, transaction.DelayTime)
			continue
		}

		blog.V(3).Infof("transaction %s delete application(%s.%s) try to do", transaction.ID, transaction.RunAs, transaction.AppID)

		end := s.doDeleteAppTrans(transaction)
		if end {
			blog.Infof("transaction %s delete application(%s.%s) finish, result(%s)",
				transaction.ID, transaction.RunAs, transaction.AppID, transaction.Status)
			break
		}

		if (transaction.CreateTime + transaction.LifePeriod) < time.Now().Unix() {
			blog.Warnf("transaction %s delete application(%s.%s) timeout", transaction.ID, transaction.RunAs, transaction.AppID)
			transaction.Status = types.OPERATION_STATUS_TIMEOUT
			break
		}

		blog.V(3).Infof("transaction %s delete application(%s.%s) not finish, waiting 3 seconds ...",
			transaction.ID, transaction.RunAs, transaction.AppID)
		time.Sleep(3 * time.Second)
	}

	s.FinishTransaction(transaction)
	reportOperateAppMetrics(transaction.RunAs, transaction.AppID, DeleteApplicationType, started)
	blog.Infof("transaction %s delete application(%s.%s) run end, result(%s)",
		transaction.ID, transaction.RunAs, transaction.AppID, transaction.Status)
}

func (s *Scheduler) doDeleteAppTrans(trans *Transaction) bool {

	appID := trans.AppID
	runAs := trans.RunAs

	blog.Infof("do transaction %s begin, to delete application(%s.%s)", trans.ID, runAs, appID)

	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	opData := trans.OpData.(*TransAPIDeleteOpdata)
	enforce := opData.Enforce

	taskGroups, err := s.store.ListTaskGroups(runAs, appID)
	if err != nil {
		blog.Error("transaction %s list taskgroup(%s %s) err(%s)", trans.ID, runAs, appID, err.Error())
		trans.Status = types.OPERATION_STATUS_FAIL
		return true
	}

	for _, taskGroup := range taskGroups {
		isEnd := task.IsTaskGroupEnd(taskGroup)
		if isEnd == false {
			if task.CanTaskGroupShutdown(taskGroup) == true {
				blog.Info("transaction %s delete application: taskGroup(%s) not int end status, kill it",
					trans.ID, taskGroup.ID)
				s.KillTaskGroup(taskGroup)
			} else {
				blog.Info("transaction %s delete application: taskGroup(%s) not int end status at current",
					trans.ID, taskGroup.ID)
			}

			if enforce == true {
				blog.Warn("transaction %s delete application: taskGroup(%s) is not end, but enforce to delete",
					trans.ID, taskGroup.ID)
			} else {
				blog.Infof("transaction %s delete application(%s.%s) pending", trans.ID, runAs, appID)
				return false
			}
		} else {
			blog.Info("transaction %s delete application: taskGroup(%s) is end", trans.ID, taskGroup.ID)
		}
	}

	for _, taskGroup := range taskGroups {
		blog.Info("transaction %s delete taskgroup(%s)", trans.ID, taskGroup.ID)
		if err := s.DeleteTaskGroup(nil, taskGroup, "delete application"); err != nil {
			blog.Error("transaction %s delete taskgroup(%s) failed: %s", trans.ID, taskGroup.ID, err.Error())
		}
	}

	versions, err := s.store.ListVersions(runAs, appID)
	if err != nil {
		blog.Error("transaction %s list version(%s.%s) failed: %s", trans.ID, runAs, appID, err.Error())
	}

	if versions != nil {
		for _, version := range versions {
			if err1 := s.store.DeleteVersion(runAs, appID, version); err1 != nil {
				blog.Error("transaction %s delete version(%s.%s %s) failed: %s",
					trans.ID, runAs, appID, version, err1.Error())
			}
		}
	}

	if err = s.store.DeleteVersionNode(runAs, appID); err != nil {
		blog.Error("delete app transaction %s: delete version node(%s.%s) err:%s", trans.ID, runAs, appID, err.Error())
	}

	err = s.store.DeleteApplication(runAs, appID)
	if err != nil {
		blog.Error("transaction %s delete application(%s.%s) failed: %s", trans.ID, runAs, appID, err.Error())
	}

	blog.Info("app transaction %s delete application(%s.%s) finish", trans.ID, runAs, appID)
	trans.Status = types.OPERATION_STATUS_FINISH

	return true
}
