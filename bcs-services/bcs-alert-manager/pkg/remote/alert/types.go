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

package alert

import (
	"errors"
)

// AlarmAnnotationKey annotations
type AlarmAnnotationKey string

const (
	// AlarmAnnotationsUUID annotations uuid
	AlarmAnnotationsUUID AlarmAnnotationKey = "uuid"
	// AlarmAnnotationsHostIP annotations hostIP
	AlarmAnnotationsHostIP AlarmAnnotationKey = "host_ip"
	// AlarmAnnotationsBody annotations message
	AlarmAnnotationsBody AlarmAnnotationKey = "message"
	// AlarmAnnotationsComment annotations comment
	AlarmAnnotationsComment AlarmAnnotationKey = "comment"
)

// AlarmLevel type for alarm level
type AlarmLevel string

const (
	// ErrorKind message type for error
	ErrorKind AlarmLevel = "Error"
	// WarnKind message type for warn
	WarnKind AlarmLevel = "Warn"
	// InfoKind message type for info
	InfoKind AlarmLevel = "Info"
)

// AlarmLabelsKey for labels
type AlarmLabelsKey string

const (
	// AlarmLabelsModuleName labels module_name
	AlarmLabelsModuleName AlarmLabelsKey = "module_name"
	// AlarmLabelsModuleIP labels module_ip
	AlarmLabelsModuleIP AlarmLabelsKey = "ip"
	// AlarmLabelsAlarmName labels alarm_name
	AlarmLabelsAlarmName AlarmLabelsKey = "alertname"
	// AlarmLabelsAlertType labels alarm_type
	AlarmLabelsAlertType AlarmLabelsKey = "alert_type"

	// AlarmLabelsClusterID labels clusterID
	AlarmLabelsClusterID AlarmLabelsKey = "cluster_id"
	// AlarmLabelsClusterNameSpace labels namespace
	AlarmLabelsClusterNameSpace AlarmLabelsKey = "namespace"
	// AlarmLabelsAlarmResourceType labels resource_type
	AlarmLabelsAlarmResourceType AlarmLabelsKey = "resource_type"
	// AlarmLabelsAlarmResourceName labels resource_name
	AlarmLabelsAlarmResourceName AlarmLabelsKey = "resource_name"

	// AlarmLabelsAlarmLevel labels level
	AlarmLabelsAlarmLevel AlarmLabelsKey = "alarm_level"

	// AlarmLabelsAlarmProjectID
	AlarmLabelsAlarmProjectID AlarmLabelsKey = "project_id"
)

const (
	// Resource show resource kind
	Resource = "resource"
	// Module show module kind
	Module = "module"
)

// AlertServer ErrInfo
var (
	// ErrInitServerFail server error
	ErrInitServerFail = errors.New("init server failed")
	// ErrInvalidateOptions options error
	ErrInvalidateOptions = errors.New("options is nil")
	// ErrInvalidateServer server address error
	ErrInvalidateServer = errors.New("invalidate server address")
	// ErrInvalidateAuth appCode|appSecret error
	ErrInvalidateAuth = errors.New("invalidate appCode | appSecret")
	// ErrBadServer serverDown error
	ErrBadServer = errors.New("alert server down")
)
