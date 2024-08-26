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

// Package mem implemented storage interface.
package mem

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type memStore struct {
	mtx   sync.Mutex
	tasks map[string]*types.Task
}

// New new memStore
func New() iface.Store {
	s := &memStore{
		tasks: make(map[string]*types.Task),
	}
	return s
}

// EnsureTable 创建db表
func (s *memStore) EnsureTable(ctx context.Context, dst ...any) error {
	return nil
}

func (s *memStore) CreateTask(ctx context.Context, task *types.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.tasks[task.GetTaskID()] = task
	return nil
}

func (s *memStore) ListTask(ctx context.Context, opt *iface.ListOption) ([]types.Task, error) {
	return nil, types.ErrNotImplemented
}

func (s *memStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return types.ErrNotImplemented
}

func (s *memStore) DeleteTask(ctx context.Context, taskID string) error {
	return types.ErrNotImplemented
}

func (s *memStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	t, ok := s.tasks[taskID]
	if ok {
		return t, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *memStore) PatchTask(ctx context.Context, opt *iface.PatchOption) error {
	return types.ErrNotImplemented
}
