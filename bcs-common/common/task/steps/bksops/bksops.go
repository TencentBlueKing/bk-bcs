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

// Package bksops defines the step implemented.
package bksops

import (
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
)

type sops struct {
	bkAppCode   string
	bkAppSecret string
}

func newSops(bkAppCode, bkAppSecret string) *sops {
	return &sops{bkAppCode, bkAppSecret}
}

func (s *sops) Execute(c *istep.Context) error {
	return nil
}

// 注意, Step名称不能修改
const (
	// BKSopsStep ...
	BKSopsStep istep.StepName = "BK_SOPS"
)

// Register ...
func Register(bkAppCode, bkAppSecret string) {
	s := newSops(bkAppCode, bkAppSecret)

	istep.Register(BKSopsStep, istep.StepExecutor(s))
}
