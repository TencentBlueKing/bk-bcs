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

// Level 修改建议级别
type Level string

const (
	// Minor 无关紧要的建议，一般是优化项
	Minor Level = "minor"
	// General 建议修改，否则可能存在缺陷
	General Level = "general"
	// Major 必须修改，极有可能存在缺陷（一般在 NewSchema 时候已经触发）
	Major Level = "major"
)

// Suggestion schema 修改建议
type Suggestion struct {
	Level    Level
	RuleName string
	NodePath string
	Detail   string
}

// Sugs 建议的集合
type Sugs []*Suggestion

func (ss Sugs) filter(level Level) Sugs {
	newSugs := Sugs{}
	for _, s := range ss {
		if s.Level == level {
			newSugs = append(newSugs, s)
		}
	}
	return newSugs
}

// Minor 过滤出级别为 Minor 的建议
func (ss Sugs) Minor() Sugs {
	return ss.filter(Minor)
}

// General 过滤出级别为 General 的建议
func (ss Sugs) General() Sugs {
	return ss.filter(General)
}

// Major 过滤出级别为 Major 的建议
func (ss Sugs) Major() Sugs {
	return ss.filter(Major)
}
