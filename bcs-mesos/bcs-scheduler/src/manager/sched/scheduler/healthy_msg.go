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
	alarm "github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// HealthyReport The goroutine function for process health check report
// When scheduler receive health-check report messege, it will create a goroutine for process this message,
func (s *Scheduler) HealthyReport(healthyResult *bcstype.HealthCheckResult) {

	taskID := healthyResult.ID
	healthy := healthyResult.Status
	message := healthyResult.Message
	checkType := healthyResult.Type

	blog.Info("healthy report: task(%s) healthy(%t) message(%s)", taskID, healthy, message)

	taskGroupID := types.GetTaskGroupID(taskID)
	if taskGroupID == "" {
		blog.Error("healthy report: can not get taskGroupId from taskID %s", taskID)
		return
	}
	taskGroup, err := s.store.FetchTaskGroup(taskGroupID)
	if err != nil {
		blog.Error("healthy report: Fetch taskgroup %s failed: %s", taskGroupID, err.Error())
		return
	}
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	task, err := s.store.FetchTask(taskID)
	if task == nil {
		blog.Warn("healthy report: fetch task %s return nil", taskID)
		return
	}

	if task.Status != types.TASK_STATUS_RUNNING {
		blog.Info("healthy report: task %s status %s, ignore", taskID, task.Status)
		return
	}

	for _, healthStatus := range task.HealthCheckStatus {
		if healthStatus.Type == checkType {
			if healthStatus.Result != healthy {
				blog.Infof("healthy report: Task(%s) running, remote healthy change to %t(%s)", taskID, healthy, message)
				if healthy == false {
					s.SendHealthMsg(alarm.WarnKind, taskGroup.RunAs, task.ID+"("+taskGroup.HostName+")"+" remote healthy false: "+message, "", nil)
				} else {
					s.SendHealthMsg(alarm.InfoKind, taskGroup.RunAs, task.ID+"("+taskGroup.HostName+")"+" remote healthy true: "+message, "", nil)
				}
			}

			if healthStatus.Result != healthy || healthStatus.Message != message {
				healthStatus.Result = healthy
				healthStatus.Message = message
				blog.Infof("healthy report: Task(%s) remote healthy changed to %t:%s ", taskID, healthy, message)
				s.store.SaveTask(task)
			}
			break
		}
	}

	return
}
