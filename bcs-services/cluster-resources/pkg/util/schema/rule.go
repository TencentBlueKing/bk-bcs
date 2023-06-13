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
	"fmt"
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

type (
	// Rule 检查规则
	Rule interface {
		Validate(schema *subSchema) Sugs
	}

	// jsonSchema 原生规则
	// 检查某些必须存在的 Key，以及部分值，由于 Go 是强类型语言，不做值类型检查

	// type 必须被设置
	typeMustNotEmpty struct{}

	// type 值必须为受支持的类型之一
	typeMustValid struct{}

	// title 字段推荐设置值
	titleUnset struct{}

	// description 字段没有设置值
	descUnset struct{}

	// default 字段没有设置值
	defaultUnset struct{}

	// minItems 没有设置值（当 items 不为 nil 时）
	minItemsUnsetOrZero struct{}

	// maxItems 没有设置值（当 items 不为 nil 时）
	maxItemsUnset struct{}

	// properties 中存在 ui: 前缀的 key，一般可能是层级错误
	uiPrefixKeyInProperties struct{}

	// ui:component 规则
	// name 建议设置值
	compNameUnset struct{}

	// name 值必须为受支持的组件名之一
	compNameMustValid struct{}

	// props.clearable 指定为 false，但组件不是 select
	clearableSetWithoutSelect struct{}

	// props.searchable 指定为 true，但组件不是 select
	searchableSetWithoutSelect struct{}

	// props.multiple 指定为 true，但组件不是 select
	multipleSetWithoutSelect struct{}

	// 当组件为 radio，则必须配置 dataSource
	dataSourceRequired struct{}

	// 当组件为 select，则 dataSource 或者 remoteConf 至少配置一项
	dataSourceOrRemoteConfigRequired struct{}

	// dataSource 与 remoteConf 同时存在
	dataSourceAndRemoteConfigAllExists struct{}

	// 设置 props.datasource 但组件不是 select 或者 radio
	dataSourceSetWithoutSelectOrRadio struct{}

	// datasource 中的每一项的 label & value 都不能为空
	dataSourceItemsMustValid struct{}

	// 设置 props.remoteConfig 但组件不是 select
	remoteConfSetWithoutSelect struct{}

	// remoteConfig.URL 不可为空，且为模板或者 http(s) url
	remoteConfURLMustValid struct{}

	// 为 bfInput 添加一些规则

	// ui:reactions 规则
	// 如果 source 都是同一个，则应该变成从 source 字段，指定 target 字段的 state
	reactionSameSourceExists struct{}

	// ui:rules 规则
	ruleMustValid struct{}

	// ui:group 规则
	// ui:group.name 必须为允许的组件名
	groupCompNameMustValid struct{}

	// 默认展开的字段，必须要存在
	defaultActiveNameMustExists struct{}

	// ui:order 规则
	// 排序字段必须存在
	orderItemsMustExists struct{}
)

func (r typeMustNotEmpty) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Major, "TypeMustNotEmpty", "type must not empty."
	if schema.Type != "" {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyType), detailTmpl}}
}

func (r typeMustValid) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Major, "TypeMustValid", "type must one of %v."
	if slice.StringInSlice(schema.Type, SchemaTypes) {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyType), fmt.Sprintf(detailTmpl, SchemaTypes)}}
}

func (r titleUnset) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := General, "TitleUnset", "title is unset, default value is property."
	if schema.Title != nil && *schema.Title != "" {
		return nil
	}
	// 如果是行数据，则降低建议等级为 Minor
	if schema.Source == SchemaSourceItems {
		level = Minor
	}
	return Sugs{{level, name, genNodePaths(schema, keyTitle), detailTmpl}}
}

func (r descUnset) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Minor, "DescUnset", "description is unset."
	if schema.Desc != nil && *schema.Desc != "" {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyDesc), detailTmpl}}
}

func (r defaultUnset) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Minor, "DefaultUnset", "default value is unset."
	if schema.Default != nil {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyDefault), detailTmpl}}
}

func (r minItemsUnsetOrZero) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Minor, "MinItemsUnsetOrZero", "minItems is unset or 0."
	if schema.Items == nil || schema.MinItems == nil || *schema.MinItems > 0 {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyMinItems), detailTmpl}}
}

func (r maxItemsUnset) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := General, "MaxItemsUnset", "maxItems is unset."
	if schema.Items == nil || schema.MaxItems == nil || *schema.MaxItems > 0 {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, keyMaxItems), detailTmpl}}
}

func (r uiPrefixKeyInProperties) Validate(schema *subSchema) (sugs Sugs) {
	level, name := Major, "UIPrefixKeyInProperties"
	detailTmpl := "key (%s) in properties can't start with `ui:`."

	for k := range schema.Properties {
		if strings.HasPrefix(k, "ui:") {
			sugs = append(
				sugs, &Suggestion{
					Level:    level,
					RuleName: name,
					NodePath: genNodePaths(schema, genSubPath(keyProperties, k)),
					Detail:   fmt.Sprintf(detailTmpl, k),
				},
			)
		}
	}
	return sugs
}

func (r compNameUnset) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := General, "UICompNameUnset", "ui:component.name unset, default is related to field type."
	if schema.UIComp == nil || schema.UIComp.Name == nil || *schema.UIComp.Name != "" {
		return nil
	}
	return Sugs{{level, name, genNodePaths(schema, genSubPath(keyUIComp, keyName)), detailTmpl}}
}

func (r compNameMustValid) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Major, "UICompNameMustValid", "ui:component.name must one of %v, current: %s."

	if schema.UIComp == nil || schema.UIComp.Name == nil || *schema.UIComp.Name == "" ||
		slice.StringInSlice(*schema.UIComp.Name, SchemaComps) {
		return nil
	}

	return Sugs{{
		level,
		name,
		genNodePaths(schema, genSubPath(keyUIComp, keyName)),
		fmt.Sprintf(detailTmpl, SchemaComps, *schema.UIComp.Name),
	}}
}

func (r clearableSetWithoutSelect) Validate(schema *subSchema) Sugs {
	level, name := Major, "ClearableWithoutSelect"
	detailTmpl := "clearable set as false when component name isn't select."

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name == compSelect {
		return nil
	}
	if schema.UIComp.Props.Clearable == nil || *schema.UIComp.Props.Clearable {
		return nil
	}

	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyClearable)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r searchableSetWithoutSelect) Validate(schema *subSchema) Sugs {
	level, name := Major, "SearchableWithoutSelect"
	detailTmpl := "searchable set as true when component name isn't select."

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name == compSelect {
		return nil
	}
	if schema.UIComp.Props.Searchable == nil || !*schema.UIComp.Props.Searchable {
		return nil
	}

	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keySearchable)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r multipleSetWithoutSelect) Validate(schema *subSchema) Sugs {
	level, name := Major, "MultipleSetWithoutSelect"
	detailTmpl := "multiple set as true when component name isn't select."

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name == compSelect {
		return nil
	}
	if schema.UIComp.Props.Multiple == nil || !*schema.UIComp.Props.Multiple {
		return nil
	}

	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyMultiple)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r dataSourceRequired) Validate(schema *subSchema) Sugs {
	level, name := Major, "DataSourceRequired"
	detailTmpl := "datasource is required when component name is radio."

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name != compRadio {
		return nil
	}
	if len(schema.UIComp.Props.DataSource) != 0 {
		return nil
	}

	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyDataSource)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r dataSourceOrRemoteConfigRequired) Validate(schema *subSchema) Sugs {
	level, name := Major, "DataSourceOrRemoteConfigRequired"
	detailTmpl := "datasource or remoteconfig is required when component name is select."

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name != compSelect {
		return nil
	}

	// 兼容多选 + enum 的情况
	if schema.Items != nil && len(schema.Items.Enum) != 0 {
		return nil
	}
	// DataSource 与 RemoteConfig 必须且只能设置一个
	if len(schema.UIComp.Props.DataSource) != 0 || schema.UIComp.Props.RemoteConf != nil {
		return nil
	}
	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyDataSource+"/"+keyRemoteConf)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r dataSourceAndRemoteConfigAllExists) Validate(schema *subSchema) Sugs {
	level, name := Major, "DataSourceOrRemoteConfigAllExists"
	detailTmpl := "datasource and remoteconfig can't all exists."

	if schema.UIComp == nil || schema.UIComp.Props == nil {
		return nil
	}

	// 任意一项不设置，都是可以的
	if len(schema.UIComp.Props.DataSource) == 0 || schema.UIComp.Props.RemoteConf == nil {
		return nil
	}
	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyDataSource+"/"+keyRemoteConf)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r dataSourceSetWithoutSelectOrRadio) Validate(schema *subSchema) Sugs {
	level, name := Major, "DataSourceSetWithoutSelect"
	detailTmpl := "datasource set when component name isn't select."

	if schema.UIComp == nil || schema.UIComp.Props == nil || schema.UIComp.Name == nil ||
		*schema.UIComp.Name == compSelect || *schema.UIComp.Name == compRadio {
		return nil
	}

	// 组件不是 select 时候，不可以设置 datasource
	if len(schema.UIComp.Props.DataSource) == 0 {
		return nil
	}
	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyDataSource)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r dataSourceItemsMustValid) Validate(schema *subSchema) (sugs Sugs) {
	level, name := Major, "DataSourceItemsMustValid"
	detailTmpl := "datasource's %s can't be empty."

	if schema.UIComp == nil || schema.UIComp.Props == nil || len(schema.UIComp.Props.DataSource) == 0 {
		return nil
	}

	propsSubPath := genSubPath(keyUIComp, keyProps)
	for idx, ds := range schema.UIComp.Props.DataSource {
		dsSubPath := genSubPathWithIdx(propsSubPath, keyDataSource, idx)
		if ds.Label == "" {
			sugs = append(
				sugs, &Suggestion{
					Level:    level,
					RuleName: name,
					NodePath: genNodePaths(schema, genSubPath(dsSubPath, keyLabel)),
					Detail:   fmt.Sprintf(detailTmpl, keyLabel),
				},
			)
		}
		if ds.Value == "" {
			sugs = append(
				sugs, &Suggestion{
					Level:    level,
					RuleName: name,
					NodePath: genNodePaths(schema, genSubPath(dsSubPath, keyValue)),
					Detail:   fmt.Sprintf(detailTmpl, keyValue),
				},
			)
		}
	}
	return sugs
}

func (r remoteConfSetWithoutSelect) Validate(schema *subSchema) Sugs {
	level, name := Major, "RemoteConfigSetWithoutSelect"
	detailTmpl := "remoteConfig set when component name isn't select"

	if schema.UIComp == nil || schema.UIComp.Props == nil ||
		schema.UIComp.Name == nil || *schema.UIComp.Name == compSelect {
		return nil
	}

	// 组件不是 select 时候，不可以设置 remoteConfig
	if schema.UIComp.Props.RemoteConf == nil {
		return nil
	}
	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyRemoteConf)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r remoteConfURLMustValid) Validate(schema *subSchema) Sugs {
	level, name := Major, "RemoteConfigURLMustValid"
	detailTmpl := "remote config url must be template or http(s) url."

	if schema.UIComp == nil || schema.UIComp.Props == nil || schema.UIComp.Props.RemoteConf == nil {
		return nil
	}

	url := schema.UIComp.Props.RemoteConf.URL
	if strings.HasPrefix(*url, "{{") && strings.HasSuffix(*url, "}}") {
		return nil
	}
	if strings.HasPrefix(*url, "http") {
		return nil
	}
	subPath := genSubPath(genSubPath(keyUIComp, keyProps), keyRemoteConf)
	return Sugs{{level, name, genNodePaths(schema, subPath), detailTmpl}}
}

func (r reactionSameSourceExists) Validate(schema *subSchema) (sugs Sugs) {
	level, name := Major, "ReactionSourceAllEqual"
	detailTmpl := "reaction source all equal, must use target from source field."

	if len(schema.UIReactions) == 0 {
		return nil
	}

	sources := set.NewStringSet()
	for idx, reaction := range schema.UIReactions {
		reactionSubPath := genSubPathWithIdx("", keyUIReactions, idx)
		if reaction.Source == nil {
			continue
		}
		if sources.Has(*reaction.Source) {
			sugs = append(
				sugs, &Suggestion{
					Level:    level,
					RuleName: name,
					NodePath: genNodePaths(schema, genSubPath(reactionSubPath, keySource)),
					Detail:   detailTmpl,
				},
			)
		}
		sources.Add(*reaction.Source)
	}
	return sugs
}

func (r ruleMustValid) Validate(schema *subSchema) (sugs Sugs) {
	level, name := Major, "RuleMustValid"
	detailTmpl := "rule must valid, ref or validator and message required."

	if len(schema.UIRules) == 0 {
		return nil
	}

	for idx, rule := range schema.UIRules {
		if rule.Ref == nil || *rule.Ref != "" {
			continue
		}
		if rule.Validator != nil && *rule.Validator != "" && rule.Message != nil && *rule.Message != "" {
			continue
		}
		sugs = append(
			sugs, &Suggestion{
				Level:    level,
				RuleName: name,
				NodePath: genNodePaths(schema, genSubPathWithIdx("", keyUIRules, idx)),
				Detail:   detailTmpl,
			},
		)
	}
	return sugs
}

func (r groupCompNameMustValid) Validate(schema *subSchema) Sugs {
	level, name, detailTmpl := Major, "groupCompNameMustValid", "ui:group.name must one of %v, current %s."

	if schema.UIGroup == nil || schema.UIGroup.Name == nil || *schema.UIGroup.Name == "" ||
		slice.StringInSlice(*schema.UIGroup.Name, SchemaComps) {
		return nil
	}

	return Sugs{{
		level,
		name,
		genNodePaths(schema, genSubPath(keyUIGroup, keyName)),
		fmt.Sprintf(detailTmpl, SchemaComps, *schema.UIGroup.Name),
	}}
}

func (r defaultActiveNameMustExists) Validate(schema *subSchema) (sugs Sugs) {
	level, name, detailTmpl := Major, "defaultActiveNameMustExists", "defaultActiveName %v must exists"

	if schema.UIGroup == nil || schema.UIGroup.Props == nil || len(schema.UIGroup.Props.DefaultActiveName) == 0 {
		return nil
	}

	subPathPrefix := genSubPath(keyUIGroup, keyProps)
	for idx, activeName := range schema.UIGroup.Props.DefaultActiveName {
		if _, ok := schema.Properties[activeName]; ok {
			continue
		}
		sugs = append(
			sugs, &Suggestion{
				Level:    level,
				RuleName: name,
				NodePath: genNodePaths(schema, genSubPathWithIdx(subPathPrefix, keyDefaultActiveName, idx)),
				Detail:   fmt.Sprintf(detailTmpl, activeName),
			},
		)
	}
	return sugs
}

func (r orderItemsMustExists) Validate(schema *subSchema) (sugs Sugs) {
	level, name, detailTmpl := Major, "orderItemsMustExists", "ui:order items %v must exists"

	if len(schema.UIOrder) == 0 {
		return nil
	}

	for idx, field := range schema.UIOrder {
		if _, ok := schema.Properties[field]; ok {
			continue
		}
		sugs = append(
			sugs, &Suggestion{
				Level:    level,
				RuleName: name,
				NodePath: genNodePaths(schema, genSubPathWithIdx("", keyUIOrder, idx)),
				Detail:   fmt.Sprintf(detailTmpl, field),
			},
		)
	}
	return sugs
}
