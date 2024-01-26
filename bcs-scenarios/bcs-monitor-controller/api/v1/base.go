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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyncState is state for syncing process of polaris service
// Syncing: syncing is in process
// Completed: syncing process is successfully finished
// Failed: syncing process failed
type SyncState string

const (
	// SyncStateNeedReSync sync state NeedReSync
	SyncStateNeedReSync SyncState = "NeedReSync"
	// SyncStateCompleted sync state Completed
	SyncStateCompleted SyncState = "Completed"
	// SyncStateFailed sync state Failed
	SyncStateFailed SyncState = "Failed"

	// NoticeFatal notice level fatal
	NoticeFatal string = "fatal"
	// NoticeWarning notice level warning
	NoticeWarning string = "warning"
	// NoticeRemind  notice level remind
	NoticeRemind string = "remind"
)

const (
	// DefaultNamespace default namespace
	DefaultNamespace = "bcs-monitor"

	// LabelKeyForScenarioName label key for scenario name
	LabelKeyForScenarioName = "monitorcontroller.bkbcs.tencent.com/scenario"
	//// LabelKeyForScenarioRepo label key for scenario repo
	// LabelKeyForScenarioRepo = "monitorcontroller.bkbcs.tencent.com/scenario-repo"

	// LabelKeyForBizID label key for biz id
	LabelKeyForBizID = "monitorcontroller.bkbcs.tencent.com/biz_id"
	// LabelKeyForAppMonitorName label key for app monitor name
	LabelKeyForAppMonitorName = "monitorcontroller.bkbcs.tencent.com/appmonitor"
	// LabelKeyForResourceType label key for resource type
	LabelKeyForResourceType = "monitorcontroller.bkbcs.tencent.com/resource_type"
	// LabelKeyForMonitorApp label key for monitor app
	LabelKeyForMonitorApp = "monitorcontroller.bkbcs.tencent.com"

	// LabelValueResourceTypeMonitorRule label key for resource type monitor rule
	LabelValueResourceTypeMonitorRule = "monitor-rule"
	// LabelValueResourceTypeNoticeGroup label key for resource type monitor rule
	LabelValueResourceTypeNoticeGroup = "notice-group"
	// LabelValueResourceTypePanel label key for resource type panel
	LabelValueResourceTypePanel = "panel"
	// LabelValueResourceTypeConfigMap label key for resource type configmap
	LabelValueResourceTypeConfigMap = "configmap"

	// AnnotationScenarioUpdateTimestamp annotation key for scenario update timestamp
	AnnotationScenarioUpdateTimestamp = "monitorextension.bkbcs.tencent.com/scenario_update"
)

// SyncStatus defines status info of syncing process
type SyncStatus struct {
	State   SyncState `json:"state"`
	Message string    `json:"message,omitempty"`
	// +optional
	LastSyncTime metav1.Time `json:"lastSyncTime,omitempty"`
	App          string      `json:"app,omitempty"`
}
