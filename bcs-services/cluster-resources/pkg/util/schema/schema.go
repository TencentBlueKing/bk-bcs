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
	"errors"
	"reflect"
	"strings"

	"github.com/spf13/cast"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// Schema ...
type Schema struct {
	raw        map[string]interface{}
	rootSchema *subSchema
}

// NewSchema instances a schema using the given JSONLoader
func NewSchema(l JSONLoader) (*Schema, error) {
	raw, err := l.Load()
	if err != nil {
		return nil, err
	}
	r, err := cast.ToStringMapE(raw)
	if err != nil {
		return nil, errors.New("schema document must be map[string]interface{}")
	}
	schema := &Schema{raw: r}
	if err = schema.parse(schema.raw); err != nil {
		return nil, err
	}
	return schema, nil
}

// Review 对 Schema 进行检查，返回修改意见
func (s *Schema) Review() (Sugs, error) {
	suggestions := Sugs{}
	for _, reviewer := range SchemaReviewers {
		if reSugs := reviewer.Review(s.rootSchema); len(reSugs) != 0 {
			suggestions = append(suggestions, reSugs...)
		}
	}
	return suggestions, nil
}

// Diff 将解析得到的 Schema 与原始数据做对比
func (s *Schema) Diff() []mapx.DiffRet {
	diffRets := []mapx.DiffRet{}
	for _, ret := range mapx.NewDiffer(s.raw, s.rootSchema.AsMap()).Do() {
		// 支持比较 ui:rules
		if strings.Contains(ret.Dotted, keyUIRules) {
			continue
		}
		// 支持忽略 ui:component, props, reactions 新增 nil
		if ret.Action == mapx.ActionAdd && strings.Contains(ret.Dotted, keyUIComp) && ret.NewVal == nil {
			continue
		}
		if ret.Action == mapx.ActionAdd && strings.Contains(ret.Dotted, keyProps) && ret.NewVal == nil {
			continue
		}
		if ret.Action == mapx.ActionAdd && strings.Contains(ret.Dotted, keyUIReactions) && ret.NewVal == nil {
			continue
		}
		diffRets = append(diffRets, ret)
	}
	return diffRets
}

// parse
func (s *Schema) parse(doc interface{}) error {
	s.rootSchema = &subSchema{Property: SchemaSourceRoot, Source: SchemaSourceRoot}
	return s.parseSchema(doc, s.rootSchema)
}

// parseSchema
func (s *Schema) parseSchema(docNode interface{}, curSchema *subSchema) error {
	// 节点数据不是 Map 类型，中止并抛出错误
	if !isKind(docNode, reflect.Map) {
		return NewSchemaInvalidErr()
	}

	m := cast.ToStringMap(docNode)

	if err := s.parseDesc(curSchema, m); err != nil {
		return err
	}

	if err := s.parseItems(curSchema, m); err != nil {
		return err
	}

	if err := s.parseProp(curSchema, m); err != nil {
		return err
	}

	return nil
}

// parseDesc
func (s *Schema) parseDesc(curSchema *subSchema, m map[string]interface{}) error {
	// type 只支持单类型，不支持复合类型，且必须存在
	if !mapx.ExistsKey(m, keyType) {
		return NewRequiredErr(curSchema, keyType)
	}
	if !isKind(m[keyType], reflect.String) {
		return NewInvalidTypeErr(curSchema, keyType, TypeString)
	}
	if v, ok := m[keyType].(string); ok {
		if !slice.StringInSlice(v, SchemaTypes) {
			return NewNotAValidTypeErr(curSchema, keyType, v)
		}
		curSchema.Type = v
	}

	// title
	if mapx.ExistsKey(m, keyTitle) && !isKind(m[keyTitle], reflect.String) {
		return NewInvalidTypeErr(curSchema, keyTitle, TypeString)
	}
	if v, ok := m[keyTitle].(string); ok {
		curSchema.Title = &v
	}

	// desc
	if mapx.ExistsKey(m, keyDesc) && !isKind(m[keyDesc], reflect.String) {
		return NewInvalidTypeErr(curSchema, keyDesc, TypeString)
	}
	if v, ok := m[keyDesc].(string); ok {
		curSchema.Desc = &v
	}

	// default
	if mapx.ExistsKey(m, keyDefault) {
		curSchema.Default = m[keyDefault]
	}

	// required 数组类型，元素为字符串类型
	if mapx.ExistsKey(m, keyRequired) {
		if !isKind(m[keyRequired], reflect.Slice) {
			return NewInvalidTypeErr(curSchema, keyRequired, TypeArray)
		}
		for idx, r := range cast.ToSlice(m[keyRequired]) {
			if !isKind(r, reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPathWithIdx("", keyRequired, idx), TypeString)
			}
			curSchema.Required = append(curSchema.Required, r.(string))
		}
	}
	return nil
}

// parseItems
func (s *Schema) parseItems(curSchema *subSchema, m map[string]interface{}) error {
	// minItems
	if mapx.ExistsKey(m, keyMinItems) {
		maxItemsIntValue := mustBeInteger(m[keyMinItems])
		if maxItemsIntValue == nil {
			return NewMustBeOfAnErr(curSchema, keyMinItems, TypeInteger)
		}
		if *maxItemsIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, keyMinItems)
		}
		curSchema.MinItems = maxItemsIntValue
	}

	// maxItems
	if mapx.ExistsKey(m, keyMaxItems) {
		maxItemsIntValue := mustBeInteger(m[keyMaxItems])
		if maxItemsIntValue == nil {
			return NewMustBeOfAnErr(curSchema, keyMaxItems, TypeInteger)
		}
		// maxItems 允许等于零太奇怪了，这里规则是最小为 1
		if *maxItemsIntValue < 1 {
			return NewMustBeGTEOneErr(curSchema, keyMaxItems)
		}
		curSchema.MaxItems = maxItemsIntValue
	}

	// uniqueItems
	if mapx.ExistsKey(m, keyUniqueItems) {
		if !isKind(m[keyUniqueItems], reflect.Bool) {
			return NewMustBeOfAErr(curSchema, keyUniqueItems, TypeBoolean)
		}
		uniqueItems := cast.ToBool(m[keyUniqueItems])
		curSchema.UniqueItems = &uniqueItems
	}

	// enum
	if mapx.ExistsKey(m, keyEnum) {
		if !isKind(m[keyEnum], reflect.Slice) {
			return NewInvalidTypeErr(curSchema, keyEnum, TypeArray)
		}
		for idx, e := range cast.ToSlice(m[keyEnum]) {
			if !isKind(e, reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPathWithIdx("", keyEnum, idx), TypeString)
			}
			curSchema.Enum = append(curSchema.Enum, e.(string))
		}
	}

	// items 是一个 schema
	if mapx.ExistsKey(m, keyItems) {
		if !isKind(m[keyItems], reflect.Map) {
			return NewInvalidTypeErr(curSchema, keyItems, TypeObject)
		}
		v := cast.ToStringMap(m[keyItems])
		newSchema := &subSchema{
			Property: curSchema.Property,
			Source:   SchemaSourceItems,
			Parent:   curSchema,
		}
		if err := s.parseSchema(v, newSchema); err != nil {
			return err
		}
		curSchema.Items = newSchema
	}
	return nil
}

// parseProp
func (s *Schema) parseProp(curSchema *subSchema, m map[string]interface{}) error {
	// properties
	if mapx.ExistsKey(m, keyProperties) {
		if err := s.parseProperties(m[keyProperties], curSchema); err != nil {
			return err
		}
	}

	// ui:component
	if mapx.ExistsKey(m, keyUIComp) {
		if err := parseUIComp(m[keyUIComp], curSchema); err != nil {
			return err
		}
	}

	// ui:props
	if mapx.ExistsKey(m, keyUIProps) {
		if err := parseUIProps(m[keyUIProps], curSchema); err != nil {
			return err
		}
	}

	// ui:reactions
	if mapx.ExistsKey(m, keyUIReactions) {
		if err := parseUIReactions(m[keyUIReactions], curSchema); err != nil {
			return err
		}
	}

	// ui:rules
	if mapx.ExistsKey(m, keyUIRules) {
		if err := parseUIRules(m[keyUIRules], curSchema); err != nil {
			return err
		}
	}

	// ui:group
	if mapx.ExistsKey(m, keyUIGroup) {
		if err := parseUIGroup(m[keyUIGroup], curSchema); err != nil {
			return err
		}
	}

	// ui:order
	if mapx.ExistsKey(m, keyUIOrder) {
		if err := parseUIOrder(m[keyUIOrder], curSchema); err != nil {
			return err
		}
	}
	return nil
}

// parseProperties
func (s *Schema) parseProperties(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Map) {
		return NewInvalidTypeErr(curSchema, keyProperties, TypeObject)
	}

	m := cast.ToStringMap(docNode)
	curSchema.Properties = map[string]*subSchema{}
	for k := range m {
		newSchema := &subSchema{
			Property: k,
			Source:   SchemaSourceProperties,
			Parent:   curSchema,
		}
		if err := s.parseSchema(m[k], newSchema); err != nil {
			return err
		}
		curSchema.Properties[k] = newSchema
	}
	return nil
}

// parseUIComp
func parseUIComp(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Map) {
		return NewInvalidTypeErr(curSchema, keyUIComp, TypeObject)
	}

	c := cast.ToStringMap(docNode)
	if len(c) == 0 {
		return NewEmptyMapErr(curSchema, keyUIComp)
	}
	curSchema.UIComp = &uiComp{}

	// ui:component.name
	if mapx.ExistsKey(c, keyName) {
		keyNameSubPath := genSubPath(keyUIComp, keyName)
		if !isKind(c[keyName], reflect.String) {
			return NewInvalidTypeErr(curSchema, keyNameSubPath, TypeString)
		}
		compName := cast.ToString(c[keyName])
		if !slice.StringInSlice(compName, SchemaComps) {
			return NewNotAValidCompErr(curSchema, keyNameSubPath, compName)
		}
		curSchema.UIComp.Name = &compName
	}

	// ui:component.props 若不存在则提前结束
	if !mapx.ExistsKey(c, keyProps) {
		return nil
	}
	if !isKind(c[keyProps], reflect.Map) {
		return NewInvalidTypeErr(curSchema, genSubPath(keyUIComp, keyProps), TypeObject)
	}
	curSchema.UIComp.Props = &uiCompProps{}

	p := cast.ToStringMap(c[keyProps])
	propsSubPath := genSubPath(keyUIComp, keyProps)

	if err := parseUICompClearable(curSchema, p, propsSubPath); err != nil {
		return err
	}

	if err := parseUICompDatasource(curSchema, p, propsSubPath); err != nil {
		return err
	}

	if err := parseUICompRemoteConf(curSchema, p, propsSubPath); err != nil {
		return err
	}

	if err := parseUICompDisabled(curSchema, p, propsSubPath); err != nil {
		return err
	}

	return nil
}

// parseUICompClearable
func parseUICompClearable(curSchema *subSchema, p map[string]interface{}, propsSubPath string) error {
	// ui:component.props.clearable
	if mapx.ExistsKey(p, keyClearable) {
		if !isKind(p[keyClearable], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyClearable), TypeBoolean)
		}
		clearable := cast.ToBool(p[keyClearable])
		curSchema.UIComp.Props.Clearable = &clearable
	}

	// ui:component.props.searchable
	if mapx.ExistsKey(p, keySearchable) {
		if !isKind(p[keySearchable], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keySearchable), TypeBoolean)
		}
		searchable := cast.ToBool(p[keySearchable])
		curSchema.UIComp.Props.Searchable = &searchable
	}

	// ui:component.props.multiple
	if mapx.ExistsKey(p, keyMultiple) {
		if !isKind(p[keyMultiple], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyMultiple), TypeBoolean)
		}
		multiple := cast.ToBool(p[keyMultiple])
		curSchema.UIComp.Props.Multiple = &multiple
	}
	return nil
}

// parseUICompDatasource
func parseUICompDatasource(curSchema *subSchema, p map[string]interface{}, propsSubPath string) error {
	// ui:component.props.datasource
	if mapx.ExistsKey(p, keyDataSource) {
		dsSubPath := genSubPath(propsSubPath, keyDataSource)
		if !isKind(p[keyDataSource], reflect.Slice) {
			return NewInvalidTypeErr(curSchema, dsSubPath, TypeArray)
		}

		for idx, _ds := range cast.ToSlice(p[keyDataSource]) {
			if !isKind(_ds, reflect.Map) {
				return NewInvalidTypeErr(curSchema, genSubPathWithIdx("", dsSubPath, idx), TypeObject)
			}
			ds := cast.ToStringMap(_ds)
			// 检查 label & value 键必须存在
			if !mapx.ExistsKey(ds, keyLabel) {
				return NewRequiredErr(curSchema, genSubPath(dsSubPath, keyLabel))
			}
			if !mapx.ExistsKey(ds, keyValue) {
				return NewRequiredErr(curSchema, genSubPath(dsSubPath, keyValue))
			}
			// 检查键值类型必须都是 string
			if !isKind(ds[keyLabel], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(dsSubPath, keyLabel), TypeString)
			}
			if !isKind(ds[keyValue], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(dsSubPath, keyValue), TypeString)
			}
			newDataSource := dataSource{
				Label: ds[keyLabel].(string), Value: ds[keyValue].(string),
			}
			if mapx.ExistsKey(ds, keyDisabled) {
				if !isKind(ds[keyDisabled], reflect.Bool) {
					return NewInvalidTypeErr(curSchema, genSubPath(dsSubPath, keyDisabled), TypeBoolean)
				}
				disabled := ds[keyDisabled].(bool)
				newDataSource.Disabled = &disabled
			}
			if mapx.ExistsKey(ds, keyTips) {
				if !isKind(ds[keyTips], reflect.String) {
					return NewInvalidTypeErr(curSchema, genSubPath(dsSubPath, keyTips), TypeString)
				}
				tips := ds[keyTips].(string)
				newDataSource.Tips = &tips
			}
			curSchema.UIComp.Props.DataSource = append(curSchema.UIComp.Props.DataSource, newDataSource)
		}
	}

	// ui:component.props.placeholder
	if mapx.ExistsKey(p, keyPlaceholder) {
		if !isKind(p[keyPlaceholder], reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyPlaceholder), TypeString)
		}
		placeholder := cast.ToString(p[keyPlaceholder])
		curSchema.UIComp.Props.Placeholder = &placeholder
	}
	return nil
}

// parseUICompRemoteConf
func parseUICompRemoteConf(curSchema *subSchema, p map[string]interface{}, propsSubPath string) error {
	// ui:component.props.remoteconfig
	if mapx.ExistsKey(p, keyRemoteConf) {
		rcSubPath := genSubPath(propsSubPath, keyRemoteConf)
		if !isKind(p[keyRemoteConf], reflect.Map) {
			return NewInvalidTypeErr(curSchema, rcSubPath, TypeObject)
		}

		rc := cast.ToStringMap(p[keyRemoteConf])
		// 若配置了 RemoteConfig，则 URL 必须是存在的
		if !mapx.ExistsKey(rc, keyURL) {
			return NewRequiredErr(curSchema, genSubPath(rcSubPath, keyURL))
		}
		url := cast.ToString(rc[keyURL])
		curSchema.UIComp.Props.RemoteConf = &remoteConf{URL: &url}

		if mapx.ExistsKey(rc, keyParams) {
			if !isKind(rc[keyParams], reflect.Map) {
				return NewInvalidTypeErr(curSchema, genSubPath(rcSubPath, keyParams), TypeObject)
			}
			curSchema.UIComp.Props.RemoteConf.Params = cast.ToStringMap(rc[keyParams])
		}
	}

	// ui:component.props.min
	if mapx.ExistsKey(p, keyMin) {
		minIntValue := mustBeInteger(p[keyMin])
		if minIntValue == nil {
			return NewMustBeOfAnErr(curSchema, genSubPath(propsSubPath, keyMin), TypeInteger)
		}
		if *minIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, genSubPath(propsSubPath, keyMin))
		}
		curSchema.UIComp.Props.Min = minIntValue
	}

	// ui:component.props.max
	if mapx.ExistsKey(p, keyMax) {
		maxIntValue := mustBeInteger(p[keyMax])
		if maxIntValue == nil {
			return NewMustBeOfAnErr(curSchema, genSubPath(propsSubPath, keyMax), TypeInteger)
		}
		if *maxIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, genSubPath(propsSubPath, keyMax))
		}
		curSchema.UIComp.Props.Max = maxIntValue
	}

	// ui:component.props.unit
	if mapx.ExistsKey(p, keyUnit) {
		if !isKind(p[keyUnit], reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyUnit), TypeString)
		}
		unit := cast.ToString(p[keyUnit])
		curSchema.UIComp.Props.Unit = &unit
	}
	return nil
}

// parseUICompDisabled
func parseUICompDisabled(curSchema *subSchema, p map[string]interface{}, propsSubPath string) error {
	// ui:component.props.disabled
	if mapx.ExistsKey(p, keyDisabled) {
		if !isKind(p[keyDisabled], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyDisabled), TypeBoolean)
		}
		disabled := cast.ToBool(p[keyDisabled])
		curSchema.UIComp.Props.Disabled = &disabled
	}

	// ui:component.props.visible
	if mapx.ExistsKey(p, keyVisible) {
		if !isKind(p[keyVisible], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyVisible), TypeBoolean)
		}
		visible := cast.ToBool(p[keyVisible])
		curSchema.UIComp.Props.Visible = &visible
	}

	// ui:component.props.type
	if mapx.ExistsKey(p, keyType) {
		if !isKind(p[keyType], reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyType), TypeString)
		}
		_type := cast.ToString(p[keyType])
		curSchema.UIComp.Props.Type = &_type
	}

	// ui:component.props.rows
	if mapx.ExistsKey(p, keyRows) {
		rowsIntValue := mustBeInteger(p[keyRows])
		if rowsIntValue == nil {
			return NewMustBeOfAnErr(curSchema, genSubPath(propsSubPath, keyRows), TypeInteger)
		}
		if *rowsIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, genSubPath(propsSubPath, keyRows))
		}
		curSchema.UIComp.Props.Rows = rowsIntValue
	}

	// ui:component.props.maxRows
	if mapx.ExistsKey(p, keyMaxRows) {
		maxRowsIntValue := mustBeInteger(p[keyMaxRows])
		if maxRowsIntValue == nil {
			return NewMustBeOfAnErr(curSchema, genSubPath(propsSubPath, keyMaxRows), TypeInteger)
		}
		if *maxRowsIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, genSubPath(propsSubPath, keyMaxRows))
		}
		curSchema.UIComp.Props.MaxRows = maxRowsIntValue
	}
	return nil
}

// parseUIProps
func parseUIProps(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Map) {
		return NewInvalidTypeErr(curSchema, keyUIProps, TypeObject)
	}

	p := cast.ToStringMap(docNode)
	if len(p) == 0 {
		return NewEmptyMapErr(curSchema, keyUIProps)
	}

	curSchema.UIProps = &uiProps{}
	// ui:props.showTitle
	if mapx.ExistsKey(p, keyShowTitle) {
		if !isKind(p[keyShowTitle], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(keyUIProps, keyShowTitle), TypeBoolean)
		}
		showTitle := cast.ToBool(p[keyShowTitle])
		curSchema.UIProps.ShowTitle = &showTitle
	}

	// ui:props.labelWidth
	if mapx.ExistsKey(p, keyLabelWidth) {
		labelWidthIntValue := mustBeInteger(p[keyLabelWidth])
		if labelWidthIntValue == nil {
			return NewMustBeOfAnErr(curSchema, genSubPath(keyUIProps, keyLabelWidth), TypeInteger)
		}
		if *labelWidthIntValue < 0 {
			return NewMustBeGTEZeroErr(curSchema, genSubPath(keyUIProps, keyLabelWidth))
		}
		curSchema.UIProps.LabelWidth = labelWidthIntValue
	}
	return nil
}

// parseUIReactions
func parseUIReactions(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Slice) {
		return NewInvalidTypeErr(curSchema, keyUIReactions, TypeArray)
	}

	var err error
	for idx, _r := range cast.ToSlice(docNode) {
		reaction := uiReaction{}
		reactionSubPath := genSubPathWithIdx("", keyUIReactions, idx)
		if !isKind(_r, reflect.Map) {
			return NewInvalidTypeErr(curSchema, reactionSubPath, TypeObject)
		}

		r := cast.ToStringMap(_r)

		// reaction.if
		if mapx.ExistsKey(r, keyIf) {
			_if := cast.ToString(r[keyIf])
			reaction.If = &_if
		}

		// reaction.target
		if mapx.ExistsKey(r, keyTarget) {
			target := cast.ToString(r[keyTarget])
			reaction.Target = &target
		}

		// reaction.then
		if !mapx.ExistsKey(r, keyThen) {
			return NewRequiredErr(curSchema, genSubPath(reactionSubPath, keyThen))
		}
		if reaction.Then, err = genUIEffect(r[keyThen], curSchema, reactionSubPath, keyThen); err != nil {
			return err
		}

		// reaction.lifetime
		if mapx.ExistsKey(r, keyLifeTime) {
			if !isKind(r[keyLifeTime], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(reactionSubPath, keyLifeTime), TypeString)
			}
			lifetime := cast.ToString(r[keyLifeTime])
			reaction.Lifetime = &lifetime
		}

		// reaction.source
		if mapx.ExistsKey(r, keySource) {
			if !isKind(r[keySource], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(reactionSubPath, keySource), TypeString)
			}
			source := cast.ToString(r[keySource])
			reaction.Source = &source
		}

		// reaction.else
		if mapx.ExistsKey(r, keyElse) {
			if reaction.Else, err = genUIEffect(r[keyElse], curSchema, reactionSubPath, keyElse); err != nil {
				return err
			}
		}
		curSchema.UIReactions = append(curSchema.UIReactions, reaction)
	}
	return nil
}

// genUIEffect
func genUIEffect(docNode interface{}, curSchema *subSchema, subPath, key string) (*uiEffect, error) {
	keySubPath := genSubPath(subPath, key)
	if !isKind(docNode, reflect.Map) {
		return nil, NewInvalidTypeErr(curSchema, keySubPath, TypeObject)
	}

	e := cast.ToStringMap(docNode)
	if len(e) == 0 {
		return nil, NewEmptyMapErr(curSchema, keySubPath)
	}

	ue := &uiEffect{}
	// then/else - actions
	if mapx.ExistsKey(e, keyActions) {
		if !isKind(e[keyActions], reflect.Slice) {
			return nil, NewInvalidTypeErr(curSchema, genSubPath(keySubPath, keyActions), TypeArray)
		}

		actions := cast.ToSlice(e[keyActions])
		for idx, a := range actions {
			if !isKind(a, reflect.String) {
				return nil, NewInvalidTypeErr(curSchema, genSubPathWithIdx(keySubPath, keyActions, idx), TypeString)
			}
			ue.Actions = append(ue.Actions, a.(string))
		}
	}

	// then/else - state
	if mapx.ExistsKey(e, keyState) {
		stateSubPath := genSubPath(keySubPath, keyState)
		if !isKind(e[keyState], reflect.Map) {
			return nil, NewInvalidTypeErr(curSchema, stateSubPath, TypeObject)
		}
		s := cast.ToStringMap(e[keyState])
		if len(s) == 0 {
			return ue, NewEmptyMapErr(curSchema, stateSubPath)
		}

		us := &uiState{}
		// then/else - state - value
		if mapx.ExistsKey(s, keyValue) {
			us.Value = s[keyValue]
		}

		// then/else - state - visible
		if mapx.ExistsKey(s, keyVisible) {
			if !isKind(s[keyVisible], reflect.Bool) {
				return nil, NewInvalidTypeErr(curSchema, genSubPath(stateSubPath, keyVisible), TypeBoolean)
			}
			visible := cast.ToBool(s[keyVisible])
			us.Visible = &visible
		}

		// then/else - state - disabled
		if mapx.ExistsKey(s, keyDisabled) {
			if !isKind(s[keyDisabled], reflect.Bool) {
				return nil, NewInvalidTypeErr(curSchema, genSubPath(stateSubPath, keyDisabled), TypeBoolean)
			}
			disabled := cast.ToBool(s[keyDisabled])
			us.Disabled = &disabled
		}
		ue.State = us
	}

	return ue, nil
}

// parseUIRules
func parseUIRules(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Slice) {
		return NewInvalidTypeErr(curSchema, keyUIRules, TypeArray)
	}

	rules := cast.ToSlice(docNode)
	for idx, _r := range rules {
		if isKind(_r, reflect.String) {
			// 引用全局规则
			ref := cast.ToString(_r)
			curSchema.UIRules = append(curSchema.UIRules, uiRule{Ref: &ref})
		} else if isKind(_r, reflect.Map) {
			// 组件自定义规则
			r := cast.ToStringMap(_r)
			ruleSubPath := genSubPathWithIdx("", keyUIRules, idx)
			// 检查 validator & message 键必须存在
			if !mapx.ExistsKey(r, keyValidator) {
				return NewRequiredErr(curSchema, genSubPath(ruleSubPath, keyValidator))
			}
			if !mapx.ExistsKey(r, keyMessage) {
				return NewRequiredErr(curSchema, genSubPath(ruleSubPath, keyMessage))
			}

			// 检查键值类型必须都是 string
			if !isKind(r[keyValidator], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(ruleSubPath, keyValidator), TypeString)
			}
			if !isKind(r[keyMessage], reflect.String) {
				return NewInvalidTypeErr(curSchema, genSubPath(ruleSubPath, keyMessage), TypeString)
			}
			validator, message := cast.ToString(r[keyValidator]), cast.ToString(r[keyMessage])
			curSchema.UIRules = append(curSchema.UIRules, uiRule{
				Validator: &validator, Message: &message,
			})
		}
	}
	return nil
}

// parseUIGroup
func parseUIGroup(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Map) {
		return NewInvalidTypeErr(curSchema, keyUIGroup, TypeObject)
	}

	g := cast.ToStringMap(docNode)
	if len(g) == 0 {
		return NewEmptyMapErr(curSchema, keyUIGroup)
	}
	curSchema.UIGroup = &uiGroup{}

	// ui:group.name
	if mapx.ExistsKey(g, keyName) {
		if !isKind(g[keyName], reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPath(keyUIGroup, keyName), TypeString)
		}
		compName := cast.ToString(g[keyName])
		if !slice.StringInSlice(compName, SchemaComps) {
			return NewNotAValidCompErr(curSchema, genSubPath(keyUIGroup, keyName), compName)
		}
		curSchema.UIGroup.Name = &compName
	}

	// ui:group.props
	if mapx.ExistsKey(g, keyProps) {
		if err := parseProps(g, curSchema); err != nil {
			return err
		}
	}

	// ui:group.style
	if mapx.ExistsKey(g, keyStyle) {
		if !isKind(g[keyStyle], reflect.Map) {
			return NewInvalidTypeErr(curSchema, genSubPath(keyUIGroup, keyStyle), TypeObject)
		}

		p := cast.ToStringMap(g[keyStyle])
		if mapx.ExistsKey(p, keyBackground) {
			if !isKind(p[keyBackground], reflect.String) {
				return NewInvalidTypeErr(curSchema, keyBackground, TypeString)
			}
			background := cast.ToString(p[keyBackground])
			curSchema.UIGroup.Style = &uiGroupStyle{Background: &background}
		}
	}
	return nil
}

func parseProps(g map[string]interface{}, curSchema *subSchema) error {
	propsSubPath := genSubPath(keyUIGroup, keyProps)
	if !isKind(g[keyProps], reflect.Map) {
		return NewInvalidTypeErr(curSchema, propsSubPath, TypeObject)
	}

	p := cast.ToStringMap(g[keyProps])
	if len(p) == 0 {
		return NewEmptyMapErr(curSchema, propsSubPath)
	}
	curSchema.UIGroup.Props = &uiGroupProps{}

	// ui:group.props.type
	if mapx.ExistsKey(p, keyType) {
		if !isKind(p[keyType], reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyType), TypeString)
		}
		_type := cast.ToString(p[keyType])
		curSchema.UIGroup.Props.Type = &_type
	}

	// ui:group.props.showTitle
	if mapx.ExistsKey(p, keyShowTitle) {
		if !isKind(p[keyShowTitle], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyShowTitle), TypeBoolean)
		}
		showTitle := cast.ToBool(p[keyShowTitle])
		curSchema.UIGroup.Props.ShowTitle = &showTitle
	}

	// ui:group.props.border
	if mapx.ExistsKey(p, keyBorder) {
		if !isKind(p[keyBorder], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyBorder), TypeBoolean)
		}
		border := cast.ToBool(p[keyBorder])
		curSchema.UIGroup.Props.Border = &border
	}

	// ui:group.props.defaultActiveName
	if mapx.ExistsKey(p, keyDefaultActiveName) {
		if !isKind(p[keyDefaultActiveName], reflect.Slice) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyDefaultActiveName), TypeArray)
		}
		for idx, n := range cast.ToSlice(p[keyDefaultActiveName]) {
			if !isKind(n, reflect.String) {
				return NewInvalidTypeErr(
					curSchema, genSubPathWithIdx(propsSubPath, keyDefaultActiveName, idx), TypeString,
				)
			}
			curSchema.UIGroup.Props.DefaultActiveName = append(
				curSchema.UIGroup.Props.DefaultActiveName, n.(string),
			)
		}
	}

	// ui:group.props.verifiable
	if mapx.ExistsKey(p, keyVerifiable) {
		if !isKind(p[keyVerifiable], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyVerifiable), TypeBoolean)
		}
		verifiable := cast.ToBool(p[keyVerifiable])
		curSchema.UIGroup.Props.Verifiable = &verifiable
	}

	// ui:group.props.hideEmptyRow
	if mapx.ExistsKey(p, keyHideEmptyRow) {
		if !isKind(p[keyHideEmptyRow], reflect.Bool) {
			return NewInvalidTypeErr(curSchema, genSubPath(propsSubPath, keyHideEmptyRow), TypeBoolean)
		}
		hideEmptyRow := cast.ToBool(p[keyHideEmptyRow])
		curSchema.UIGroup.Props.HideEmptyRow = &hideEmptyRow
	}
	return nil
}

// parseUIOrder
func parseUIOrder(docNode interface{}, curSchema *subSchema) error {
	if !isKind(docNode, reflect.Slice) {
		return NewInvalidTypeErr(curSchema, keyUIOrder, TypeArray)
	}

	orders := cast.ToSlice(docNode)
	for idx, o := range orders {
		if !isKind(o, reflect.String) {
			return NewInvalidTypeErr(curSchema, genSubPathWithIdx("", keyUIOrder, idx), TypeString)
		}
		curSchema.UIOrder = append(curSchema.UIOrder, o.(string))
	}
	return nil
}
