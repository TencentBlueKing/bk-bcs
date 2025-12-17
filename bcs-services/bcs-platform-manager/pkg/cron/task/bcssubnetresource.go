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
	"context"
	"log"

	"github.com/hibiken/asynq"
)

// HandleBcsSubnetResourceTask : handle bcs subnet resource task
func HandleBcsSubnetResourceTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("handle bcs subnet resource task: %s", t.Payload())
	// Email delivery code ...
	return nil
}

// HandleVpcIPMonitorTask : handle vpc ip monitor task
func HandleVpcIPMonitorTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("handle vpc ip monitor task: %s", t.Payload())
	// Email delivery code ...
	return nil
}

// HandleVpcOverlayNoticeTask : handle vpc overlay notice task
func HandleVpcOverlayNoticeTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("handle vpc overlay notice task: %s", t.Payload())
	// Email delivery code ...
	return nil
}
