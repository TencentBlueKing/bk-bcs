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

// Package models xxx
package models

import "time"

// ActivityStatus is the activity status
type ActivityStatus uint

// Validate validates the activity status
func (a ActivityStatus) Validate() error {
	if a > ActivityStatusPending {
		return nil
	}
	return nil
}

// String returns the string value of activity status
func (a ActivityStatus) String() string {
	switch a {
	case ActivityStatusSuccess:
		return "success"
	case ActivityStatusFailed:
		return "failed"
	case ActivityStatusPending:
		return "pending"
	default:
		return "unknow"
	}
}

// GetStatus returns the activity status
func GetStatus(s string) ActivityStatus {
	switch s {
	case "success":
		return ActivityStatusSuccess
	case "failed":
		return ActivityStatusFailed
	case "pending":
		return ActivityStatusPending
	default:
		return ActivityStatusUnknow
	}
}

const (
	// ActivityStatusUnknow means the activity status is unknow
	ActivityStatusUnknow ActivityStatus = iota
	// ActivityStatusSuccess means the activity is success
	ActivityStatusSuccess
	// ActivityStatusFailed means the activity is failed
	ActivityStatusFailed
	// ActivityStatusPending means the activity is pending
	ActivityStatusPending
)

// Activity is the activity model
type Activity struct {
	ID           uint           `json:"id" gorm:"primary_key"`
	ProjectCode  string         `json:"project_code" gorm:"not null;index:union_search;size:32"`
	ResourceType string         `json:"resource_type" gorm:"not null;index:union_search;size:16"`
	ResourceName string         `json:"resource_name"`
	ResourceID   string         `json:"resource_id"`
	ActivityType string         `json:"activity_type" gorm:"not null;index:union_search;size:16"`
	Status       ActivityStatus `json:"status"`
	Username     string         `json:"username" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at" gorm:"index:union_search;type:timestamp not null"`
	Description  string         `json:"description"`
	Extra        string         `json:"extra"`
}
