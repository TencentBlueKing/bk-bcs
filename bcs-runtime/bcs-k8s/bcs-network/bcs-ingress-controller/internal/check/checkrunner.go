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

// Package check 通过额外goroutine，定时检查组件运行状态
package check

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

type Interval int

const (
	CheckPerMin Interval = iota
	CheckPer10Min
)

// CheckRunner start check list
type CheckRunner struct {
	ctx           context.Context
	checkPerMin   []Checker
	checkPer10Min []Checker
}

// NewCheckRunner return new check runner
func NewCheckRunner(ctx context.Context) *CheckRunner {
	return &CheckRunner{
		ctx:           ctx,
		checkPerMin:   make([]Checker, 0),
		checkPer10Min: make([]Checker, 0),
	}
}

// Register register checker
func (c *CheckRunner) Register(checker Checker, interval Interval) *CheckRunner {
	switch interval {
	case CheckPerMin:
		c.checkPerMin = append(c.checkPerMin, checker)
	case CheckPer10Min:
		c.checkPer10Min = append(c.checkPer10Min, checker)
	default:
		c.checkPerMin = append(c.checkPerMin, checker)
	}
	return c
}

// Start 定时启动注册的所有checker
func (c *CheckRunner) Start() {
	go func() {
		tickerMin := time.NewTicker(time.Minute)
		tickerMin10 := time.NewTicker(time.Minute * 10)
		defer func() {
			tickerMin.Stop()
			tickerMin10.Stop()
		}()
		for {
			select {
			case <-tickerMin.C:
				for _, item := range c.checkPerMin {
					go item.Run()
				}
			case <-tickerMin10.C:
				for _, item := range c.checkPer10Min {
					go item.Run()
				}
			case <-c.ctx.Done():
				blog.Infof("Stop run checker")
				return
			}
		}
	}()
}
