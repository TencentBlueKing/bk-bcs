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

// CheckRunner start check list
type CheckRunner struct {
	ctx       context.Context
	checkList []Checker
}

// NewCheckRunner return new check runner
func NewCheckRunner(ctx context.Context) *CheckRunner {
	return &CheckRunner{
		ctx:       ctx,
		checkList: make([]Checker, 0),
	}
}

// Register register checker
func (c *CheckRunner) Register(checker Checker) *CheckRunner {
	c.checkList = append(c.checkList, checker)
	return c
}

// Start 定时启动注册的所有checker
func (c *CheckRunner) Start() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, item := range c.checkList {
					go item.Run()
				}
			case <-c.ctx.Done():
				blog.Infof("Stop run checker")
				return
			}
		}
	}()
}
