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

package utils

import "strings"

// AlarmType xxx
type AlarmType int32

// AlarmTypeMap xxx
var AlarmTypeMap = map[AlarmType]string{
	SMS_ALARM:     "sms",
	RTX_ALALRM:    "rtx",
	WEIXIN_ALALRM: "weixin",
	MAIL_ALARM:    "email",
	VOICE_ALARM:   "voice",
}

const (
	// SMS_ALARM xxx
	SMS_ALARM AlarmType = 1 << 1
	// RTX_ALALRM xxx
	RTX_ALALRM AlarmType = 1 << 2
	// WEIXIN_ALALRM xxx
	WEIXIN_ALALRM AlarmType = 1 << 3
	// MAIL_ALARM xxx
	MAIL_ALARM AlarmType = 1 << 4
	// VOICE_ALARM xxx
	VOICE_ALARM AlarmType = 1 << 5

	// INFO_ALARM xxx
	INFO_ALARM AlarmType = RTX_ALALRM | WEIXIN_ALALRM
	// WARN_ALARM xxx
	WARN_ALARM AlarmType = RTX_ALALRM | WEIXIN_ALALRM | SMS_ALARM
	// ERROR_ALARM xxx
	ERROR_ALARM AlarmType = RTX_ALALRM | WEIXIN_ALALRM | SMS_ALARM | VOICE_ALARM
)

// IsValid xxx
func (a AlarmType) IsValid() bool {
	return int32(a) != 0
}

// IsSMS xxx
func (a AlarmType) IsSMS() bool {
	return (a & SMS_ALARM) == SMS_ALARM
}

// IsRtx xxx
func (a AlarmType) IsRtx() bool {
	return (a & RTX_ALALRM) == RTX_ALALRM
}

// IsWeiXin xxx
func (a AlarmType) IsWeiXin() bool {
	return (a & WEIXIN_ALALRM) == WEIXIN_ALALRM
}

// IsMail xxx
func (a AlarmType) IsMail() bool {
	return (a & MAIL_ALARM) == MAIL_ALARM
}

// IsVoice xxx
func (a AlarmType) IsVoice() bool {
	return (a & VOICE_ALARM) == VOICE_ALARM
}

// String 用于打印
func (a AlarmType) String() string {
	var str []string
	if a.IsSMS() {
		str = append(str, AlarmTypeMap[SMS_ALARM])
	}
	if a.IsRtx() {
		str = append(str, AlarmTypeMap[RTX_ALALRM])
	}

	if a.IsWeiXin() {
		str = append(str, AlarmTypeMap[WEIXIN_ALALRM])
	}

	if a.IsMail() {
		str = append(str, AlarmTypeMap[MAIL_ALARM])
	}

	if a.IsVoice() {
		str = append(str, AlarmTypeMap[VOICE_ALARM])
	}

	return strings.Join(str, "|")
}
