/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package k8s xxx
package k8s

import (
	"fmt"
	"k8s.io/apiserver/pkg/apis/audit"
	"time"
)

// AuditLog k8s audit log
type AuditLog struct {
	Events map[string]AuditEvent
}

// AuditEvent single k8s audit log
type AuditEvent struct {
	audit.Event               // 嵌入 Kubernetes 的 AuditEvent 结构体
	Duration    time.Duration `json:"duration"` // 添加 Duration 属性
}

// RequestStats 结构体用于存储请求统计信息
type RequestStats struct {
	UserAgent     string
	Verb          string
	URI           string
	Count         int
	MaxAuditEvent *AuditEvent
	TotalTime     time.Duration
}

// CalculateDurations 计算每个请求的耗时并存储在结构体中
func (log *AuditLog) CalculateDurations() {
	for key, event := range log.Events {
		if event.Stage == audit.StageResponseComplete {
			event.Duration = event.StageTimestamp.Sub(event.RequestReceivedTimestamp.Time)
		}
		log.Events[key] = event
	}
}

// GetRequestStats 根据请求的 requestReceivedTimestamp, userAgent, verb, uri 统计请求数量和平均耗时
func (log *AuditLog) GetRequestStats() map[string]RequestStats {
	var stats = make(map[string]RequestStats)

	for _, event := range log.Events {
		key := fmt.Sprintf("%s-%s-%s", event.UserAgent, event.Verb, event.RequestURI)
		if addedStat, ok := stats[key]; !ok {
			stats[key] = RequestStats{
				UserAgent:     event.UserAgent,
				Verb:          event.Verb,
				URI:           event.RequestURI,
				Count:         1,
				MaxAuditEvent: &event,
				TotalTime:     event.Duration,
			}
		} else {
			addedStat.Count++
			addedStat.TotalTime += event.Duration
			if event.Duration > addedStat.MaxAuditEvent.Duration {
				addedStat.MaxAuditEvent = &event
			}
			stats[key] = addedStat
		}
	}

	return stats
}
