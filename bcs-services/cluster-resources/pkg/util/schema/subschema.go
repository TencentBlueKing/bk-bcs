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

package schema

import (
	"github.com/fatih/structs"
	"github.com/spf13/cast"
)

// 定制化的 Schema 定义，没有支持全量的 structsSchema 字段，额外支持 bkui-form 扩展字段
type subSchema struct {
	Property string     `structs:"-"`
	Parent   *subSchema `structs:"-"`
	// 来源自（items/properties/root）
	Source string `structs:"-"`

	Type  string  `structs:"type"`
	Title *string `structs:"title,omitempty"`
	Desc  *string `structs:"description,omitempty"`

	Default  interface{} `structs:"default,omitempty"`
	Required []string    `structs:"required,omitempty"`
	Enum     []string    `structs:"enum,omitempty"`

	MinItems    *int  `structs:"minItems,omitempty"`
	MaxItems    *int  `structs:"maxItems,omitempty"`
	UniqueItems *bool `structs:"uniqueItems,omitempty"`

	Items      *subSchema            `structs:"items,omitempty"`
	Properties map[string]*subSchema `structs:"properties,omitempty"`

	UIComp      *uiComp      `structs:"ui:component,omitempty"`
	UIProps     *uiProps     `structs:"ui:props,omitempty"`
	UIReactions []uiReaction `structs:"ui:reactions,omitempty"`
	UIRules     []uiRule     `structs:"ui:rules,omitempty"`
	UIGroup     *uiGroup     `structs:"ui:group,omitempty"`
	UIOrder     []string     `structs:"ui:order,omitempty"`
}

// AsMap 将 schema 转换成 map[string]interface{}
func (ss *subSchema) AsMap() map[string]interface{} {
	ret, _ := NewGoLoader(structs.Map(ss)).Load()
	return cast.ToStringMap(ret)
}

type uiComp struct {
	Name  *string      `structs:"name"`
	Props *uiCompProps `structs:"props,omitempty"`
}

type uiCompProps struct {
	Clearable   *bool        `structs:"clearable,omitempty"`
	Searchable  *bool        `structs:"searchable,omitempty"`
	Multiple    *bool        `structs:"multiple,omitempty"`
	DataSource  []dataSource `structs:"datasource,omitempty"`
	Placeholder *string      `structs:"placeholder,omitempty"`
	RemoteConf  *remoteConf  `structs:"remoteConfig,omitempty"`
	Min         *int         `structs:"min,omitempty"`
	Max         *int         `structs:"max,omitempty"`
	Unit        *string      `structs:"unit,omitempty"`
	Visible     *bool        `structs:"visible,omitempty"`
	Disabled    *bool        `structs:"disabled,omitempty"`
	Type        *string      `structs:"type,omitempty"`
	Rows        *int         `structs:"rows,omitempty"`
	MaxRows     *int         `structs:"maxRows,omitempty"`
}

type dataSource struct {
	Label    string  `structs:"label"`
	Value    string  `structs:"value"`
	Disabled *bool   `structs:"disabled,omitempty"`
	Tips     *string `structs:"tips,omitempty"`
}

type remoteConf struct {
	URL    *string                `structs:"url"`
	Params map[string]interface{} `structs:"params"`
}

type uiProps struct {
	ShowTitle  *bool `structs:"showTitle,omitempty"`
	LabelWidth *int  `structs:"labelWidth,omitempty"`
}

type uiReaction struct {
	Lifetime *string   `structs:"lifetime,omitempty"`
	Source   *string   `structs:"source,omitempty"`
	Target   *string   `structs:"target"`
	If       *string   `structs:"if"`
	Then     *uiEffect `structs:"then"`
	Else     *uiEffect `structs:"else,omitempty"`
}

type uiEffect struct {
	Actions []string `structs:"actions,omitempty"`
	State   *uiState `structs:"state,omitempty"`
}

type uiState struct {
	Value    interface{} `structs:"value,omitempty"`
	Visible  *bool       `structs:"visible,omitempty"`
	Disabled *bool       `structs:"disabled,omitempty"`
}

type uiRule struct {
	Ref       *string `structs:"ref,omitempty"`
	Validator *string `structs:"validator,omitempty"`
	Message   *string `structs:"message,omitempty"`
}

type uiGroup struct {
	Name  *string       `structs:"name,omitempty"`
	Props *uiGroupProps `structs:"props"`
	Style *uiGroupStyle `structs:"style,omitempty"`
}

type uiGroupProps struct {
	Type              *string  `structs:"type,omitempty"`
	ShowTitle         *bool    `structs:"showTitle,omitempty"`
	Border            *bool    `structs:"border,omitempty"`
	HideEmptyRow      *bool    `structs:"hideEmptyRow,omitempty"`
	DefaultActiveName []string `structs:"defaultActiveName,omitempty"`
	Verifiable        *bool    `structs:"verifiable,omitempty"`
}

type uiGroupStyle struct {
	Background *string `structs:"background"`
}
