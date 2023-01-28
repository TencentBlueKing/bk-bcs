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

const (
	// TypeArray 数组类型
	TypeArray = "array"
	// TypeBoolean 布尔类型
	TypeBoolean = "boolean"
	// TypeInteger 整数类型
	TypeInteger = "integer"
	// TypeNumber 数字类型
	TypeNumber = "number"
	// TypeNull 空类型
	TypeNull = "null"
	// TypeObject 对象类型
	TypeObject = "object"
	// TypeString 字符串类型
	TypeString = "string"
)

// SchemaTypes Schema 中允许的字段类型
var SchemaTypes = []string{
	TypeArray,
	TypeBoolean,
	TypeInteger,
	TypeNumber,
	TypeNull,
	TypeObject,
	TypeString,
}

const (
	// jsonSchema 标准字段（部分）
	keyTitle       = "title"
	keyDesc        = "description"
	keyType        = "type"
	keyItems       = "items"
	keyProperties  = "properties"
	keyRequired    = "required"
	keyDefault     = "default"
	keyMinItems    = "minItems"
	keyMaxItems    = "maxItems"
	keyUniqueItems = "uniqueItems"
	keyEnum        = "enum"

	// bkui-form 扩展字段（一级）
	keyUIComp      = "ui:component"
	keyUIProps     = "ui:props"
	keyUIReactions = "ui:reactions"
	keyUIRules     = "ui:rules"
	keyUIGroup     = "ui:group"
	keyUIOrder     = "ui:order"

	// bkui-form 扩展字段（二级）
	// ui:component
	keyName  = "name"
	keyProps = "props"

	// ui:props
	keyShowTitle  = "showTitle"
	keyLabelWidth = "labelWidth"

	// ui:reactions
	keyLifeTime = "lifetime"
	keySource   = "source"
	keyTarget   = "target"
	keyIf       = "if"
	keyThen     = "then"
	keyElse     = "else"

	// ui:rules
	keyValidator = "validator"
	keyMessage   = "message"

	// ui:group
	keyStyle = "style"

	// bkui-form 扩展字段（三级）
	// ui:component - props
	keyClearable   = "clearable"
	keySearchable  = "searchable"
	keyMultiple    = "multiple"
	keyDataSource  = "datasource"
	keyPlaceholder = "placeholder"
	keyRemoteConf  = "remoteConfig"
	keyMin         = "min"
	keyMax         = "max"
	keyUnit        = "unit"
	keyRows        = "rows"
	keyMaxRows     = "maxRows"

	// ui:reactions - then/else
	keyActions = "actions"
	keyState   = "state"

	// ui:group - style
	keyBackground = "background"

	// ui:group - props
	keyBorder            = "border"
	keyDefaultActiveName = "defaultActiveName"
	keyVerifiable        = "verifiable"
	keyHideEmptyRow      = "hideEmptyRow"

	// bkui-form 扩展字段（四级）
	// ui:component - props - datasource
	keyLabel    = "label"
	keyValue    = "value"
	keyDisabled = "disabled"
	keyTips     = "tips"

	// ui:component - props - remoteConfig
	keyParams = "params"
	keyURL    = "url"

	// ui:reactions - then/else - state
	keyVisible = "visible"
)

const (
	// 下拉框
	compSelect = "select"
	// 定制数组
	compBFArray = "bfArray"
	// 定制输入框（支持单位等）
	compBFInput = "bfInput"
	// 复选框
	compCheckBox = "checkbox"
	// 单选框
	compRadio = "radio"
	// 手风琴
	compCollapse = "collapse"
	// 卡片
	compCard = "card"
	// Tab
	compTab = "tab"
)

// SchemaComps 受支持的组件（可按需要增加）
var SchemaComps = []string{
	compSelect,
	compBFArray,
	compBFInput,
	compCheckBox,
	compRadio,
	compCollapse,
	compCard,
	compTab,
}

const (
	// SchemaPropertyRoot ...
	SchemaPropertyRoot = "root"

	// SchemaSourceRoot ...
	SchemaSourceRoot = "root"
	// SchemaSourceProperties ...
	SchemaSourceProperties = "properties"
	// SchemaSourceItems ...
	SchemaSourceItems = "items"
)
