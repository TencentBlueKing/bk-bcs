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

// TemplateSpace 定义了模板文件文件夹的模型
type TemplateSpace struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	ProjectCode string             `json:"projectCode" bson:"projectCode"`
	Description string             `json:"description" bson:"description"`
	Fav         bool               `json:"fav" bson:"-"` // 是否收藏
	Tags        []string           `json:"tags" bson:"tags"`
}

// ToMap trans templatespace to map
func (t *TemplateSpace) ToMap() map[string]interface{} {
	if t == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = t.ID.Hex()
	m["name"] = t.Name
	m["projectCode"] = t.ProjectCode
	m["description"] = t.Description
	m["fav"] = t.Fav
	m["tags"] = t.Tags
	return m
}
