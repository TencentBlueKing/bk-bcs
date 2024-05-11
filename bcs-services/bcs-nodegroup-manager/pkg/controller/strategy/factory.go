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

package strategy

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster"
	mgr "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	bufferExecutor               = &BufferStrategyExecutor{}
	hierarchicalStrategyExecutor = &HierarchicalStrategyExecutor{}
)

// Factory strategy factory
type Factory interface {
	// Init init different kind of strategies
	Init()
	// GetStrategyExecutor get strategy executor by kind
	GetStrategyExecutor(strategy *storage.Strategy) (Executor, error)
}

type strategyFactory struct {
	opt *Options
}

// Options option for strategy factory
type Options struct {
	// resource manager interface for data retrieve
	ResourceManager mgr.Client
	// storage for access database
	Storage storage.Storage
	// client for cluster request
	ClusterClient cluster.Client
}

// NewFactory new strategy factory
func NewFactory(opt *Options) Factory {
	return &strategyFactory{
		opt: opt,
	}
}

// Init init different kind of strategies
func (f *strategyFactory) Init() {
	bufferExecutor = NewBufferStrategyExecutor(f.opt)
	hierarchicalStrategyExecutor = NewHierarchicalStrategyExecutor(f.opt)
}

// GetStrategyExecutor get strategy executor by kind
func (f *strategyFactory) GetStrategyExecutor(strategy *storage.Strategy) (Executor, error) {
	switch strategy.Type {
	case storage.BufferStrategyType:
		return bufferExecutor, nil
	case storage.HierarchicalStrategyType:
		return hierarchicalStrategyExecutor, nil
	default:
		return nil, fmt.Errorf("unknown strategy type:%s", strategy.Type)
	}
}

// Executor interface of strategy executor, implement by different kind of executors
type Executor interface {
	IsAbleToScaleDown(strategy *storage.NodeGroupMgrStrategy) (int, bool, error)
	IsAbleToScaleUp(strategy *storage.NodeGroupMgrStrategy) (int, bool, int, error)
	HandleNodeMetadata()
	CreateNodeUpdateAction(strategy *storage.NodeGroupMgrStrategy, action *storage.NodeGroupAction) error
}

func checkIfTaskExecuting(strategyName string, storageCli storage.Storage) (bool, error) {
	tasks, err := storageCli.ListTasksByStrategy(strategyName, &storage.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("list strategy %s tasks err:%s", strategyName, err.Error())
	}
	for _, task := range tasks {
		if time.Now().Add(5 * time.Minute).After(task.BeginExecuteTime) {
			return true, nil
		}
		if task.IsExecuting() {
			return true, nil
		}
	}
	return false, nil
}
