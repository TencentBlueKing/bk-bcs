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

package models

import "time"

type BCSClusterInfo struct {
	ID uint `gorm:"primary_key"`
	// ClusterId is a "one-to-one" field which connects this clsuterInfo to a bke cluster object
	ClusterId        string
	SourceProjectId  string `gorm:"size:32;unique_index:idx_source_cluster_project"`
	SourceClusterId  string `gorm:"size:100;unique_index:idx_source_cluster_project"`
	ClusterType      uint   `gorm:""`
	TkeClusterId     string
	TkeClusterRegion string
	CreatedAt        time.Time
}
