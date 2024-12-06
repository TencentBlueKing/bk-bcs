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

package multicluster

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

const (
	// EmptyCreator 无创建者
	EmptyCreator = "--"

	// LabelCreator 创建者标签
	LabelCreator = "io.tencent.paas.creator"

	// OpEQ 等于
	OpEQ = "="
	// OpNotEQ 不等于
	OpNotEQ = "!="
	// OpIn 包含
	OpIn = "In"
	// OpNotIn 不包含
	OpNotIn = "NotIn"
	// OpExists 存在
	OpExists = "Exists"
	// OpDoesNotExist 不存在
	OpDoesNotExist = "DoesNotExist"

	// SortByName 按照名称排序
	SortByName SortBy = "name"
	// SortByNamespace 按照命名空间排序
	SortByNamespace SortBy = "namespace"
	// SortByAge 按照创建时间排序
	SortByAge SortBy = "age"

	// OrderDesc 降序
	OrderDesc Order = "desc"
	// OrderAsc 升序
	OrderAsc Order = "asc"
)

// SortBy 排序字段
type SortBy string

// Order 排序方式
type Order string

// QueryFilter 查询条件
type QueryFilter struct {
	Creator       []string // -- 代表无创建者
	Name          string
	CreateSource  *clusterRes.CreateSource
	LabelSelector []*clusterRes.LabelSelector
	IP            string   // IP 过滤条件，包括IPV4、IPV6、HostIP，目前仅 Pod 支持
	Status        []string // 状态过滤条件，目前仅 Deployment 支持
	SortBy        SortBy
	Order         Order
	Limit         int
	Offset        int
}

// Filter 过滤器
type Filter func(resources []*storage.Resource) []*storage.Resource

// ApplyFilter 应用过滤器
func ApplyFilter(resources []*storage.Resource, filters ...Filter) []*storage.Resource {
	for _, f := range filters {
		resources = f(resources)
	}
	return resources
}

// LabelSelectorString 转换为标签选择器字符串
// 操作符，=, In, NotIn, Exists, DoesNotExist，如果是 Exists/DoesNotExist，如果是, values为空，如果是=，values只有一个值，
// 如果是in/notin，values有多个值
func (f *QueryFilter) LabelSelectorString() string {
	var ls []string
	for _, v := range f.LabelSelector {
		if len(v.Values) == 0 && v.Op != OpExists && v.Op != OpDoesNotExist {
			continue
		}
		switch v.Op {
		case OpEQ:
			ls = append(ls, fmt.Sprintf("%s=%s", v.Key, v.Values[0]))
		case OpNotEQ:
			ls = append(ls, fmt.Sprintf("%s!=%s", v.Key, v.Values[0]))
		case OpIn:
			values := strings.Join(v.Values, ",")
			ls = append(ls, fmt.Sprintf("%s in (%s)", v.Key, values))
		case OpNotIn:
			values := strings.Join(v.Values, ",")
			ls = append(ls, fmt.Sprintf("%s notin (%s)", v.Key, values))
		case OpExists:
			ls = append(ls, v.Key)
		case OpDoesNotExist:
			ls = append(ls, fmt.Sprintf("!%s", v.Key))
		}
	}
	return strings.Join(ls, ",")
}

// ToConditions 转换为查询条件
func (f *QueryFilter) ToConditions() []*operator.Condition {
	conditions := []*operator.Condition{}

	// creator 过滤条件
	var emptyCreator bool
	for _, v := range f.Creator {
		if v == EmptyCreator {
			emptyCreator = true
		}
	}
	if len(f.Creator) > 0 {
		if emptyCreator {
			// 使用全角符号代替 '.',区分字段分隔，无创建者的资源，creator字段为 null
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					"data.metadata.annotations." + mapx.ConvertPath(LabelCreator): nil}))
		} else {
			conditions = append(conditions, operator.NewLeafCondition(
				operator.In, map[string]interface{}{
					"data.metadata.annotations." + mapx.ConvertPath(LabelCreator): f.Creator}))
		}
	}

	// name 过滤条件
	if f.Name != "" {
		conditions = append(conditions, operator.NewLeafCondition(
			operator.Con, map[string]string{"data.metadata.name": f.Name}))
	}

	// labelSelector 过滤条件
	for _, v := range f.LabelSelector {
		if len(v.Values) == 0 && v.Op != OpExists && v.Op != OpDoesNotExist {
			continue
		}
		switch v.Op {
		case OpEQ:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values[0]}))
		case OpNotEQ:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Not, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values[0]}))
		case OpIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.In, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpNotIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Nin, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpExists:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Ext, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): ""}))
		case OpDoesNotExist:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf("data.metadata.labels.%s", mapx.ConvertPath(v.Key)): nil}))
		}
	}
	return conditions
}

// CreatorFilter 创建者过滤器
func (f *QueryFilter) CreatorFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if len(f.Creator) == 0 {
		return resources
	}

	var emptyCreator bool
	for _, v := range f.Creator {
		if v == EmptyCreator {
			emptyCreator = true
		}
	}

	if emptyCreator {
		for _, v := range resources {
			if mapx.GetStr(v.Data, []string{"metadata", "annotations", mapx.ConvertPath(LabelCreator)}) == "" {
				result = append(result, v)
			}
		}
	} else {
		for _, v := range resources {
			if slice.StringInSlice(
				mapx.GetStr(v.Data, []string{"metadata", "annotations", mapx.ConvertPath(LabelCreator)}), f.Creator) {
				result = append(result, v)
			}
		}
	}

	return result
}

// NameFilter 名称过滤器
func (f *QueryFilter) NameFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.Name == "" {
		return resources
	}
	for _, v := range resources {
		if strings.Contains(mapx.GetStr(v.Data, "metadata.name"), f.Name) {
			result = append(result, v)
		}
	}
	return result
}

// CreateSourceFilter 创建来源过滤器
func (f *QueryFilter) CreateSourceFilter(resources []*storage.Resource) []*storage.Resource {
	if f.CreateSource == nil {
		return resources
	}

	resources = f.createSourceSourceFilter(resources)
	resources = f.createSourceTemplateNameFilter(resources)
	resources = f.createSourceTemplateVersionFilter(resources)
	resources = f.createSourceChartNameFilter(resources)

	return resources
}

// createSourceSourceFilter 来源过滤器
func (f *QueryFilter) createSourceSourceFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.CreateSource.Source == "" {
		return resources
	}

	for _, v := range resources {
		createSource, _ := formatter.ParseCreateSource(v.Data)
		if strings.Contains(createSource, f.CreateSource.Source) {
			result = append(result, v)
		}
	}
	return result
}

// createSourceTemplateNameFilter 模板名称来源过滤器
func (f *QueryFilter) createSourceTemplateNameFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.CreateSource.Template == nil {
		return resources
	}
	if f.CreateSource.Template.TemplateName == "" {
		return resources
	}

	for _, v := range resources {
		templateName := mapx.GetStr(v.Data, []string{"metadata", "annotations", resCsts.TemplateNameAnnoKey})
		if strings.Contains(templateName, f.CreateSource.Template.TemplateName) {
			result = append(result, v)
		}
	}
	return result
}

// createSourceTemplateVersionFilter 模板版本来源过滤器
func (f *QueryFilter) createSourceTemplateVersionFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.CreateSource.Template == nil {
		return resources
	}
	if f.CreateSource.Template.TemplateVersion == "" {
		return resources
	}

	for _, v := range resources {
		templateVersion := mapx.GetStr(v.Data, []string{"metadata", "annotations", resCsts.TemplateVersionAnnoKey})
		if strings.Contains(templateVersion, f.CreateSource.Template.TemplateVersion) {
			result = append(result, v)
		}
	}
	return result
}

// createSourceChartNameFilter ChartName来源过滤器
func (f *QueryFilter) createSourceChartNameFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.CreateSource.Chart == nil {
		return resources
	}
	if f.CreateSource.Chart.ChartName == "" {
		return resources
	}

	for _, v := range resources {
		chartName := mapx.GetStr(v.Data, []string{"metadata", "labels", resCsts.HelmChartAnnoKey})
		if strings.Contains(chartName, f.CreateSource.Chart.ChartName) {
			result = append(result, v)
		}
	}
	return result
}

// LabelSelectorFilter 标签选择器过滤器
func (f *QueryFilter) LabelSelectorFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if len(f.LabelSelector) == 0 {
		return resources
	}
	var requirements []*labels.Requirement
	for _, v := range f.LabelSelector {
		rq, err := labels.NewRequirement(v.Key, transOperator(v.Op), v.Values)
		if err != nil {
			continue
		}
		requirements = append(requirements, rq)
	}
	selector := labels.NewSelector()
	for _, rq := range requirements {
		selector = selector.Add(*rq)
	}
	for _, v := range resources {
		lbs := make(map[string]string)
		for k, v := range mapx.GetMap(v.Data, "metadata.labels") {
			lbs[k] = fmt.Sprintf("%v", v)
		}
		if !selector.Matches(labels.Set(lbs)) {
			continue
		}
		result = append(result, v)
	}
	return result
}

func transOperator(op string) selection.Operator {
	switch op {
	case OpEQ:
		return selection.Equals
	case OpNotEQ:
		return selection.NotEquals
	case OpIn:
		return selection.In
	case OpNotIn:
		return selection.NotIn
	case OpExists:
		return selection.Exists
	case OpDoesNotExist:
		return selection.DoesNotExist
	default:
		return selection.Equals
	}
}

// StatusFilter 状态过滤器
func (f *QueryFilter) StatusFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if len(f.Status) == 0 {
		return resources
	}
	if len(resources) == 0 {
		return resources
	}
	apiVersion := mapx.GetStr(resources[0].Data, "apiVersion")
	kind := resources[0].ResourceType
	formatFunc := formatter.GetFormatFunc(kind, apiVersion)
	for _, v := range resources {
		ext := formatFunc(v.Data)
		if ext == nil {
			continue
		}
		if status, ok := ext["status"].(string); ok {
			if slice.StringInSlice(status, f.Status) {
				result = append(result, v)
			}
		}
	}
	return result
}

// IPFilter IP过滤器
func (f *QueryFilter) IPFilter(resources []*storage.Resource) []*storage.Resource {
	result := []*storage.Resource{}
	if f.IP == "" {
		return resources
	}
	for _, v := range resources {
		// svc ip
		if strings.Contains(mapx.GetStr(v.Data, "spec.clusterIP"), f.IP) {
			result = append(result, v)
			continue
		}
		// svc 双栈
		for _, ip := range mapx.GetList(v.Data, "spec.clusterIPs") {
			if strings.Contains(ip.(string), f.IP) {
				result = append(result, v)
				break
			}
		}
		// pod host ip
		if strings.Contains(mapx.GetStr(v.Data, "status.hostIP"), f.IP) {
			result = append(result, v)
			continue
		}
		// pod 双栈
		for _, item := range mapx.GetList(v.Data, "status.podIPs") {
			ip := item.(map[string]interface{})["ip"].(string)
			if strings.Contains(ip, f.IP) {
				result = append(result, v)
				break
			}
		}
	}
	return result
}

// Sort 排序
func (f *QueryFilter) Sort(resources []*storage.Resource) {
	var sortInterface sort.Interface
	switch f.SortBy {
	case SortByNamespace:
		sortInterface = SortResourcesByNamespace(resources)
	case SortByAge:
		sortInterface = SortResourcesByAge(resources)
	default:
		sortInterface = SortResourcesByName(resources)
	}
	if f.Order == OrderDesc {
		sort.Sort(sort.Reverse(sortInterface))
		return
	}
	sort.Sort(sortInterface)
}

// SortResourcesByName 按照名称排序
type SortResourcesByName []*storage.Resource

func (s SortResourcesByName) Len() int {
	return len(s)
}

func (s SortResourcesByName) Less(i, j int) bool {
	return mapx.GetStr(s[i].Data, "metadata.name") < mapx.GetStr(s[j].Data, "metadata.name")
}

func (s SortResourcesByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortResourcesByNamespace 按照命名空间排序
type SortResourcesByNamespace []*storage.Resource

func (s SortResourcesByNamespace) Len() int {
	return len(s)
}

func (s SortResourcesByNamespace) Less(i, j int) bool {
	return mapx.GetStr(s[i].Data, "metadata.namespace") < mapx.GetStr(s[j].Data, "metadata.namespace")
}

func (s SortResourcesByNamespace) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortResourcesByAge 按照创建时间排序
type SortResourcesByAge []*storage.Resource

func (s SortResourcesByAge) Len() int {
	return len(s)
}

func (s SortResourcesByAge) Less(i, j int) bool {
	return mapx.GetStr(s[i].Data, "metadata.creationTimestamp") < mapx.GetStr(s[j].Data, "metadata.creationTimestamp")
}

func (s SortResourcesByAge) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Page 分页
func (f *QueryFilter) Page(resources []*storage.Resource) []*storage.Resource {
	f.Sort(resources)
	if f.Offset >= len(resources) {
		return []*storage.Resource{}
	}
	end := f.Offset + f.Limit
	if end > len(resources) {
		end = len(resources)
	}
	return resources[f.Offset:end]
}
