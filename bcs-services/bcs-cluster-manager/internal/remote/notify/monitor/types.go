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

// Package monitor xxx
package monitor

import (
	"time"
)

// MetricType 自定义指标类型
type MetricType string

var (
	// Event 自定义事件上报
	Event MetricType = "event"
	// Metrics 自定义指标上报
	Metrics MetricType = "metrics"
)

// ReportItems report metrics/event item
type ReportItems struct {
	DataId      int64   `json:"data_id"`
	AccessToken string  `json:"access_token"`
	Data        []*Item `json:"data"`
}

// Item metrics/event body
type Item struct {
	Target    string            `json:"target"`
	EventName string            `json:"event_name,omitempty"`
	Event     EventItem         `json:"event,omitempty"`
	Metrics   map[string]int    `json:"metrics,omitempty"`
	Dimension map[string]string `json:"dimension"`
	Timestamp int64             `json:"timestamp"`
}

// EventItem event body
type EventItem struct {
	Content string `json:"content"`
}

// ReportStatus report resp
type ReportStatus struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Result    string `json:"result"`
}

// basicConfig basic config
type basicConfig struct {
	// server data http server
	server string
	// dataId dataId channel
	dataId int64
	// accessToken token
	accessToken string
}

// ServerConfig server config
type serverConfig struct {
	basicConfig
	// timeout default 60s
	timeout time.Duration
	// customType (metrics / event)
	customType MetricType
	// debug http debug
	debug bool
}

// eventMetricsBody metrics/event body
type eventMetricsBody struct {
	eventName string
	eventBody string
	metrics   map[string]int
	dimension map[string]string
}
