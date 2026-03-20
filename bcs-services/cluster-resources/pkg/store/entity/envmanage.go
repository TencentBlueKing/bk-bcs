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
)

// EnvManage 定义了环境管理的模型
type EnvManage struct {
	ID                primitive.ObjectID  `json:"id" bson:"_id"`
	Env               string              `json:"env" bson:"env"`
	ProjectCode       string              `json:"projectCode" bson:"projectCode"`
	ClusterNamespaces []ClusterNamespaces `json:"clusterNamespaces" bson:"clusterNamespaces"`
}

// ToMap trans EnvManage to map
func (e *EnvManage) ToMap() map[string]interface{} {
	if e == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = e.ID.Hex()
	m["env"] = e.Env
	clusterNamespaces := make([]map[string]interface{}, 0)
	for _, v := range e.ClusterNamespaces {
		clusterNamespaces = append(clusterNamespaces, map[string]interface{}{
			"clusterID":  v.ClusterID,
			"nsgroup":    v.Nsgroup,
			"namespaces": v.Namespaces,
		})
	}
	m["clusterNamespaces"] = clusterNamespaces
	return m
}
