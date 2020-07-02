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

package dynamicquery

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

type reqDynamic struct {
	req    *restful.Request
	tank   operator.Tank
	filter qFilter
	name   string
	isGen  bool
	isPost bool

	offset    int
	limit     int
	table     string
	selector  []string
	condition *operator.Condition
	features  operator.M
	data      operator.M
	body      operator.M
}

// get a new instance of reqDynamic, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqDynamic(req *restful.Request, filter qFilter, name string) *reqDynamic {
	return &reqDynamic{
		req:    req,
		tank:   getNewTank(),
		name:   name,
		filter: filter,
		offset: 0,
		limit:  -1,
		isPost: req.Request.Method == "POST",
	}
}

// reset clean the condition, data etc. so that the reqDynamic can be ready for
// next op.
func (rd *reqDynamic) reset() {
	rd.condition = nil
	rd.selector = nil
	rd.data = nil
	rd.table = ""
}

func (rd *reqDynamic) getBody() operator.M {
	if rd.body == nil {
		if err := codec.DecJsonReader(rd.req.Request.Body, &rd.body); err != nil {
			blog.Errorf("get body failed: %v", err)
		}
	}
	return rd.body
}

func (rd *reqDynamic) getParam(key string) (r string) {
	if rd.isPost {
		body := rd.getBody()
		r, _ = body[key].(string)
	} else {
		r = rd.req.QueryParameter(key)
	}
	return
}

func (rd *reqDynamic) getQueryParamJson() []byte {
	if rd.isPost {
		body := rd.getBody()
		body[clusterIdTag] = rd.req.PathParameter(clusterIdTag)
		var r []byte
		_ = codec.EncJson(body, &r)
		return r
	}

	query := rd.req.Request.URL.Query()
	query[clusterIdTag] = []string{rd.req.PathParameter(clusterIdTag)}
	return getQueryJson(query)
}

// dynamic data table is like: "clusterId_Type", for instance: "BCS-K8S-10001_Deployment".
// rd.table will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getTable() string {
	if rd.table == "" {
		rd.table = rd.req.PathParameter(clusterIdTag) + "_" + rd.name
	}
	return rd.table
}

// getSelector return a slice of string contains select key for db query.
// rd.selector will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getSelector() []string {
	if rd.selector == nil {
		s := rd.getParam(fieldTag)
		rd.selector = strings.Split(s, ",")
	}
	return rd.selector
}

func (rd *reqDynamic) getOffset() int {
	s := rd.getParam(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.offset = r
	}
	return rd.offset
}

func (rd *reqDynamic) getLimit() int {
	s := rd.getParam(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.limit = r
	}
	return rd.limit
}

func (rd *reqDynamic) getExtra() (extra operator.M) {
	raw := rd.getParam(extraTag)
	if raw == "" {
		return
	}

	lib.NewExtra(raw).Unmarshal(&extra)
	return
}

func (rd *reqDynamic) getFeat() *operator.Condition {
	if rd.condition == nil {
		r := rd.filter.getCondition()

		// handle the extra field
		extra := rd.getExtra()
		features := make(operator.M)
		for k, v := range extra {
			if _, ok := v.([]interface{}); !ok {
				features[k] = []interface{}{v}
				continue
			}
			features[k] = v
		}

		if len(features) > 0 {
			r = r.And(operator.NewCondition(operator.In, features))
		}
		rd.condition = r
	}

	return rd.condition
}

func (rd *reqDynamic) generateFilter() error {
	if !rd.isGen {
		if err := codec.DecJson(rd.getQueryParamJson(), &(rd.filter)); err != nil {
			return err
		}
		rd.isGen = true
	}
	return nil
}

func (rd *reqDynamic) queryDynamic() ([]interface{}, error) {
	if err := rd.generateFilter(); err != nil {
		return nil, err
	}
	return rd.get(rd.getFeat())
}

func (rd *reqDynamic) get(condition *operator.Condition) (r []interface{}, err error) {
	tank := rd.tank.From(rd.getTable()).Filter(condition).
		Offset(rd.getOffset()).Limit(rd.getLimit()).Select(rd.getSelector()...).Query()

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

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (rd *reqDynamic) exit() {
	if rd.tank != nil {
		rd.tank.Close()
	}
}

func getQueryJson(s url.Values) (p []byte) {
	r := make(map[string]string)
	for k, v := range s {
		if len(v) > 0 {
			r[k] = v[0]
		}
	}
	_ = codec.EncJson(r, &p)
	return
}

func urlPath(oldUrl string) string {
	return urlPrefix + oldUrl
}

func fetchNamespace(r []interface{}, result []string) []string {
	if result == nil {
		result = make([]string, 0)
	}

	nsMap := make(map[string]bool, 0)
	for _, ns := range result {
		nsMap[ns] = false
	}

	for _, item := range r {
		mapItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		ns, ok := mapItem[namespaceTag]
		if !ok {
			continue
		}

		nsStr, ok := ns.(string)
		if !ok {
			continue
		}

		if _, ok := nsMap[nsStr]; !ok {
			nsMap[nsStr] = true
		}
	}

	for key, value := range nsMap {
		if value {
			result = append(result, key)
		}
	}
	return result
}
