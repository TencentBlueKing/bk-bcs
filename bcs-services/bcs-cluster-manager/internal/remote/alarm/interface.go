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

package alarm

import (
	"errors"
	"time"
)

// AlarmInterface for alarm
type AlarmInterface interface {
	// ShieldHostAlarmConfig shield host alarm
	ShieldHostAlarmConfig(user string, config *ShieldHost) error
	// Name client name
	Name() string
}

var (
	// DefaultTimeOut default timeout
	DefaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// ShieldHost parameters
type ShieldHost struct {
	BizID    string
	HostList []HostInfo
}

// HostInfo xxx
type HostInfo struct {
	IP      string
	CloudID uint64
}
