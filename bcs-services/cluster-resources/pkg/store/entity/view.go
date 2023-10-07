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

// View 定义了视图配置的模型
type View struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	ProjectCode string             `json:"projectCode" bson:"projectCode"`
	ClusterID   string             `json:"clusterID" bson:"clusterID"`
	Namespace   string             `json:"namespace" bson:"namespace"`
	Filter      *ViewFilter        `json:"filter" bson:"filter"` // 筛选条件，如创建时间、label
	Scope       ViewScope          `json:"scope" bson:"scope"`
	CreateBy    string             `json:"createBy" bson:"createBy"`
	CreateAt    int64              `json:"createAt" bson:"createAt"`
	UpdateAt    int64              `json:"updateAt" bson:"updateAt"`
}

// ViewFilter 视图筛选条件
type ViewFilter struct {
	Name          string            `json:"name" bson:"name"`
	Creator       []string          `json:"creator" bson:"creator"`
	LabelSelector map[string]string `json:"labelSelector" bson:"labelSelector"`
}

// ViewScope 视图可见范围
type ViewScope int

const (
	// ViewScopePrivate 私有视图
	ViewScopePrivate ViewScope = 0
	// ViewScopePublic 公共视图
	ViewScopePublic ViewScope = 1
)

// ToMap trans view to map
func (v *View) ToMap() map[string]interface{} {
	if v == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = v.ID.Hex()
	m["name"] = v.Name
	m["projectCode"] = v.ProjectCode
	m["clusterID"] = v.ClusterID
	m["namespace"] = v.Namespace
	m["filter"] = make(map[string]interface{}, 0)
	m["scope"] = int(v.Scope)

	if v.Filter != nil {
		m["filter"] = map[string]interface{}{
			"name":          v.Filter.Name,
			"creator":       v.Filter.Creator,
			"labelSelector": v.Filter.LabelSelector,
		}
	}
	return m
}
