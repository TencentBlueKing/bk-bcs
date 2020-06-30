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
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

// reqMetric define a unit for operating metric data based
// on restful.Request.
type reqMetric struct {
	req  *restful.Request
	tank operator.Tank

	offset    int
	limit     int
	table     string
	selector  []string
	condition *operator.Condition
	features  operator.M
	data      operator.M
}

func newReqMetric(req *restful.Request) *reqMetric {
	return &reqMetric{
		req:    req,
		tank:   getNewTank(),
		offset: 0,
		limit:  -1,
	}
}

// reset clean the condition, data etc. so that the reqMetric can be ready for
// next op.
func (rm *reqMetric) reset() {
	rm.condition = nil
	rm.features = nil
	rm.data = nil
}

// metric data table is clusterId
func (rm *reqMetric) getTable() string {
	if rm.table == "" {
		rm.table = rm.req.PathParameter(clusterIdTag)
	}
	return rm.table
}

// getSelector return a slice of string contains select key for db query.
// rm.selector will be save since first call, so reset() should be called if doing another op.
func (rm *reqMetric) getSelector() []string {
	if rm.selector == nil {
		s := rm.req.QueryParameter(fieldTag)
		rm.selector = strings.Split(s, ",")
	}
	return rm.selector
}

func (rm *reqMetric) getOffset() int {
	s := rm.req.QueryParameter(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rm.offset = r
	}
	return rm.offset
}

func (rm *reqMetric) getLimit() int {
	s := rm.req.QueryParameter(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rm.limit = r
	}
	return rm.limit
}

func (rm *reqMetric) getExtra() (extra operator.M) {
	raw := rm.req.QueryParameter(extraTag)
	if raw == "" {
		return
	}

	lib.NewExtra(raw).Unmarshal(&extra)
	return
}

func (rm *reqMetric) getMetricFeat() *operator.Condition {
	if rm.condition == nil {
		rm.condition = rm.getBaseFeat(metricFeatTags)
	}
	return rm.condition
}

func (rm *reqMetric) getQueryFeat() *operator.Condition {
	if rm.condition == nil {
		condition := rm.getBaseFeat(queryFeatTags)
		for _, key := range queryExtraTags {
			if v := rm.req.QueryParameter(key); v != "" {
				condition = condition.AddOp(operator.In, key, strings.Split(v, ","))
			}
		}
		rm.condition = condition
	}
	return rm.condition
}

func (rm *reqMetric) getBaseFeat(resourceFeatList []string) *operator.Condition {
	features := make(operator.M, len(resourceFeatList))
	for _, key := range resourceFeatList {
		features[key] = rm.req.PathParameter(key)
	}
	rm.features = features

	// handle the extra field
	extra := rm.getExtra()
	for k, v := range extra {
		features[k] = v
	}

	return operator.NewCondition(operator.Eq, features)
}

func (rm *reqMetric) getReqData() (operator.M, error) {
	if rm.data == nil {
		var tmp types.BcsStorageMetricIf
		if err := codec.DecJsonReader(rm.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		data := lib.CopyMap(rm.features)
		data[dataTag] = tmp.Data
		rm.data = data
	}
	return rm.data, nil
}

func (rm *reqMetric) getMetric() ([]interface{}, error) {
	return rm.get(rm.getMetricFeat())
}

func (rm *reqMetric) queryMetric() ([]interface{}, error) {
	return rm.get(rm.getQueryFeat())
}

func (rm *reqMetric) get(condition *operator.Condition) (r []interface{}, err error) {
	tank := rm.tank.From(rm.getTable()).Filter(condition).
		Offset(rm.getOffset()).Limit(rm.getLimit()).Select(rm.getSelector()...).Query()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return
	}

	r = tank.GetValue()

	// Some time-field need to be format before return
	for i := range r {
		for _, t := range needTimeFormatList {
			tmp, ok := r[i].(map[string]interface{})[t].(time.Time)
			if !ok {
				continue
			}
			r[i].(map[string]interface{})[t] = tmp.Format(timeLayout)
		}
	}
	return
}

func (rm *reqMetric) put() (err error) {
	tank := rm.tank.From(rm.getTable()).Filter(rm.getMetricFeat())

	data, err := rm.getReqData()
	if err != nil {
		return
	}

	// Update or insert
	timeNow := time.Now()

	queryTank := tank.Query()
	if err = queryTank.GetError(); err != nil {
		blog.Errorf("Failed to check if resource exist. err: %v", err)
		return
	}
	if queryTank.GetLen() == 0 {
		data.Update(createTimeTag, timeNow)
	}

	tank = tank.Index(indexKeys...).Upsert(data.Update(updateTimeTag, timeNow))
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to update. err: %v", err)
		return
	}
	return
}

func (rm *reqMetric) remove() (err error) {
	tank := rm.tank.From(rm.getTable()).Filter(rm.getMetricFeat()).RemoveAll()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to remove. err: %v", err)
		return
	}
	if changeInfo := tank.GetChangeInfo(); changeInfo.Removed == 0 {
		return storageErr.ResourceDoesNotExist
	}
	return
}

func (rm *reqMetric) tables() (r []interface{}, err error) {
	tank := rm.tank.Tables()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to list tables. err: %v", err)
		return
	}

	r = tank.GetValue()
	return
}

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (rm *reqMetric) exit() {
	if rm.tank != nil {
		rm.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return oldUrl
}
