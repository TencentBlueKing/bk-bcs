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

package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// Audit for audit
type Audit struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id"`
	ProjectCode         string             `json:"project_code" bson:"projectCode"`
	ClusterID           string             `json:"cluster_id" bson:"clusterID"`
	CollectorConfigID   int                `json:"collector_config_id" bson:"collectorConfigID"`
	CollectorConfigName string             `json:"collector_config_name" bson:"collectorConfigName"`
	DataID              int                `json:"data_id" bson:"dataID"`
	BKLogConfigName     string             `json:"bk_log_config_name" bson:"bkLogConfigName"`
	CreatedAt           utils.JSONTime     `json:"created_at" bson:"createdAt"`
	UpdatedAt           utils.JSONTime     `json:"updated_at" bson:"updatedAt"`
}
