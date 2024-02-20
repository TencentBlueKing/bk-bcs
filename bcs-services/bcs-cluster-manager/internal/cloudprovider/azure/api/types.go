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

package api

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

const (
	// winTypeOS windows os
	winTypeOS = "win"
	// kubeletType  kubelet
	kubeletType = "kubelet"

	// NormalState 成功
	NormalState = "Succeeded"
	// CreatingState 创建中
	CreatingState = "Creating"
	// StartingState 启动中
	StartingState = "Starting"
	// StoppingState 停止中
	StoppingState = "Stopping"
	// UpdatingState 更新中
	UpdatingState = "Updating"
	// ScalingState 扩缩容中
	ScalingState = "Scaling"
)

// 频率
const (
	_ = iota
	frequency1
	frequency2
	frequency3
	frequency4
	frequency5
)

// 轮询频率(Polling frequency)
var (
	// pollFrequency1 对Azure结果轮询频率为1秒一次
	pollFrequency1 = &runtime.PollUntilDoneOptions{Frequency: frequency1 * time.Second}
	// pollFrequency2 对Azure结果轮询频率为2秒一次
	pollFrequency2 = &runtime.PollUntilDoneOptions{Frequency: frequency2 * time.Second}
	// pollFrequency3 对Azure结果轮询频率为3秒一次
	pollFrequency3 = &runtime.PollUntilDoneOptions{Frequency: frequency3 * time.Second}
	// pollFrequency4 对Azure结果轮询频率为4秒一次
	pollFrequency4 = &runtime.PollUntilDoneOptions{Frequency: frequency4 * time.Second}
	// pollFrequency5 对Azure结果轮询频率为5秒一次
	pollFrequency5 = &runtime.PollUntilDoneOptions{Frequency: frequency5 * time.Second}
)

const (
	// NodeResourceGroup xxx
	NodeResourceGroup  = "nodeResourceGroup"
	aksManagedPoolName = "aks-managed-poolName"
)

// ImagePublishers for OS image publishers info
var ImagePublishers = map[string]string{
	"Canonical": "UbuntuServer",
	"OpenLogic": "CentOS",
	"RedHat":    "RHEL",
	"credativ":  "Debian",
	"SUSE":      "SLES",
}
