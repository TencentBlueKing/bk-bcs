/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"github.com/fatih/structs"
	spb "google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
)

// NewAnnos ...
func NewAnnos(funcs ...AnnoFunc) Annotations {
	annos := Annotations{
		Perm: Perm{
			Page:  ObjPerm{},
			Items: map[ResUID]ObjPerm{},
		},
		FeatureFlag: map[FeatureFlagKey]bool{},
	}
	for _, f := range funcs {
		f(&annos)
	}
	return annos
}

// NewFeatureFlag 向注解中添加 FeatureFlag
func NewFeatureFlag(featureFlag FeatureFlagKey, enabled bool) AnnoFunc {
	return func(a *Annotations) {
		a.FeatureFlag[featureFlag] = enabled
	}
}

// NewPagePerm 向注解中添加 PagePerm
func NewPagePerm(objName ObjName, detail PermDetail) AnnoFunc {
	return func(a *Annotations) {
		a.Perm.Page[objName] = detail
	}
}

// NewItemPerm 向注解中添加 ItemPerm
func NewItemPerm(uid ResUID, objName ObjName, detail PermDetail) AnnoFunc {
	return func(a *Annotations) {
		if itemPerm, exists := a.Perm.Items[uid]; exists {
			itemPerm[objName] = detail
		} else {
			a.Perm.Items[uid] = map[ObjName]PermDetail{
				objName: detail,
			}
		}
	}
}

// AnnoFunc ...
type AnnoFunc func(*Annotations)

// Annotations Web 注解，参考结构：
/*
{
    "perm": {
        "page": {
            "createBtn": {
                "clickable": false,
                "tip": "没有权限"
            }
        },
        "items": {
            "{{ uid }}": {
                "editBtn": {
                    "clickable": false,
                    "tip": "系统命名空间不能编辑",
                    "applyURL": ""
                }
            }
        }
    },
    "featureFlag": {
		"serviceMonitor": false
	}
}
*/
type Annotations struct {
	Perm        Perm
	FeatureFlag map[FeatureFlagKey]bool
}

// ToPbStruct 将 Annotations 转换成 Struct 对象指针
// NOTE structs.Map 解析嵌套非原生类型，不会生成 map[string]interface{}，因此手动解析
func (a Annotations) ToPbStruct() (*spb.Struct, error) {
	// 解析 perm.page
	permPageMap := map[string]interface{}{}
	for objName, permDetail := range a.Perm.Page {
		permPageMap[string(objName)] = structs.Map(permDetail)
	}

	// 解析 perm.items
	permItemsMap := map[string]interface{}{}
	for uid, objPerm := range a.Perm.Items {
		resPerm := map[string]interface{}{}
		for objName, permDetail := range objPerm {
			resPerm[string(objName)] = structs.Map(permDetail)
		}
		permItemsMap[string(uid)] = resPerm
	}

	// 解析 featureFlag
	featureFlagMap := map[string]interface{}{}
	for ff, enabled := range a.FeatureFlag {
		featureFlagMap[string(ff)] = enabled
	}

	// 转换成 Struct 对象指针
	return pbstruct.Map2pbStruct(
		map[string]interface{}{
			"perm": map[string]interface{}{
				"page":  permPageMap,
				"items": permItemsMap,
			},
			"featureFlag": featureFlagMap,
		},
	)
}

// Perm 权限注解
type Perm struct {
	Page  ObjPerm
	Items map[ResUID]ObjPerm
}

// FeatureFlagKey FeatureFlag 键
type FeatureFlagKey string

// ObjName 前端页面对象名称（按钮等）
type ObjName string

// ObjPerm 前端对象权限信息集合
type ObjPerm map[ObjName]PermDetail

// ResUID 资源唯一 ID
type ResUID string

// PermDetail 权限信息
type PermDetail struct {
	Clickable bool   `structs:"clickable"`
	Tip       string `structs:"tip"`
	ApplyURL  string `structs:"applyURL"`
}
