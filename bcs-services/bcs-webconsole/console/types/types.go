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

package types

import "time"

// WebSocketConfig is config
type WebSocketConfig struct {
	Height          int
	Width           int
	Privilege       bool
	Cmd             []string
	Tty             bool
	WebConsoleImage string
	Token           string
	Origin          string
	User            string
	PodName         string
}

type RespBase struct {
	Code      int         `json:"code"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
}

type Permissions struct {
	Test   bool `json:"test"`
	Prod   bool `json:"prod"`
	Create bool `json:"create"`
}

type APIResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// UserPodData 用户的pod数据
type UserPodData struct {
	UserName   string
	ProjectID  string
	ClustersID string
	PodName    string
	SessionID  string
	CrateTime  time.Time
}
