/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"fmt"
	"sync"

	"k8s.io/klog/v2"
)

// DefaultExecutorRegistry implements ExecutorRegistry
type DefaultExecutorRegistry struct {
	executors map[string]ActionExecutor
	mu        sync.RWMutex
}

// NewExecutorRegistry creates a new executor registry
func NewExecutorRegistry() *DefaultExecutorRegistry {
	return &DefaultExecutorRegistry{
		executors: make(map[string]ActionExecutor),
	}
}

// RegisterExecutor registers an action executor
func (r *DefaultExecutorRegistry) RegisterExecutor(executor ActionExecutor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	actionType := executor.Type()
	if _, exists := r.executors[actionType]; exists {
		return fmt.Errorf("executor for action type %s already registered", actionType)
	}

	r.executors[actionType] = executor
	klog.Infof("Registered executor for action type: %s", actionType)
	return nil
}

// GetExecutor returns an executor for the given action type
func (r *DefaultExecutorRegistry) GetExecutor(actionType string) (ActionExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[actionType]
	if !exists {
		return nil, fmt.Errorf("no executor registered for action type: %s", actionType)
	}

	return executor, nil
}

// ListExecutors returns all registered executors
func (r *DefaultExecutorRegistry) ListExecutors() []ActionExecutor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executors := make([]ActionExecutor, 0, len(r.executors))
	for _, executor := range r.executors {
		executors = append(executors, executor)
	}

	return executors
}
