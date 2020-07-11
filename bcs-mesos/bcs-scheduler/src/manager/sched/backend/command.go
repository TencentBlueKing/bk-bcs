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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"strings"
)

func (b *backend) GetCommand(ID string) (*commtypes.BcsCommandInfo, error) {
	return b.store.FetchCommand(ID)
}

func (b *backend) DeleteCommand(ID string) error {
	return b.store.DeleteCommand(ID)
}

func (b *backend) DoCommand(command *commtypes.BcsCommandInfo) error {

	kind := command.Spec.CommandTargetRef.Kind
	name := command.Spec.CommandTargetRef.Name
	ns := command.Spec.CommandTargetRef.Namespace

	if len(command.Spec.Taskgroups) > 0 {
		for _, taskgroupID := range command.Spec.Taskgroups {
			taskGroup, err := b.store.FetchTaskGroup(taskgroupID)
			if err != nil {
				blog.Errorf("fetch taskgroup(%s) error:%s, send command fail", taskgroupID, err.Error())
				return err
			}
			var taskgroupCommand commtypes.TaskgroupCommandInfo
			taskgroupCommand.TaskgroupId = taskGroup.ID
			for _, task := range taskGroup.Taskgroup {
				if command.Spec.CommandTargetRef.Image != "" && !strings.Contains(task.Image, command.Spec.CommandTargetRef.Image) {
					blog.Infof("command %s task %s image don't match %s, and continue",
						command.Id, task.ID, command.Spec.CommandTargetRef.Image)
					continue
				}

				var taskCommand commtypes.TaskCommandInfo
				taskCommand.TaskId = task.ID
				taskCommand.Status = commtypes.TaskCommandStatusStaging
				taskCommand.Message = "still not send command to task"
				taskCommand.CommInspect = nil
				taskgroupCommand.Tasks = append(taskgroupCommand.Tasks, &taskCommand)
			}
			command.Status.Taskgroups = append(command.Status.Taskgroups, &taskgroupCommand)
		}
	} else if kind != "Deployment" {
		appName := name
		taskGroups, err := b.store.ListTaskGroups(ns, appName)
		if err != nil {
			blog.Error("list taskgroup(%s.%s) error:%s, send command fail", ns, appName, err.Error())
			return err
		}
		for _, taskGroup := range taskGroups {
			var taskgroupCommand commtypes.TaskgroupCommandInfo
			taskgroupCommand.TaskgroupId = taskGroup.ID
			for _, task := range taskGroup.Taskgroup {
				if command.Spec.CommandTargetRef.Image != "" && !strings.Contains(task.Image, command.Spec.CommandTargetRef.Image) {
					blog.Infof("command %s task %s image don't match %s, and continue",
						command.Id, task.ID, command.Spec.CommandTargetRef.Image)
					continue
				}

				var taskCommand commtypes.TaskCommandInfo
				taskCommand.TaskId = task.ID
				taskCommand.Status = commtypes.TaskCommandStatusStaging
				taskCommand.Message = "still not send command to task"
				taskCommand.CommInspect = nil
				taskgroupCommand.Tasks = append(taskgroupCommand.Tasks, &taskCommand)
			}
			command.Status.Taskgroups = append(command.Status.Taskgroups, &taskgroupCommand)
		}
	} else {
		err := b.getDepoymentTaskgroup(command)
		if err != nil {
			return err
		}
	}

	if err := b.store.SaveCommand(command); err != nil {
		blog.Error("save command(%s) err:%s", command.Id, err.Error())
		return err
	}
	// go do
	go b.sched.RunCommand(command)

	return nil
}

func (b *backend) getDepoymentTaskgroup(command *commtypes.BcsCommandInfo) error {

	name := command.Spec.CommandTargetRef.Name
	ns := command.Spec.CommandTargetRef.Namespace

	deployment, err := b.store.FetchDeployment(ns, name)
	if err != nil {
		blog.Error("send command to deployment(%s.%s), fetch deployment err:%s", ns, name, err.Error())
		err = fmt.Errorf("fetch deployment(%s.%s) err: %s", ns, name, err.Error())
		return err
	}

	if deployment != nil && deployment.Application != nil {
		appName := deployment.Application.ApplicationName
		taskGroups, err := b.store.ListTaskGroups(ns, appName)
		if err != nil {
			blog.Error("list taskgroup(%s.%s) error:%s, send command fail", ns, appName, err.Error())
			return err
		}
		for _, taskGroup := range taskGroups {
			var taskgroupCommand commtypes.TaskgroupCommandInfo
			taskgroupCommand.TaskgroupId = taskGroup.ID
			for _, task := range taskGroup.Taskgroup {
				if command.Spec.CommandTargetRef.Image != "" && !strings.Contains(task.Image, command.Spec.CommandTargetRef.Image) {
					blog.Infof("command %s task %s image don't match %s, and continue",
						command.Id, task.ID, command.Spec.CommandTargetRef.Image)
					continue
				}

				var taskCommand commtypes.TaskCommandInfo
				taskCommand.TaskId = task.ID
				taskCommand.Status = commtypes.TaskCommandStatusStaging
				taskCommand.Message = "still not send command to task"
				taskCommand.CommInspect = nil
				taskgroupCommand.Tasks = append(taskgroupCommand.Tasks, &taskCommand)
			}
			command.Status.Taskgroups = append(command.Status.Taskgroups, &taskgroupCommand)
		}
	}

	if deployment != nil && deployment.ApplicationExt != nil {
		appName := deployment.ApplicationExt.ApplicationName
		taskGroups, err := b.store.ListTaskGroups(ns, appName)
		if err != nil {
			blog.Error("list taskgroup(%s.%s) error:%s, send command fail", ns, appName, err.Error())
			return err
		}
		for _, taskGroup := range taskGroups {
			var taskgroupCommand commtypes.TaskgroupCommandInfo
			taskgroupCommand.TaskgroupId = taskGroup.ID
			for _, task := range taskGroup.Taskgroup {
				if command.Spec.CommandTargetRef.Image != "" && !strings.Contains(task.Image, command.Spec.CommandTargetRef.Image) {
					blog.Infof("command %s task %s image don't match %s, and continue",
						command.Id, task.ID, command.Spec.CommandTargetRef.Image)
					continue
				}

				var taskCommand commtypes.TaskCommandInfo
				taskCommand.TaskId = task.ID
				taskCommand.Status = commtypes.TaskCommandStatusStaging
				taskCommand.Message = "still not send command to task"
				taskCommand.CommInspect = nil
				taskgroupCommand.Tasks = append(taskgroupCommand.Tasks, &taskCommand)
			}
			command.Status.Taskgroups = append(command.Status.Taskgroups, &taskgroupCommand)
		}
	}

	return nil
}
