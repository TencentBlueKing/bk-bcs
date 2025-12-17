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

// Package task xxx
package task

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/cron/options"
)

// CronTask : cron task
type CronTask struct {
	Task      *asynq.Task
	Cron      string
	QueueName string
}

// NewCronTask : create cron task
func NewCronTask(taskType string, payload []byte) (*CronTask, error) {
	cronTask := &CronTask{}
	switch taskType {
	case options.TypeBcsSubnetResource:
		cronTask.Task = asynq.NewTask(options.TypeBcsSubnetResource, payload, asynq.Unique(60*time.Minute))
		cronTask.Cron = config.G.Cron.BcsSubnetResourceCron
		cronTask.QueueName = options.BcsSubnetResourceQueueName
		return cronTask, nil
	case options.TypeVpcIPMonitor:
		cronTask.Task = asynq.NewTask(options.TypeVpcIPMonitor, payload, asynq.Unique(60*time.Minute))
		cronTask.Cron = config.G.Cron.VpcIPMonitorCron
		cronTask.QueueName = options.VpcIPMonitorQueueName
		return cronTask, nil
	case options.TypeVpcOverlayNotice:
		cronTask.Task = asynq.NewTask(options.TypeVpcOverlayNotice, payload, asynq.Unique(60*time.Minute))
		cronTask.Cron = config.G.Cron.VpcOverlayNoticeCron
		cronTask.QueueName = options.VpcOverlayNoticeQueueName
		return cronTask, nil
	default:
		return nil, fmt.Errorf("unknown task type: %s", taskType)
	}
}
