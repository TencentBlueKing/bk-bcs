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

// TemplateSpaceCollect 定义了模板文件文件夹收藏的模型
type TemplateSpaceCollect struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	TemplateSpaceID primitive.ObjectID `json:"templateSpaceID" bson:"templateSpaceID"`
	ProjectCode     string             `json:"projectCode" bson:"projectCode"`
	Username        string             `json:"username" bson:"username"`
	CreateAt        int64              `json:"createAt" bson:"createAt"`
}

// TemplateSpaceAndCollect 定义了模板文件文件夹及收藏的模型
type TemplateSpaceAndCollect struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	TemplateSpaceID primitive.ObjectID `json:"templateSpaceID" bson:"templateSpaceID"`
	ProjectCode     string             `json:"projectCode" bson:"projectCode"`
	Username        string             `json:"username" bson:"username"`
	CreateAt        int64              `json:"createAt" bson:"createAt"`
	Name            string             `json:"name" bson:"name"` // 文件夹名称，从文件夹表获取
}

// ToMap trans template space collect to map
func (t *TemplateSpaceAndCollect) ToMap() map[string]interface{} {
	if t == nil {
		return nil
	}
	m := make(map[string]interface{}, 0)
	m["id"] = t.ID.Hex()
	m["templateSpaceID"] = t.TemplateSpaceID.Hex()
	m["projectCode"] = t.ProjectCode
	m["username"] = t.Username
	m["CreateAt"] = t.CreateAt
	m["name"] = t.Name
	return m
}
