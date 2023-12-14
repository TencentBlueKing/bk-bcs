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

// Package search provides custom search functions.
package search

import (
	"fmt"
	"strings"

	"gorm.io/gen/field"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
)

// TableName is table name which support search.
type TableName string

const (
	// TemplateSpace is template space table
	TemplateSpace TableName = "template_spaces"
	// Template is template table
	Template TableName = "templates"
	// TemplateRevision is template revision table
	TemplateRevision TableName = "template_revisions"
	// TemplateSet is template set table
	TemplateSet TableName = "template_sets"
	// TemplateVariable is template space table
	TemplateVariable TableName = "template_variables"
	// ReleasedAppTemplate is released app template table
	ReleasedAppTemplate TableName = "released_app_templates"
	// ConfigItem is config item table
	ConfigItem TableName = "config_items"
	// ReleasedConfigItem is released config item table
	ReleasedConfigItem TableName = "released_config_items"

	/*
		Following tables are not real database tables which only used for search due to convenience and consistency
	*/

	// TmplBoundUnnamedApp is template bound unnamed app table
	TmplBoundUnnamedApp TableName = "template_bound_unnamed_apps"
	// TmplBoundNamedApp is template bound named app table
	TmplBoundNamedApp TableName = "template_bound_named_apps"
	// TmplRevisionBoundUnnamedApp is template revision bound unnamed app table
	TmplRevisionBoundUnnamedApp TableName = "template_revision_bound_unnamed_apps"
	// TmplRevisionBoundNamedApp is template revision bound unnamed app table
	TmplRevisionBoundNamedApp TableName = "template_revision_bound_named_apps"
)

// supportedFields is supported search fields of tables
var supportedFields = map[TableName][]string{
	// real tables
	TemplateSpace:       {"name", "memo", "creator", "reviser"},
	Template:            {"name", "path", "memo", "creator", "reviser"},
	TemplateRevision:    {"revision_name", "revision_memo", "name", "path", "creator"},
	TemplateSet:         {"name", "memo", "creator", "reviser"},
	TemplateVariable:    {"name", "memo", "creator", "reviser"},
	ReleasedAppTemplate: {"revision_name", "revision_memo", "name", "path", "creator"},
	ConfigItem:          {"name", "path", "memo", "creator", "reviser"},
	ReleasedConfigItem:  {"name", "path", "memo", "creator"},

	// unreal tables
	TmplBoundUnnamedApp:         {"app_name", "template_revision_name"},
	TmplBoundNamedApp:           {"app_name", "template_revision_name", "release_name"},
	TmplRevisionBoundUnnamedApp: {"app_name"},
	TmplRevisionBoundNamedApp:   {"app_name", "release_name"},
}

// supportedFieldsMap is supported search fields map of tables
var supportedFieldsMap = map[TableName]map[string]struct{}{
	// real tables
	TemplateSpace:       {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
	Template:            {"name": {}, "path": {}, "memo": {}, "creator": {}, "reviser": {}},
	TemplateRevision:    {"revision_name": {}, "revision_memo": {}, "name": {}, "path": {}, "creator": {}},
	TemplateSet:         {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
	TemplateVariable:    {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
	ReleasedAppTemplate: {"revision_name": {}, "revision_memo": {}, "name": {}, "path": {}, "creator": {}},
	ConfigItem:          {"name": {}, "path": {}, "memo": {}, "creator": {}, "reviser": {}},
	ReleasedConfigItem:  {"name": {}, "path": {}, "memo": {}, "creator": {}},

	// unreal tables
	TmplBoundUnnamedApp:         {"app_name": {}, "template_revision_name": {}},
	TmplBoundNamedApp:           {"app_name": {}, "template_revision_name": {}, "release_name": {}},
	TmplRevisionBoundUnnamedApp: {"app_name": {}},
	TmplRevisionBoundNamedApp:   {"app_name": {}, "release_name": {}},
}

// defaultFields is default search fields when field is not specified
var defaultFields = map[TableName][]string{
	// real tables
	TemplateSpace:       {"name"},
	Template:            {"name"},
	TemplateRevision:    {"revision_name"},
	TemplateSet:         {"name"},
	TemplateVariable:    {"name"},
	ReleasedAppTemplate: {"revision_name"},
	ConfigItem:          {"name"},
	ReleasedConfigItem:  {"name"},

	// unreal tables
	TmplBoundUnnamedApp:         {"app_name"},
	TmplBoundNamedApp:           {"app_name"},
	TmplRevisionBoundUnnamedApp: {"app_name"},
	TmplRevisionBoundNamedApp:   {"app_name"},
}

// getGenFieldsMap get the map for `table column name` => `gorm/gen field object`
func getGenFieldsMap(q *gen.Query) map[TableName]map[string]field.String {
	return map[TableName]map[string]field.String{
		TemplateSpace: {
			"name":    q.TemplateSpace.Name,
			"memo":    q.TemplateSpace.Memo,
			"creator": q.TemplateSpace.Creator,
			"reviser": q.TemplateSpace.Reviser,
		},
		Template: {
			"name":    q.Template.Name,
			"path":    q.Template.Path,
			"memo":    q.Template.Memo,
			"creator": q.Template.Creator,
			"reviser": q.Template.Reviser,
		},
		TemplateRevision: {
			"revision_name": q.TemplateRevision.RevisionName,
			"revision_memo": q.TemplateRevision.RevisionMemo,
			"name":          q.TemplateRevision.Name,
			"path":          q.TemplateRevision.Path,
			"creator":       q.TemplateRevision.Creator,
		},
		TemplateSet: {
			"name":    q.TemplateSet.Name,
			"memo":    q.TemplateSet.Memo,
			"creator": q.TemplateSet.Creator,
			"reviser": q.TemplateSet.Reviser,
		},
		TemplateVariable: {
			"name":    q.TemplateVariable.Name,
			"memo":    q.TemplateVariable.Memo,
			"creator": q.TemplateVariable.Creator,
			"reviser": q.TemplateVariable.Reviser,
		},
		ReleasedAppTemplate: {
			"revision_name": q.ReleasedAppTemplate.TemplateRevisionName,
			"revision_memo": q.ReleasedAppTemplate.TemplateRevisionMemo,
			"name":          q.ReleasedAppTemplate.Name,
			"path":          q.ReleasedAppTemplate.Path,
			"creator":       q.ReleasedAppTemplate.Creator,
		},
		ConfigItem: {
			"name":    q.ConfigItem.Name,
			"path":    q.ConfigItem.Path,
			"memo":    q.ConfigItem.Memo,
			"creator": q.ConfigItem.Creator,
			"reviser": q.ConfigItem.Reviser,
		},
		ReleasedConfigItem: {
			"name":    q.ReleasedConfigItem.Name,
			"path":    q.ReleasedConfigItem.Path,
			"memo":    q.ReleasedConfigItem.Memo,
			"creator": q.ReleasedConfigItem.Creator,
		},
	}
}

// Searcher is the interface for search
type Searcher interface {
	SearchExprs(q *gen.Query) []field.Expr
	SearchFields() []string
}

// searcher implements the Searcher interface
type searcher struct {
	fields          []string
	value           string
	tableName       TableName
	supportedFields []string //nolint:unused
}

// NewSearcher new a Searcher
func NewSearcher(fieldsStr string, value string, table TableName) (Searcher, error) {
	fields := make([]string, 0)
	if fieldsStr != "" {
		fields = strings.Split(fieldsStr, ",")
	}

	// validate the fields
	supported := supportedFieldsMap[table]
	var badFields []string
	for _, field := range fields {
		if _, ok := supported[field]; !ok {
			badFields = append(badFields, field)
		}
	}
	if len(badFields) > 0 {
		return nil, fmt.Errorf("not support field in %v, supported fields is %v", badFields, supportedFields[table])
	}

	return &searcher{
		fields:    fields,
		value:     value,
		tableName: table,
	}, nil

}

// SearchExprs implements the interface method
func (s *searcher) SearchExprs(q *gen.Query) []field.Expr {
	// if search value is not set, no need to search
	if s.value == "" {
		return []field.Expr{}
	}

	searchVal := "%" + s.value + "%"
	exprs := make([]field.Expr, 0)
	// if search fields is not specified, use the default search fields
	if len(s.fields) == 0 {
		for _, f := range defaultFields[s.tableName] {
			exprs = append(exprs, getGenFieldsMap(q)[s.tableName][f].Like(searchVal))
		}
		return exprs
	}

	for _, f := range s.fields {
		exprs = append(exprs, getGenFieldsMap(q)[s.tableName][f].Like(searchVal))
	}
	return exprs
}

// SearchFields implements the interface method
func (s *searcher) SearchFields() []string {
	// if search value is not set, no need to search
	if s.value == "" {
		return []string{}
	}

	// if search fields is not specified, use the default search fields
	if len(s.fields) == 0 {
		return defaultFields[s.tableName]
	}

	return s.fields
}
