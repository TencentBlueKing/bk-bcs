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

package tmp

// ShieldHostAlarmRequest shield host alarm request
type ShieldHostAlarmRequest struct {
	AppID       string `json:"app_id"`
	IPList      string `json:"ip_list"`
	ShieldStart string `json:"shield_start"`
	ShieldEnd   string `json:"shield_end"`
	Operator    string `json:"operator"`
}

// ShieldHostAlarmResponse shield host alarm response
type ShieldHostAlarmResponse struct {
	Result  bool         `json:"result"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    ShieldIDList `json:"data"`
}

// ShieldIDList alarmID
type ShieldIDList struct {
	ShieldID []int `json:"shield_id"`
}
