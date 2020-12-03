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

package watch

import (
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

// EventType event type
type EventType int

const (
	// EventMetricUpd metric update event
	EventMetricUpd EventType = iota
	// EventMetricDel metric delete event
	EventMetricDel
	// EventDynamicUpd dynamic update event
	EventDynamicUpd
)

func (et EventType) String() string {
	return eventTypeNames[et]
}

var (
	eventTypeNames = map[EventType]string{
		EventMetricUpd:  "MetricUpdate",
		EventMetricDel:  "MetricDelete",
		EventDynamicUpd: "DynamicUpdate",
	}
)

// MetricEvent metric event
type MetricEvent struct {
	ID     string
	Type   EventType
	Metric *types.Metric
	First  bool
	Last   bool
	Meta   map[string]btypes.ObjectMeta
}
