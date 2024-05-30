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

// Package supply xxx
package supply

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes/supply/taskhandler"
)

// TaskController task controller
type TaskController struct {
	opt                 *scenes.Options
	machineTestHandler  *taskhandler.MachineTestHandler
	machineCleanHandler *taskhandler.MachineCleanHandler
}

// NewTaskController new controller
func NewTaskController(opt *scenes.Options) scenes.Controller {
	return &TaskController{opt: opt}
}

// Init init controller
func (c *TaskController) Init(opts ...scenes.Option) error {
	machineTestHandler := &taskhandler.MachineTestHandler{
		Storage:     c.opt.Storage,
		BksopsCli:   c.opt.BKsopsCli,
		JobCli:      c.opt.JobCli,
		CrCli:       c.opt.BkCrCli,
		CcCli:       c.opt.BkccCli,
		Interval:    c.opt.Interval,
		Concurrency: c.opt.Concurrency,
		ClusterCli:  c.opt.ClusterMgrCli,
		ResourceCli: c.opt.ResourceMgrCli,
	}
	machineCleanHandler := &taskhandler.MachineCleanHandler{
		Storage:     c.opt.Storage,
		BksopsCli:   c.opt.BKsopsCli,
		JobCli:      c.opt.JobCli,
		Interval:    c.opt.Interval,
		Concurrency: c.opt.Concurrency,
	}
	c.machineTestHandler = machineTestHandler
	c.machineCleanHandler = machineCleanHandler
	machineTestHandler.Init()
	machineCleanHandler.Init()
	return nil
}

// Options NodeGroupController implementation
func (c *TaskController) Options() *scenes.Options {
	return c.opt
}

// Run NodeGroupController implementation
func (c *TaskController) Run(ctx context.Context) {
	go c.machineTestHandler.Run(ctx)
	go c.machineCleanHandler.Run(ctx)
}
