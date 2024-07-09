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
	"github.com/coreos/go-semver/semver"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TemplateVersion 定义了模板文件版本的模型
type TemplateVersion struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	ProjectCode   string             `json:"projectCode" bson:"projectCode"`
	Description   string             `json:"description" bson:"description"`
	TemplateName  string             `json:"templateName" bson:"templateName"`
	TemplateSpace string             `json:"templateSpace" bson:"templateSpace"`
	Version       string             `json:"version" bson:"version"`
	EditFormat    string             `json:"editFormat" bson:"editFormat"`
	Content       string             `json:"content" bson:"content"`
	Creator       string             `json:"creator" bson:"creator"`
	CreateAt      int64              `json:"createAt" bson:"createAt"`
	Latest        bool               `json:"latest" bson:"-"` // 是否是最新版本，不存储在数据库
}

// ToMap trans TemplateVersion to map
func (t *TemplateVersion) ToMap() map[string]interface{} {
	if t == nil {
		return nil
	}
	editFormat := t.EditFormat
	if editFormat == "" {
		editFormat = "yaml"
	}
	m := make(map[string]interface{}, 0)
	m["id"] = t.ID.Hex()
	m["projectCode"] = t.ProjectCode
	m["description"] = t.Description
	m["templateName"] = t.TemplateName
	m["templateSpace"] = t.TemplateSpace
	m["version"] = t.Version
	m["editFormat"] = editFormat
	m["content"] = t.Content
	m["creator"] = t.Creator
	m["createAt"] = t.CreateAt
	m["latest"] = t.Latest
	return m
}

// TemplateID 定义了模板文件的唯一标识
type TemplateID struct {
	TemplateSpace   string `json:"templateSpace"`
	TemplateName    string `json:"templateName"`
	TemplateVersion string `json:"templateVersion"`
}

// VersionsSortByVersion sort template version by version
type VersionsSortByVersion []*TemplateVersion

// Len xxx
func (r VersionsSortByVersion) Len() int { return len(r) }

// Less xxx
func (r VersionsSortByVersion) Less(i, j int) bool {
	vA, err := semver.NewVersion(r[i].Version)
	if err != nil {
		return false
	}
	vB, err := semver.NewVersion(r[j].Version)
	if err != nil {
		return false
	}
	return vB.LessThan(*vA)
}

// Swap xxx
func (r VersionsSortByVersion) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// TemplateDeploy 定义了模板部署的一些标识
type TemplateDeploy struct {
	TemplateSpace   string `json:"templateSpace"`
	TemplateName    string `json:"templateName"`
	TemplateVersion string `json:"templateVersion"`
	Content         string `json:"content"`
}
