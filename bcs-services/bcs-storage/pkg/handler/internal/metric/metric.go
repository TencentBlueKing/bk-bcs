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
 *
 */

package metric

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/metric"
)

var (
	queryFeatTags  = []string{constants.ClusterIDTag}
	queryExtraTags = []string{constants.NamespaceTag, constants.TypeTag, constants.NameTag}
	metricFeatTags = []string{constants.ClusterIDTag, constants.NamespaceTag, constants.TypeTag, constants.NameTag}
)

type general struct {
	ctx      context.Context
	featList []string
	raw      operator.M
	metric   *storage.Metric
	//
	extra  string
	offset int64
	limit  int64
	data   operator.M
}

func (g *general) getExtra() operator.M {
	extra := make(operator.M)
	if g.extra == "" {
		return extra
	}

	lib.NewExtra(g.extra).Unmarshal(&extra)
	return extra
}

func (g *general) getBaseFeatures() operator.M {
	features := make(operator.M, len(g.featList))
	for _, key := range g.featList {
		features[key] = g.raw[key]
	}

	// handle the extra field
	extra := g.getExtra()
	for k, v := range extra {
		features[k] = v
	}
	return features
}

func (g *general) getBaseFeat() *operator.Condition {
	features := g.getBaseFeatures()
	return operator.NewLeafCondition(operator.Eq, features)
}

func (g *general) getMetricFeat() *operator.Condition {
	return g.getBaseFeat()
}

func (g *general) getMetric() ([]operator.M, error) {
	condition := g.getMetricFeat()
	resourceType := g.metric.ClusterId

	opt := &lib.StoreGetOption{
		Cond:   condition,
		Offset: g.offset,
		Limit:  g.limit,
	}

	return metric.GetData(g.ctx, resourceType, opt)
}

func (g *general) getReqData(features operator.M) operator.M {
	data := lib.CopyMap(features)
	data[constants.DataTag] = g.data
	return data
}

func (g *general) putMetric() error {
	resourceType := g.metric.ClusterId
	features := g.getBaseFeatures()
	data := g.getReqData(features)

	opt := &lib.StorePutOption{
		Cond:          operator.NewLeafCondition(operator.Eq, features),
		UpdateTimeKey: constants.UpdateTimeTag,
		CreateTimeKey: constants.CreateTimeTag,
	}

	return metric.PutData(g.ctx, resourceType, data, opt)
}

func (g *general) removeMetric() error {
	resourceType := g.metric.ClusterId
	condition := g.getMetricFeat()

	opt := &lib.StoreRemoveOption{
		Cond: condition,
	}

	return metric.RemoveData(g.ctx, resourceType, opt)
}

func (g *general) getQueryFeat() *operator.Condition {
	condition := g.getBaseFeat()
	condList := []*operator.Condition{condition}

	for _, key := range queryExtraTags {
		if v := g.raw[key]; v != "" {
			condList = append(
				condList,
				operator.NewLeafCondition(
					operator.In,
					operator.M{
						key: strings.Split(v.(string), ","),
					},
				),
			)
		}
	}

	return operator.NewBranchCondition(operator.And, condList...)
}

func (g *general) queryMetric() ([]operator.M, error) {
	condition := g.getQueryFeat()
	resourceType := g.metric.ClusterId

	opt := &lib.StoreGetOption{
		Cond:   condition,
		Offset: g.offset,
		Limit:  g.limit,
	}

	return metric.GetData(g.ctx, resourceType, opt)
}

// HandlerGetMetric 获取指标
func HandlerGetMetric(ctx context.Context, req *storage.GetMetricRequest) ([]operator.M, error) {
	m := &storage.Metric{
		Name:      req.Name,
		Type:      req.Type,
		Namespace: req.Namespace,
		ClusterId: req.ClusterId,
	}
	g := &general{
		ctx:      ctx,
		extra:    req.Extra,
		metric:   m,
		featList: metricFeatTags,
		limit:    int64(req.Limit),
		offset:   int64(req.Offset),
		raw:      util.StructToMap(m),
	}

	return g.getMetric()
}

// HandlerPutMetric 创建指标
func HandlerPutMetric(ctx context.Context, req *storage.PutMetricRequest) error {
	m := &storage.Metric{
		Name:      req.Name,
		Type:      req.Type,
		Namespace: req.Namespace,
		ClusterId: req.ClusterId,
	}
	g := &general{
		ctx:      ctx,
		extra:    req.Extra,
		metric:   m,
		featList: metricFeatTags,
		raw:      util.StructToMap(m),
		data:     util.StructToMap(req.Data),
	}

	return g.putMetric()
}

// HandlerDeleteMetric 删除指标
func HandlerDeleteMetric(ctx context.Context, req *storage.DeleteMetricRequest) error {
	m := &storage.Metric{
		Name:      req.Name,
		Type:      req.Type,
		Namespace: req.Namespace,
		ClusterId: req.ClusterId,
	}
	g := &general{
		ctx:      ctx,
		metric:   m,
		featList: metricFeatTags,
		raw:      util.StructToMap(m),
	}

	return g.removeMetric()
}

// HandlerQueryMetric 获取指标
func HandlerQueryMetric(ctx context.Context, req *storage.QueryMetricRequest) ([]operator.M, error) {
	m := &storage.Metric{
		Name:      req.Name,
		Type:      req.Type,
		Namespace: req.Namespace,
		ClusterId: req.ClusterId,
	}
	g := &general{
		ctx:      ctx,
		extra:    req.Extra,
		metric:   m,
		featList: queryFeatTags,
		limit:    int64(req.Limit),
		offset:   int64(req.Offset),
		raw:      util.StructToMap(m),
	}

	return g.queryMetric()
}

// HandlerListMetricTables 获取所有表名
func HandlerListMetricTables(ctx context.Context) ([]string, error) {
	return metric.GetList(ctx)
}
