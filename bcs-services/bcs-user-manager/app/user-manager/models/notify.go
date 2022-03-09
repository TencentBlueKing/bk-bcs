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

package models

import (
	"time"
)

const (
	NotifyByEmail NotifyType = iota
	NotifyByRtx
)

// NotifyType is the message type of the notify, 0 is email, 1 is rtx.
type NotifyType uint8

func (n NotifyType) String() string {
	switch n {
	case NotifyByEmail:
		return "email"
	case NotifyByRtx:
		return "rtx"
	default:
		return "none"
	}
}

// NotifyPhase is the phase of the notify
type NotifyPhase uint8

const (
	NonePhase NotifyPhase = iota
	OverduePhase
	DayPhase
	WeekPhase
	MonthPhase
)

type BcsTokenNotify struct {
	ID         uint        `json:"id" gorm:"primary_key"`
	Token      string      `json:"token" gorm:"size:64;index"`
	NotifyType NotifyType  `json:"notify_type" gorm:"type:tinyint(1);comment:'0:email,1:rtx'"`
	Phase      NotifyPhase `json:"phase" gorm:"type:tinyint(1);comment:'1:overdue,2:day,3:week,4:month'"`
	Result     bool        `json:"result"`
	Message    string      `json:"message" gorm:"size:255"`
	RequestID  string      `json:"request_id" gorm:"size:64"`
	CreatedAt  time.Time   `json:"created_at"`
	DeletedAt  *time.Time  `json:"deleted_at" gorm:"index"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
