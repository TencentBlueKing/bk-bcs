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

// Reviewer Schema 检查器
type Reviewer struct {
	Rules []Rule
}

// Review 执行检查
func (r Reviewer) Review(schema *subSchema) Sugs {
	suggestions := Sugs{}
	for _, rule := range r.Rules {
		if sugs := rule.Validate(schema); sugs != nil && len(sugs) != 0 {
			suggestions = append(suggestions, sugs...)
		}
	}
	// 当前 schema 的 items 校验
	if schema.Items != nil {
		if itemsSugs := r.Review(schema.Items); len(itemsSugs) != 0 {
			suggestions = append(suggestions, itemsSugs...)
		}
	}
	// 当前 schema 的 properties 逐项校验
	if len(schema.Properties) != 0 {
		for _, v := range schema.Properties {
			if propSugs := r.Review(v); len(propSugs) != 0 {
				suggestions = append(suggestions, propSugs...)
			}
		}
	}
	return suggestions
}

// NativeReviewer json schema 原生规则检查
var NativeReviewer = Reviewer{
	Rules: []Rule{
		typeMustNotEmpty{},
		typeMustValid{},
		titleUnset{},
		descUnset{},
		defaultUnset{},
		minItemsUnsetOrZero{},
		maxItemsUnset{},
		uiPrefixKeyInProperties{},
	},
}

// UICompReviewer ui:component 字段检查
var UICompReviewer = Reviewer{
	Rules: []Rule{
		compNameUnset{},
		compNameMustValid{},
		clearableSetWithoutSelect{},
		searchableSetWithoutSelect{},
		multipleSetWithoutSelect{},
		dataSourceRequired{},
		dataSourceOrRemoteConfigRequired{},
		dataSourceAndRemoteConfigAllExists{},
		dataSourceSetWithoutSelectOrRadio{},
		remoteConfSetWithoutSelect{},
		dataSourceItemsMustValid{},
		remoteConfURLMustValid{},
	},
}

// UIReactionsReviewer ui:reactions 字段检查
var UIReactionsReviewer = Reviewer{
	Rules: []Rule{
		reactionSameSourceExists{},
	},
}

// UIRulesReviewer ui:rules 字段检查
var UIRulesReviewer = Reviewer{
	Rules: []Rule{
		ruleMustValid{},
	},
}

// UIGroupReviewer ui:group 字段检查
var UIGroupReviewer = Reviewer{
	Rules: []Rule{
		groupCompNameMustValid{},
		defaultActiveNameMustExists{},
	},
}

// UIOrderReviewer ui:order 字段检查
var UIOrderReviewer = Reviewer{
	Rules: []Rule{
		orderItemsMustExists{},
	},
}

// SchemaReviewers 受支持的检查器
var SchemaReviewers = []Reviewer{
	NativeReviewer,
	UICompReviewer,
	UIReactionsReviewer,
	UIRulesReviewer,
	UIGroupReviewer,
	UIOrderReviewer,
}
