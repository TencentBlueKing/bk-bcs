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
 */

// Package client xxx
package client

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/hibiken/asynq"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/cron/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/cron/task"
)

// NewScheduler creates a new scheduler and runs it.
func NewScheduler() (*asynq.Scheduler, error) {
	blog.Info("create schedule task: cron job")
	scheduler := asynq.NewScheduler(options.NewRedisConnOpt(), nil)

	cronTasks, err := NewCronTasks()
	if err != nil {
		return nil, err
	}
	// You can use cron spec string to specify the schedule.
	for _, cronTask := range cronTasks {
		var entryID string
		entryID, err = scheduler.Register(cronTask.Cron, cronTask.Task, asynq.Queue(cronTask.QueueName))
		if err != nil {
			return nil, err
		}
		blog.Infof("registered an entry: %q\n", entryID)
	}

	return scheduler, nil
}

// NewCronTasks create cron tasks
func NewCronTasks() ([]*task.CronTask, error) {
	blog.Info("create schedule task: cron job")
	cronTasks := []*task.CronTask{}
	var err error
	var cronTask *task.CronTask
	cronTask, err = task.NewCronTask(options.TypeBcsSubnetResource, nil)
	if err != nil {
		return nil, err
	}
	cronTasks = append(cronTasks, cronTask)
	cronTask, err = task.NewCronTask(options.TypeVpcIPMonitor, nil)
	if err != nil {
		return nil, err
	}
	cronTasks = append(cronTasks, cronTask)
	cronTask, err = task.NewCronTask(options.TypeVpcOverlayNotice, nil)
	if err != nil {
		return nil, err
	}
	cronTasks = append(cronTasks, cronTask)

	return cronTasks, nil
}
