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

package upgrader

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

// UpgradeHelper is a helper for upgrade
type UpgradeHelper interface {
	// HelperName return the name of the helper
	HelperName() string
}

// Helper is an implementation for interface UpgradeHelper
type Helper struct {
	DB drivers.DB
	// TODO: 添加clusterManagerClient调用bcs-cluster-manager提供的接口，添加httpClient去调用bcs-saas cc模块提供的接口
}

// HelperOpt is option for Helper
type HelperOpt struct {
	DB drivers.DB
}

// Name is the method of Helper to implement interface UpgradeHelper
func (h *Helper) HelperName() string {
	return "bcs-upgrade-helper"
}

// NewUpgradeHelper new a Helper instance
func NewUpgradeHelper(opt *HelperOpt) *Helper {
	return &Helper{
		DB: opt.DB,
	}
}
