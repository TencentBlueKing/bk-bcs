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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/samber/lo"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

const (
	// ConName 资源名称
	ConName string = "data.metadata.name"
	// ConNamespace 资源命名空间
	ConNamespace string = "data.metadata.namespace"
	// ConAge 资源创建时间
	ConAge string = "data.metadata.creationTimestamp"
	// ConLabels 资源标签
	ConLabels string = "data.metadata.labels"
	// ConAnnotations 资源注解
	ConAnnotations string = "data.metadata.annotations"
)

// creatorToCondition creator condition
func (q *StorageQuery) creatorToCondition() []*operator.Condition {
	conditions := []*operator.Condition{}
	if len(q.QueryFilter.Creator) == 0 && len(q.ViewFilter.Creator) == 0 {
		return conditions
	}
	// 无创建者的情况比较特殊
	viewCreator := q.ViewFilter.Creator
	queryCreator := q.QueryFilter.Creator
	if lo.Contains(viewCreator, EmptyCreator) {
		viewCreator = []string{EmptyCreator}
	}
	if lo.Contains(queryCreator, EmptyCreator) {
		queryCreator = []string{EmptyCreator}
	}
	// creator 过滤条件
	var creator []string
	switch {
	case len(viewCreator) == 0:
		creator = queryCreator
	case len(queryCreator) == 0:
		creator = viewCreator
	default:
		// 返回交集创建人
		creator = lo.Intersect(viewCreator, queryCreator)
	}

	if lo.Contains(creator, EmptyCreator) {
		// 使用全角符号代替 '.',区分字段分隔，无创建者的资源，creator字段为 null
		conditions = append(conditions, operator.NewLeafCondition(
			operator.Eq, map[string]interface{}{
				ConAnnotations + "." + mapx.ConvertPath(LabelCreator): nil}))
	} else {
		conditions = append(conditions, operator.NewLeafCondition(
			operator.In, map[string]interface{}{
				ConAnnotations + "." + mapx.ConvertPath(LabelCreator): creator}))
	}
	return conditions
}

// nameToCondition name condition
func (q *StorageQuery) nameToCondition() []*operator.Condition {
	conditions := []*operator.Condition{}
	// creator 过滤条件
	if len(q.QueryFilter.Name) != 0 && len(q.ViewFilter.Name) != 0 {
		con := operator.NewBranchCondition(
			operator.And,
			operator.NewLeafCondition(operator.Con, map[string]string{ConName: q.QueryFilter.Name}),
			operator.NewLeafCondition(operator.Con, map[string]string{ConName: q.ViewFilter.Name}),
		)
		conditions = append(conditions, con)
		return conditions
	}
	if len(q.ViewFilter.Name) != 0 {
		conditions = append(conditions, operator.NewLeafCondition(
			operator.Con, map[string]string{ConName: q.ViewFilter.Name}))
		return conditions
	}
	if len(q.QueryFilter.Name) != 0 {
		conditions = append(conditions, operator.NewLeafCondition(
			operator.Con, map[string]string{ConName: q.QueryFilter.Name}))
		return conditions
	}
	return conditions
}

// labelToCondition label condition
func (q *StorageQuery) labelToCondition() []*operator.Condition {
	conditions := []*operator.Condition{}
	// nolint:gocritic
	labelSelector := append(q.ViewFilter.LabelSelector, q.QueryFilter.LabelSelector...)
	// labelSelector 过滤条件
	for _, v := range labelSelector {
		if len(v.Values) == 0 && v.Op != OpExists && v.Op != OpDoesNotExist {
			continue
		}
		switch v.Op {
		case OpEQ:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): v.Values[0]}))
		case OpNotEQ:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Not, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): v.Values[0]}))
		case OpIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.In, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpNotIn:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Nin, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): v.Values}))
		case OpExists:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Ext, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): ""}))
		case OpDoesNotExist:
			conditions = append(conditions, operator.NewLeafCondition(
				operator.Eq, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(v.Key)): nil}))
		}
	}
	return conditions
}

// createSourceToCondition create source condition
func (q *StorageQuery) createSourceToCondition() []*operator.Condition {
	conditions := []*operator.Condition{}
	if q.ViewFilter.CreateSource == nil && q.QueryFilter.CreateSource == nil {
		return conditions
	}

	var createSources []*clusterRes.CreateSource

	if q.ViewFilter.CreateSource != nil {
		createSources = append(createSources, q.ViewFilter.CreateSource)
	}
	if q.QueryFilter.CreateSource != nil {
		createSources = append(createSources, q.QueryFilter.CreateSource)
	}

	// 创建来源source筛选
	templateCon := operator.NewBranchCondition(
		operator.Or,
		operator.NewLeafCondition(
			operator.Eq, map[string]interface{}{fmt.Sprintf(ConAnnotations+".%s",
				mapx.ConvertPath(constants.TemplateSourceType)): constants.TemplateSourceTypeValue}),
		operator.NewLeafCondition(
			operator.Eq, map[string]interface{}{fmt.Sprintf(ConLabels+".%s",
				mapx.ConvertPath(constants.TemplateSourceType)): constants.TemplateSourceTypeValue}),
	)
	helmCon := operator.NewLeafCondition(
		operator.Eq, map[string]interface{}{fmt.Sprintf(ConLabels+".%s",
			mapx.ConvertPath(constants.HelmSourceType)): constants.HelmCreateSource})

	// 要排除掉helm和template的情况
	webCon := operator.NewBranchCondition(
		operator.And,
		operator.NewBranchCondition(
			operator.Nor,
			templateCon,
			helmCon,
		),
		operator.NewBranchCondition(
			operator.Or,
			operator.NewLeafCondition(operator.Ne, map[string]interface{}{
				fmt.Sprintf(ConAnnotations+".%s", mapx.ConvertPath(constants.CreatorAnnoKey)): nil}),
			operator.NewLeafCondition(operator.Ne, map[string]interface{}{
				fmt.Sprintf(ConLabels+".%s", mapx.ConvertPath(constants.UpdaterAnnoKey)): nil})),
	)
	for _, cs := range createSources {
		switch cs.Source {
		case constants.TemplateCreateSource:
			// 筛选模板名称和模板版本
			var temNameVersionCon *operator.Condition
			if cs.Template != nil {
				var temNameCon []*operator.Condition
				// template 需要筛选名称和版本
				if cs.Template.TemplateName != "" {
					temNameCon = append(temNameCon, operator.NewLeafCondition(operator.Con, map[string]interface{}{
						fmt.Sprintf(ConAnnotations+".%s",
							mapx.ConvertPath(constants.TemplateNameAnnoKey)): cs.Template.TemplateName}))

				}
				if cs.Template.TemplateVersion != "" {
					temNameCon = append(temNameCon, operator.NewLeafCondition(operator.Con, map[string]interface{}{
						fmt.Sprintf(ConAnnotations+".%s",
							mapx.ConvertPath(constants.TemplateVersionAnnoKey)): cs.Template.TemplateVersion}))
				}
				temNameCon = append(temNameCon, templateCon)
				temNameVersionCon = operator.NewBranchCondition(operator.And, temNameCon...)
			} else {
				temNameVersionCon = templateCon
			}
			conditions = append(conditions, temNameVersionCon)
		case constants.HelmCreateSource:
			var helmChartCon *operator.Condition
			// helm 筛选chart名称
			if cs.Chart != nil && cs.Chart.ChartName != "" {
				chartCon := operator.NewLeafCondition(operator.Con, map[string]interface{}{
					fmt.Sprintf(ConLabels+".%s",
						mapx.ConvertPath(constants.HelmChartAnnoKey)): cs.Chart.ChartName})
				helmChartCon = operator.NewBranchCondition(operator.And, helmCon, chartCon)
			} else {
				helmChartCon = helmCon
			}
			conditions = append(conditions, helmChartCon)
		case constants.WebCreateSource:
			conditions = append(conditions, webCon)
		case constants.ClientCreateSource:
			conditions = append(conditions, operator.NewBranchCondition(
				operator.Nor, templateCon, helmCon, webCon,
			))
		}
	}

	return conditions
}

// ipToCondition ip condition
func (q *StorageQuery) ipToCondition() []*operator.Condition {
	conditions := []*operator.Condition{}
	if q.QueryFilter.IP == "" {
		return conditions
	}
	// ip 过滤条件
	con := operator.NewBranchCondition(
		operator.Or,
		operator.NewLeafCondition(operator.Con, map[string]string{"data.spec.clusterIP": q.QueryFilter.IP}),
		operator.NewLeafCondition(operator.Eq,
			operator.M{"data.spec.clusterIPs": operator.M{"$elemMatch": operator.M{"$regex": q.QueryFilter.IP}}}),
		operator.NewLeafCondition(operator.Con, map[string]string{"data.status.hostIP": q.QueryFilter.IP}),
		operator.NewLeafCondition(operator.Eq, operator.M{"data.status.podIPs": operator.M{"$elemMatch": operator.
			M{"ip": operator.M{"$regex": q.QueryFilter.IP}}}}),
	)
	conditions = append(conditions, con)
	return conditions
}
