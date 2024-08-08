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

// Package hello defines the hello step.
package hello

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
)

// hello hello
type hello struct{}

// NewHello ...
func NewHello() iface.StepWorkerInterface {
	return &hello{}
}

// DoWork for worker exec task
func (s *hello) DoWork(ctx context.Context, work *istep.Work) error {
	fmt.Println("Hello")
	// time.Sleep(30 * time.Second)
	if err := work.AddCommonParams("name", "hello"); err != nil {
		return err
	}
	return nil
}

func init() {
	// 使用结构体注册
	istep.Register("hello", NewHello())

	// 使用函数注册
	istep.Register("sum", istep.StepWorkerFunc(Sum))
}
