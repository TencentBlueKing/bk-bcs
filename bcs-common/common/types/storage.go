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

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
)

// BcsStorageDynamicIf define storage dynamic interface data interaction
type BcsStorageDynamicIf struct {
	Data interface{} `json:"data"`
}

// BcsStorageDynamicBatchDeleteIf define storage dynamic batch delete interface data interaction
type BcsStorageDynamicBatchDeleteIf struct {
	UpdateTimeBegin int64 `json:"updateTimeBegin"`
	UpdateTimeEnd   int64 `json:"updateTimeEnd"`
}

// BcsStorageWatchIf define storage watch interface data interaction
type BcsStorageWatchIf struct {
	Data interface{} `json:"data"`
}

// BcsStorageEventIf define storage event interface data interaction
type BcsStorageEventIf struct {
	ID string `json:"id"`
	// Env describes where is the event from, mesos or k8s
	Env EventEnv `json:"env"`
	// Kind describes which kind of resources this event belongs to, pod or rc or ..
	Kind EventKind `json:"kind"`
	// Level describes the level of this event, normal or warning or ..
	Level EventLevel `json:"level"`
	// Component describes which component of bcs create this event, scheduler or controller or ..
	Component EventComponent `json:"component"`
	// Type describes the type of this event, created or killing or ..
	Type string `json:"type"`
	// Describe is the detail of this event.
	Describe string `json:"describe"`
	// ClusterId describes which cluster is the event from.
	ClusterId string `json:"clusterId"`
	// EventTime is the time when the event occurs.
	EventTime int64 `json:"eventTime"`
	// ExtraInfo contains the specific info for each event.
	ExtraInfo EventExtraInfo `json:"extraInfo"`
	// Data is the raw data of event.
	Data interface{} `json:"data"`
}

type EventExtraInfo struct {
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Kind      ExtraKind `json:"kind"`
}

type ExtraKind string

const (
	ApplicationExtraKind ExtraKind = "application"
)

type EventEnv string

const (
	Event_Env_K8s   EventEnv = "k8s"
	Event_Env_Mesos EventEnv = "mesos"
)

type EventKind string

const (
	TaskEventKind EventKind = "task"
)

type EventLevel string

const (
	Event_Level_Warning EventLevel = "warning"
	Event_Level_Normal  EventLevel = "normal"
)

type EventComponent string

const (
	Event_Component_Scheduler  EventComponent = "scheduler"
	Event_Component_Controller EventComponent = "controller"
)

// BcsStorageClusterIf define storage config interface data interaction
// Send it to storage and get config data back.
type BcsStorageClusterIf struct {
	Service  string   `json:"service"`
	ZkIp     []string `json:"zkIp"`
	MasterIp []string `json:"masterIp"`
	DnsIp    []string `json:"dnsIp"`
	City     string   `json:"city"`
	JfrogUrl string   `json:"jfrogUrl"`
	NeedNat  bool     `json:"needNat"`
}

// BcsStorageRenderIf define storage render interface data interaction
// Call Gen() after initialization and send the base64 code to render bash.
type BcsStorageRenderIf struct {
	Version string `json:"version"`

	Data interface{} `json:"data"`
	// the username, it defines the folder where the package is.
	// Example: clusterkeeper, updater or other user.
	User string `json:"user"`
}

// Gen generate string
func (render *BcsStorageRenderIf) Gen() (r string, err error) {
	s, err := json.Marshal(render)
	if err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(s), nil
}

// GetData get data
func (render *BcsStorageRenderIf) GetData() (dc *DeployConfig, err error) {
	var tmp []byte
	if err = codec.EncJson(render.Data, &tmp); err != nil {
		return
	}
	dc = new(DeployConfig)
	if err = codec.DecJson(tmp, dc); err != nil {
		return
	}
	return
}

// BcsStorageAlarmIf define storage alarm interface data interaction
type BcsStorageAlarmIf struct {
	ClusterId    string      `json:"clusterId"`
	Namespace    string      `json:"namespace"`
	Message      string      `json:"message"`
	Source       string      `json:"source"`
	Module       string      `json:"module"`
	Type         string      `json:"type"`
	ReceivedTime int64       `json:"receivedTime"`
	Data         interface{} `json:"data"`
}

// BcsStorageMetricIf define storage metric interface data interaction
type BcsStorageMetricIf struct {
	Data interface{} `json:"data"`
}

// BcsStorageClusterRelationIf define storage cluster-ip relationship interface data interaction
type BcsStorageClusterRelationIf struct {
	Ips []string `json:"ips"`
}

// BcsStorageHostIf define storage set host config interface data interaction
type BcsStorageHostIf struct {
	Ip        string      `json:"ip"`
	ClusterId string      `json:"clusterId"`
	Data      interface{} `json:"data"`
}

// BcsStorageStableVersionIf define storage stableVersion interface data interaction
type BcsStorageStableVersionIf struct {
	Version string `json:"version"`
}

// WatchOptions watch options
type WatchOptions struct {
	// Only watch the node itself, including children added, children removed and node value change.
	// Will not receive existing children's event.
	SelfOnly bool `json:"selfOnly"`

	// Max time of events will received. Watch will be ended after the last event. 0 for infinity.
	MaxEvents uint `json:"maxEvents"`

	// The max waiting time of each event. Watch will be ended after timeout. 0 for no limit.
	Timeout time.Duration `json:"timeout"`

	// The value-change event will be checked if it's different from last status. If not then this event
	// will be ignored. And it will not trigger timeout reset.
	MustDiff string `json:"mustDiff"`
}
