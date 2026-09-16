/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import (
	"errors"
	"strings"
)

func normalizePage(p Page) Page {
	if p.Page <= 0 {
		p.Page = defaultPage
	}
	if p.PageSize <= 0 {
		p.PageSize = defaultPageSize
	}
	return p
}

// Paginate 按 page（从 1 起）切列表
func Paginate(items []Instance, page, pageSize int) (int, []Instance) {
	total := len(items)
	if page <= 0 || pageSize <= 0 {
		return total, []Instance{}
	}
	start := (page - 1) * pageSize
	if start >= total {
		return total, []Instance{}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return total, items[start:end]
}

// MatchKeyword 对 id / display_name 做不区分大小写包含匹配
func MatchKeyword(id, displayName, keyword string) bool {
	if keyword == "" {
		return true
	}
	kw := strings.ToLower(keyword)
	return strings.Contains(strings.ToLower(id), kw) ||
		strings.Contains(strings.ToLower(displayName), kw)
}

// UniqueApprovers 去空去重审批人
func UniqueApprovers(vals ...string) []string {
	seen := make(map[string]struct{}, len(vals))
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// SplitManagers 拆分 managers 字符串
func SplitManagers(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ";", ",")
	s = strings.ReplaceAll(s, " ", ",")
	if s == "" {
		return nil
	}
	return UniqueApprovers(strings.Split(s, ",")...)
}

// ResolveParent 优先 filter.parent，否则取 ancestors 最后一级
func ResolveParent(filter Filter) (typ, id string) {
	if filter.Parent.ID != "" {
		return filter.Parent.Type, filter.Parent.ID
	}
	if n := len(filter.Ancestors); n > 0 {
		last := filter.Ancestors[n-1]
		return last.Type, last.ID
	}
	return "", ""
}

// AncestorID 按类型取 parent 或 ancestors 中的 ID
func AncestorID(filter Filter, typ string) string {
	if filter.Parent.Type == typ && filter.Parent.ID != "" {
		return filter.Parent.ID
	}
	for _, a := range filter.Ancestors {
		if a.Type == typ && a.ID != "" {
			return a.ID
		}
	}
	return ""
}

func applyRequires(ins Instance, requires []string) map[string]interface{} {
	out := map[string]interface{}{attrID: ins.ID}
	if len(requires) == 0 {
		if ins.DisplayName != "" {
			out[attrDisplayName] = ins.DisplayName
		}
		if ins.IAMPath != "" {
			out[attrIAMPath] = ins.IAMPath
		}
		if len(ins.Approvers) > 0 {
			out[attrApprovers] = ins.Approvers
		}
		return out
	}
	want := make(map[string]struct{}, len(requires))
	for _, r := range requires {
		want[r] = struct{}{}
	}
	if _, ok := want[attrDisplayName]; ok {
		out[attrDisplayName] = ins.DisplayName
	}
	if _, ok := want[attrIAMPath]; ok && ins.IAMPath != "" {
		out[attrIAMPath] = ins.IAMPath
	}
	if _, ok := want[attrApprovers]; ok && len(ins.Approvers) > 0 {
		out[attrApprovers] = ins.Approvers
	}
	return out
}

func applyRequiresList(items []Instance, requires []string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, ins := range items {
		out = append(out, applyRequires(ins, requires))
	}
	return out
}

// ParseNSID 从 IAM 命名空间 ID 解析集群 ID
func ParseNSID(nsID string) (string, error) {
	s := strings.Split(nsID, ":")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		return "", errors.New("invalid ns id")
	}
	return "BCS-K8S-" + s[0], nil
}
