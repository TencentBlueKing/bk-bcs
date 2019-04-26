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

package bsalarm

import (
	"bk-bcs/bcs-common/common/bcs-health/types"
)

const (
	defaultVersion = "v0.0.1"
	Delim          = '\n'
	DefaultSock    = "kafka.sock"
)

type Meta struct {
	APIVersion string `json:"api_version"`
}

type AlarmEvent struct {
	Meta `json:"meta"`
	// module specification
	Spec ModuleSpec `json:"spec"`
	// Event details
	Event Event `json:"event"`
	// extension informations about this event.
	Extensions EventExtension `json:"extensions"`
}

type ModuleSpec struct {
	// name of the module who send this event.
	ModuleName string `json:"module_name"`

	// cluster id that this module belongs to.
	ClusterID string `json:"cluster_id"`

	// if this is a container related service,
	// please set container's namespace.
	Namespace string `json:"namespace"`

	// ip address of this module.
	IP string `json:"ip"`

	// the version of this module.
	ModuleVersion string `json:"version"`
}

type Event struct {
	// which this event affiliation is, should be one of user, platform and both.
	// both means that this event shoud be cared by both user and platform.
	Affiliation types.AffiliationType `json:"affiliation"`

	// user defined application alarm level, which is value of
	// label with the key "io.tencent.bcs.monitor.level"
	AppAlarmLevel string `json:"app_alarm_level"`

	// the reason of this event occurred, which should
	// be a brief and short string like: "ContainerRestarted",
	// "ContainerHealthCheckFailed", "PullImageFailed", etc.
	Reason string `json:"reason"`

	// detailed messages about this event, which
	// will be send in the alarm.
	Messages string `json:"messages"`

	// event uuid, which is unique.
	UUID string `json:"uuid"`

	// when did this event occurred, which is
	// a unix time in seconds.
	AtTime int64 `json:"at_time"`
}

type EventExtension struct {
	// special infomations specified with this event,
	// which can be used in event filter.
	Labels map[string]string `json:"labels"`

	// user defined data, which can be the structed data.
	Context string `json:"context"`
}

type ExporterConfig struct {
	DataID string
}

type StdInput struct {
	DataID string      `json:"data_id"`
	Data   interface{} `json:"data"`
	UUID   string      `json:"uuid"`
}

type OutputType string

const (
	// which represent this is a log message.
	LogType OutputType = "log"
	// which represent this is a command execult result
	ResultType OutputType = "result"
)

type StdOutput struct {
	Success bool   `json:"success"`
	UUID    string `json:"uuid"`
	Message string `json:"message"`
}
