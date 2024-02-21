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

// Package routematch is for route match operation
package routematch

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNoMatched no matched route pattern
var ErrNoMatched = errors.New("no matched route pattern")

// RouteMatcher is route matcher
type RouteMatcher struct {
	// routeMap is a map of method => [pattern]
	routeMap map[string][]string
	// patternMap is a map of patternRex => pattern
	patternMap map[string]string
}

// Route is route object
type Route struct {
	// Method is http method, eg: GET, POST, PUT, DELETE
	Method string
	// Pattern is path pattern, eg: /api/v1/config/biz/{biz_id}/apps/{app_id}
	Pattern string
}

// NewRouteMatcher 创建路由匹配器
func NewRouteMatcher(routes []Route) *RouteMatcher {
	routeMap := make(map[string][]string)
	patternMap := make(map[string]string)
	for _, r := range routes {
		// 将路由中的参数替换为正则表达式
		// eg: /api/v1/config/biz/{biz_id}/apps/{app_id} => /api/v1/config/biz/[^/]+/apps/[^/]+
		routeRegex := regexp.MustCompile("{[^/]+}")
		patternRegex := routeRegex.ReplaceAllString(r.Pattern, "[^/]+")
		patternMap[patternRegex] = r.Pattern

		method := strings.ToUpper(r.Method)
		if _, ok := routeMap[method]; !ok {
			routeMap[method] = make([]string, 0)
		}
		routeMap[method] = append(routeMap[method], patternRegex)
	}

	return &RouteMatcher{routeMap, patternMap}
}

// Match 获取最长匹配的路由pattern
// eg: method相同的情况下，/api/v1/config/biz/{biz_id}/apps/{app_id} 可匹配 /api/v1/config/biz/2/apps/1
func (m *RouteMatcher) Match(method, path string) (string, error) {
	var longestMatch string

	method = strings.ToUpper(method)
	patterns, ok := m.routeMap[method]
	if !ok {
		return "", ErrNoMatched
	}

	for _, pattern := range patterns {
		// 使用正则表达式查找匹配
		// NOCC:gas/error(忽略)
		match, _ := regexp.MatchString("^(/bscp)?"+pattern+"/?$", path)

		// 如果找到匹配并且比当前最长匹配更长，则更新最长匹配
		if match && len(pattern) > len(longestMatch) {
			longestMatch = pattern
		}
	}

	if len(longestMatch) == 0 {
		return "", ErrNoMatched
	}

	return m.patternMap[longestMatch], nil
}

// RouteMap return routemap of RouteMatcher
func (m *RouteMatcher) RouteMap() map[string][]string {
	return m.routeMap
}

// PatternMap return patternMap of RouteMatcher
func (m *RouteMatcher) PatternMap() map[string]string {
	return m.patternMap
}
