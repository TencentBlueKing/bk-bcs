/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package search

import (
	"fmt"
	"strings"

	"gorm.io/gen/field"

	"bscp.io/pkg/dal/gen"
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
)

// supportedFields is supported search fields of tables
var supportedFields = map[TableName][]string{
	TemplateSpace:    {"name", "memo", "creator", "reviser"},
	Template:         {"name", "path", "memo", "creator", "reviser"},
	TemplateRevision: {"revision_name", "revision_memo", "creator"},
	TemplateSet:      {"name", "memo", "creator", "reviser"},
	TemplateVariable: {"name", "memo", "creator", "reviser"},
}

// supportedFieldsMap is supported search fields map of tables
var supportedFieldsMap = map[TableName]map[string]struct{}{
	TemplateSpace:    {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
	Template:         {"name": {}, "path": {}, "memo": {}, "creator": {}, "reviser": {}},
	TemplateRevision: {"revision_name": {}, "revision_memo": {}, "creator": {}},
	TemplateSet:      {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
	TemplateVariable: {"name": {}, "memo": {}, "creator": {}, "reviser": {}},
}

// defaultFields is default search fields when field is not specified
var defaultFields = map[TableName][]string{
	TemplateSpace:    {"name"},
	Template:         {"name"},
	TemplateRevision: {"revision_name"},
	TemplateSet:      {"name"},
	TemplateVariable: {"name"},
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
	supportedFields []string
}

// NewSearcher new a searcher
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

	exprs := make([]field.Expr, 0)
	// if search fields is not specified, use the default search fields
	if len(s.fields) == 0 {
		for _, f := range defaultFields[s.tableName] {
			exprs = append(exprs, getGenFieldsMap(q)[s.tableName][f].Regexp("(?i)"+s.value))
		}
		return exprs
	}

	for _, f := range s.fields {
		exprs = append(exprs, getGenFieldsMap(q)[s.tableName][f].Regexp("(?i)"+s.value))
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
