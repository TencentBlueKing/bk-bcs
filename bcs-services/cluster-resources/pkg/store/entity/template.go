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

// Template 定义了模板文件元数据的模型
type Template struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	ProjectCode   string             `json:"projectCode" bson:"projectCode"`
	Description   string             `json:"description" bson:"description"`
	TemplateSpace string             `json:"templateSpace" bson:"templateSpace"`
	ResourceType  string             `json:"resourceType" bson:"resourceType"`
	Creator       string             `json:"creator" bson:"creator"`
	Updator       string             `json:"updator" bson:"updator"`
	CreateAt      int64              `json:"createAt" bson:"createAt"`
	UpdateAt      int64              `json:"updateAt" bson:"updateAt"`
	Tags          []string           `json:"tags" bson:"tags"`
	VersionMode   string             `json:"versionMode" bson:"versionMode"`
	Version       string             `json:"version" bson:"version"`
}

// ToMap trans Template to map
func (t *Template) ToMap() map[string]interface{} {
	if t == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = t.ID.Hex()
	m["name"] = t.Name
	m["projectCode"] = t.ProjectCode
	m["description"] = t.Description
	m["templateSpace"] = t.TemplateSpace
	m["resourceType"] = t.ResourceType
	m["creator"] = t.Creator
	m["updator"] = t.Updator
	m["createAt"] = t.CreateAt
	m["updateAt"] = t.UpdateAt
	m["tags"] = t.Tags
	m["versionMode"] = t.VersionMode
	m["version"] = t.Version
	return m
}
