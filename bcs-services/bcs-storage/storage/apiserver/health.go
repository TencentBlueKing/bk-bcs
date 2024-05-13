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

// Package apiserver xxx
package apiserver

import (
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
)

// BcsHealthIf interface
type BcsHealthIf struct {
	Code    int          `json:"code"`
	OK      bool         `json:"ok"`
	Data    healthStatus `json:"data"`
	Message string       `json:"message"`
}

// SubStatus sub status
type SubStatus struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type healthStatus map[string]*SubStatus

var (
	storageHealth = &BcsHealthIf{
		Code: 0,
		OK:   true,
		Data: healthStatus{
			mongodbConfigKey: {OK: true, Message: ""},
			zkConfigKey:      {OK: true, Message: ""},
			queueConfigKey:   {OK: true, Message: ""},
		},
		Message: "",
	}
	healthLock sync.Mutex
)

// SetUnhealthy set storage to unhealthy status
func SetUnhealthy(key string, message interface{}) {
	healthLock.Lock()
	defer healthLock.Unlock()
	storageHealth.OK = false
	storageHealth.Code = common.BcsErrStorageStatusNotReady
	storageHealth.Message = common.BcsErrStorageStatusNotReadyStr

	hs := storageHealth.Data[key]
	if hs == nil {
		hs = &SubStatus{OK: true}
		storageHealth.Data[key] = hs
	}
	if !hs.OK {
		return
	}
	hs.OK = false
	hs.Message = fmt.Sprintf("%v", message)
}

// GetHealth get health info
func GetHealth() metric.HealthMeta {
	healthLock.Lock()
	defer healthLock.Unlock()
	message := storageHealth.Message

	if !storageHealth.Data[mongodbConfigKey].OK {
		message += " | " + storageHealth.Data[mongodbConfigKey].Message // nolint goconst
	}
	if !storageHealth.Data[zkConfigKey].OK {
		message += " | " + storageHealth.Data[zkConfigKey].Message
	}
	if !storageHealth.Data[queueConfigKey].OK {
		message += " | " + storageHealth.Data[queueConfigKey].Message
	}

	return metric.HealthMeta{
		IsHealthy:   storageHealth.OK,
		Message:     message,
		CurrentRole: metric.SlaveRole,
	}
}
