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

package metric

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

// getTable xxx
// metric data table is clusterId
func getTable(req *restful.Request) string {
	return req.PathParameter(clusterIDTag)
}

func getExtra(req *restful.Request) operator.M {
	extra := make(operator.M)
	raw := req.QueryParameter(extraTag)
	if raw == "" {
		return extra
	}

	_ = lib.NewExtra(raw).Unmarshal(&extra)
	return extra
}

func getMetricFeat(req *restful.Request) *operator.Condition {
	return getBaseFeat(req, metricFeatTags)
}

func getQueryFeat(req *restful.Request) *operator.Condition {
	condition := getBaseFeat(req, queryFeatTags)
	condList := []*operator.Condition{condition}
	for _, key := range queryExtraTags {
		if v := req.QueryParameter(key); v != "" {
			condList = append(condList, operator.NewLeafCondition(
				operator.In, operator.M{key: strings.Split(v, ",")}))
		}
	}
	return operator.NewBranchCondition(operator.And, condList...)
}

func getBaseFeatures(req *restful.Request, resourceFeatList []string) operator.M {
	features := make(operator.M, len(resourceFeatList))
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}

	// handle the extra field
	extra := getExtra(req)
	for k, v := range extra {
		features[k] = v
	}
	return features
}

func getBaseFeat(req *restful.Request, resourceFeatList []string) *operator.Condition {
	features := getBaseFeatures(req, resourceFeatList)
	return operator.NewLeafCondition(operator.Eq, features)
}

func getReqData(req *restful.Request, features operator.M) (operator.M, error) {
	var tmp types.BcsStorageMetricIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	data := lib.CopyMap(features)
	data[dataTag] = tmp.Data
	return data, nil
}

func getMetric(req *restful.Request) ([]operator.M, error) {
	return get(req, getMetricFeat(req))
}

func queryMetric(req *restful.Request) ([]operator.M, error) {
	return get(req, getQueryFeat(req))
}

func get(req *restful.Request, condition *operator.Condition) (result []operator.M, err error) {
	// 表名
	resourceType := getTable(req)
	// 参数
	var offset, limit int64

	if offset, err = lib.GetQueryParamInt64(req, offsetTag, 0); err != nil {
		return nil, err
	}

	if limit, err = lib.GetQueryParamInt64(req, limitTag, 0); err != nil {
		return nil, err
	}

	// option
	opt := &lib.StoreGetOption{
		Cond:   condition,
		Offset: offset,
		Limit:  limit,
	}

	return GetData(req.Request.Context(), resourceType, opt)
}

func put(req *restful.Request) error {
	// 表名
	resourceType := getTable(req)
	// 参数
	features := getBaseFeatures(req, metricFeatTags)
	data, err := getReqData(req, features)
	if err != nil {
		return err
	}

	// option
	opt := &lib.StorePutOption{
		Cond:          operator.NewLeafCondition(operator.Eq, features),
		UpdateTimeKey: updateTimeTag,
		CreateTimeKey: createTimeTag,
	}

	return PutData(req.Request.Context(), resourceType, data, opt)
}

func remove(req *restful.Request) error {
	// 表名
	resourceType := getTable(req)
	// 参数
	condition := getMetricFeat(req)

	// option
	opt := &lib.StoreRemoveOption{
		Cond: condition,
	}
	return RemoveData(req.Request.Context(), resourceType, opt)
}

func tables(req *restful.Request) ([]string, error) {
	return GetList(req.Request.Context())
}
