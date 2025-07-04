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

// Package types defines the data structures for database models.
package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionPushEvent is the name of the push events collection in MongoDB.
const CollectionPushEvent = "push_events"

// PushEvent represents a push event record in the database.
type PushEvent struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty"`
	EventID             string              `bson:"event_id" json:"event_id"`
	RuleID              string              `bson:"rule_id" json:"rule_id"`
	Domain              string              `bson:"domain" json:"domain"`
	EventDetail         EventDetail         `bson:"event_detail" json:"event_detail"`
	PushLevel           string              `bson:"push_level" json:"push_level"`
	Status              int                 `bson:"status" json:"status"`
	NotificationResults NotificationResults `bson:"notification_results" json:"notification_results"`
	Dimension           Dimension           `bson:"dimension" json:"dimension"`
	BkBizName           string              `bson:"bk_biz_name" json:"bk_biz_name"`
	MetricData          MetricData          `bson:"metric_data" json:"metric_data"`
	CreatedAt           time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `bson:"updated_at" json:"updated_at"`
}

// EventDetail holds the detailed information of an event.
type EventDetail struct {
	Fields map[string]string `bson:"fields" json:"fields"`
}

// NotificationResults stores the results of notification channels.
type NotificationResults struct {
	Fields map[string]string `bson:"fields" json:"fields"`
}

// Dimension holds the dimensional information of an event.
type Dimension struct {
	Fields map[string]string `bson:"fields" json:"fields"`
}

// MetricData represents the metric data associated with an event.
type MetricData struct {
	MetricValue float64   `bson:"metric_value" json:"metric_value"`
	Timestamp   time.Time `bson:"timestamp" json:"timestamp"`
}
