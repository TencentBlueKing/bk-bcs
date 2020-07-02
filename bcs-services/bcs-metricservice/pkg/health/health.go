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

package health

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"sync"
)

const (
	StorageKey = "storage"
	ApiKey     = "api"
)

// healthZ interface
type BcsHealthIf struct {
	Code        int             `json:"code"`
	OK          bool            `json:"ok"`
	Data        healthStatus    `json:"data"`
	Message     string          `json:"message"`
	CurrentRole metric.RoleType `json:"currentRole"`
}

type SubStatus struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type healthStatus map[string]*SubStatus

var (
	health = &BcsHealthIf{
		Code: 0,
		OK:   true,
		Data: healthStatus{
			StorageKey: {OK: true, Message: ""},
			ApiKey:     {OK: true, Message: ""},
		},
		Message: "",
	}
	healthLock sync.Mutex
)

func GetHealth() metric.HealthMeta {
	healthLock.Lock()
	defer healthLock.Unlock()
	message := health.Message

	health.OK = true
	if !health.Data[StorageKey].OK {
		message += " | " + health.Data[StorageKey].Message
		health.OK = false
	}
	if !health.Data[ApiKey].OK {
		message += " | " + health.Data[ApiKey].Message
		health.OK = false
	}
	if health.OK {
		message = ""
	}

	return metric.HealthMeta{
		IsHealthy:   health.OK,
		Message:     message,
		CurrentRole: health.CurrentRole,
	}
}

func SetUnhealthy(key string, message interface{}) {
	healthLock.Lock()
	defer healthLock.Unlock()
	health.Code = common.BcsErrMetricStatusNotReady
	health.Message = common.BcsErrMetricStatusNotReadyStr

	hs := health.Data[key]
	if hs == nil {
		hs = &SubStatus{OK: true}
		health.Data[key] = hs
	}
	if !hs.OK {
		return
	}
	hs.OK = false
	hs.Message = fmt.Sprintf("%v", message)
}

func SetHealth(key string) {
	healthLock.Lock()
	defer healthLock.Unlock()
	if status := health.Data[key]; status != nil {
		status.OK = true
		status.Message = ""
	}
}

func SetMaster() {
	healthLock.Lock()
	defer healthLock.Unlock()
	health.CurrentRole = metric.MasterRole
}
func SetSlave() {
	healthLock.Lock()
	defer healthLock.Unlock()
	health.CurrentRole = metric.SlaveRole
}
