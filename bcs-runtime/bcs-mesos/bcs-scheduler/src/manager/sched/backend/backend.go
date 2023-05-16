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

// Package backend xxx
package backend

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/store"
)

type backend struct {
	sched *scheduler.Scheduler
	store store.Store
}

// NewBackend xxx
func NewBackend(sched *scheduler.Scheduler, zkStore store.Store) Backend {
	return &backend{
		sched: sched,
		store: zkStore,
	}
}

// ClusterId xxx
func (b *backend) ClusterId() string {
	return b.sched.ClusterId
}

// GetRole xxx
func (b *backend) GetRole() string {
	return b.sched.Role
}
