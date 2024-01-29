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
	ID                primitive.ObjectID  `json:"id" bson:"_id"`
	Name              string              `json:"name" bson:"name"`
	ProjectCode       string              `json:"projectCode" bson:"projectCode"`
	ClusterNamespaces []ClusterNamespaces `json:"clusterNamespaces" bson:"clusterNamespaces"`
	Filter            *ViewFilter         `json:"filter" bson:"filter"` // 筛选条件，如创建时间、label
	Scope             ViewScope           `json:"scope" bson:"scope"`
	CreateBy          string              `json:"createBy" bson:"createBy"`
	CreateAt          int64               `json:"createAt" bson:"createAt"`
	UpdateAt          int64               `json:"updateAt" bson:"updateAt"`
}

// ViewFilter 视图筛选条件
type ViewFilter struct {
	Name          string          `json:"name" bson:"name"`
	Creator       []string        `json:"creator" bson:"creator"`
	LabelSelector []LabelSelector `json:"labelSelector" bson:"labelSelector"`
}

// ViewScope 视图可见范围
type ViewScope int

const (
	// ViewScopePrivate 私有视图
	ViewScopePrivate ViewScope = 0
	// ViewScopePublic 公共视图
	ViewScopePublic ViewScope = 1
)

// ClusterNamespaces 集群命名空间
type ClusterNamespaces struct {
	ClusterID  string   `json:"clusterID" bson:"clusterID"`
	Namespaces []string `json:"namespaces" bson:"namespaces"`
}

// LabelSelector 视图筛选条件
type LabelSelector struct {
	Key    string   `json:"key" bson:"key"`
	Op     string   `json:"op" bson:"op"`
	Values []string `json:"values" bson:"values"`
}

// ToMap trans labelSelector to map
func (l *LabelSelector) ToMap() map[string]interface{} {
	if l == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["key"] = l.Key
	m["op"] = l.Op
	m["values"] = l.Values
	return m
}

// ToMap trans view to map
func (v *View) ToMap() map[string]interface{} {
	if v == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = v.ID.Hex()
	m["name"] = v.Name
	m["projectCode"] = v.ProjectCode
	m["filter"] = make(map[string]interface{}, 0)
	m["scope"] = int(v.Scope)

	cns := make([]map[string]interface{}, 0)
	for _, v := range v.ClusterNamespaces {
		cns = append(cns, map[string]interface{}{
			"clusterID":  v.ClusterID,
			"namespaces": v.Namespaces,
		})
	}

	m["clusterNamespaces"] = cns

	if v.Filter != nil {
		ls := make([]map[string]interface{}, 0)
		for _, v := range v.Filter.LabelSelector {
			ls = append(ls, v.ToMap())
		}
		m["filter"] = map[string]interface{}{
			"name":          v.Filter.Name,
			"creator":       v.Filter.Creator,
			"labelSelector": ls,
		}
	}
	return m
}
