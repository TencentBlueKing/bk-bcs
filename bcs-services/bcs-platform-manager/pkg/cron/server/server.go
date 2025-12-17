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

// NewAsynqServer creates a new server and runs it.
func NewAsynqServer() error {
	blog.Info("start a asynq server")
	srv := asynq.NewServer(
		options.NewRedisConnOpt(),
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				options.BcsSubnetResourceQueueName: 1,
				options.VpcIPMonitorQueueName:      1,
				options.VpcOverlayNoticeQueueName:  1,
			},
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(options.TypeBcsSubnetResource, task.HandleBcsSubnetResourceTask)
	mux.HandleFunc(options.TypeVpcIPMonitor, task.HandleVpcIPMonitorTask)
	mux.HandleFunc(options.TypeVpcOverlayNotice, task.HandleVpcOverlayNoticeTask)
	// ...register other handlers...

	go func() {
		err := srv.Run(mux)
		if err != nil {
			panic("run asynq server failed: " + err.Error())
		}
	}()

	return nil
}
