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

package action

import (
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	jsoniter "github.com/json-iterator/go" // nolint
)

// LogAction struct
type LogAction struct {
	Name      string
	ClusterID string
}

// Add add action
func (logAction *LogAction) Add(syncData *SyncData) {
	jsonString, err := jsoniter.Marshal(syncData.Data)
	if err != nil {
		glog.Errorf("Add json.Marshal fail: %v", syncData.Data)
	}
	glog.Infof("Add Cluster: %s - data: %s", logAction.ClusterID, string(jsonString))
}

// Delete delete action
func (logAction *LogAction) Delete(syncData *SyncData) {
	jsonString, err := jsoniter.Marshal(syncData.Data)
	if err != nil {
		glog.Errorf("Delete json.Marshal fail: %v", syncData.Data)
	}
	glog.Infof("Delete Cluster: %s - data: %s", logAction.ClusterID, string(jsonString))
}

// Update update action
func (logAction *LogAction) Update(syncData *SyncData) {
	jsonString, err := jsoniter.Marshal(syncData.Data)
	if err != nil {
		glog.Errorf("Update json.Marshal fail: %v", syncData.Data)
	}
	glog.Infof("Update Cluster: %s - data: %s", logAction.ClusterID, string(jsonString))
}
