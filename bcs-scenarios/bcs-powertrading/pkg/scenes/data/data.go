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

// Package data for operation data
package data

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes/data/handler"
)

// Controller data controller
type Controller struct {
	opt           *scenes.Options
	deviceHandler *handler.DeviceDataHandler
}

// NewDataController new controller
func NewDataController(opt *scenes.Options) scenes.Controller {
	return &Controller{opt: opt}
}

// Init init controller
func (c *Controller) Init(opts ...scenes.Option) error {
	deviceDataHandler := &handler.DeviceDataHandler{
		Opts: c.opt,
	}
	c.deviceHandler = deviceDataHandler
	deviceDataHandler.Init()
	return nil
}

// Options Controller implementation
func (c *Controller) Options() *scenes.Options {
	return c.opt
}

// Run Controller implementation
func (c *Controller) Run(ctx context.Context) {
	go c.deviceHandler.Run(ctx)
}
