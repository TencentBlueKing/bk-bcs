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

// transaction for command

package scheduler

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// RunCommand run specified command in taskgroup
func (s *Scheduler) RunCommand(command *commtypes.BcsCommandInfo) {
	if len(command.Status.Taskgroups) == 0 {
		return
	}
	// lock command
	s.store.LockCommand(command.Id)
	defer s.store.UnLockCommand(command.Id)
	// lock application
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(command.Status.Taskgroups[0].TaskgroupId)
	s.store.LockApplication(runAs + "." + appID)
	defer s.store.UnLockApplication(runAs + "." + appID)

	blog.Info("begin send command(%s)", command.Id)
	for _, taskGroup := range command.Status.Taskgroups {
		taskGroupID := taskGroup.TaskgroupId
		taskGroupInfo, err := s.store.FetchTaskGroup(taskGroupID)
		if err != nil {
			blog.Warn("get taskgroup(%s) to do command err:%s", taskGroupID, err.Error())
			for _, task := range taskGroup.Tasks {
				task.Status = commtypes.TaskCommandStatusFailed
				task.Message = err.Error()
			}
			continue
		}

		for _, task := range taskGroup.Tasks {
			taskID := task.TaskId
			taskInfo, err := s.store.FetchTask(taskID)
			if err != nil {
				blog.Warn("get task(%s) to do command err:%s", taskID, err.Error())
				task.Status = commtypes.TaskCommandStatusFailed
				task.Message = err.Error()
				continue
			}
			if taskInfo.Status != types.TASK_STATUS_RUNNING {
				blog.Warn("task(%s) not in runnning, cannot send command", taskID)
				task.Status = commtypes.TaskCommandStatusFailed
				task.Message = "task not in running status"
				continue
			}

			msg := &types.RequestCommandTask{
				ID:         command.Id,
				TaskId:     taskID,
				Env:        command.Spec.Env,
				Cmd:        command.Spec.Command,
				User:       command.Spec.User,
				WorkingDir: command.Spec.WorkingDir,
				Privileged: command.Spec.Privileged,
			}
			bcsMsg := &types.BcsMessage{
				Type:               types.Msg_Req_COMMAND_TASK.Enum(),
				RequestCommandTask: msg,
			}
			_, err = s.SendBcsMessage(taskGroupInfo, bcsMsg)
			if err != nil {
				blog.Warn("send command to task(%s) err:%s", taskID, err.Error())
				task.Status = commtypes.TaskCommandStatusFailed
				task.Message = err.Error()
				continue
			}

			task.Status = commtypes.TaskCommandStatusRunning
			task.Message = "command in running"
			blog.Info("send command(%s) to task(%s)", command.Id, taskID)
		}
	}

	s.store.SaveCommand(command)
	blog.Info("finish send command(%s), wait for result", command.Id)
	return
}
