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
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"

	"github.com/emicklei/go-restful"
)

type reqDynamic struct {
	req    *restful.Request
	store  *lib.Store
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
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	store.SetSoftDeletion(true)
	return &reqDynamic{
		req:    req,
		store:  store,
		name:   name,
		filter: filter,
		offset: 0,
		limit:  0,
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

func (rd *reqDynamic) getQueryParamJSON() []byte {
	if rd.isPost {
		body := rd.getBody()
		body[clusterIDTag] = rd.req.PathParameter(clusterIDTag)
		var r []byte
		_ = codec.EncJson(body, &r)
		return r
	}

	query := rd.req.Request.URL.Query()
	query[clusterIDTag] = []string{rd.req.PathParameter(clusterIDTag)}
	return getQueryJSON(query)
}

// dynamic data table is like: "clusterId_Type", for instance: "BCS-K8S-10001_Deployment".
// rd.table will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getTable() string {
	if rd.table == "" {
		rd.table = rd.name
	}
	return rd.table
}

// getSelector return a slice of string contains select key for db query.
// rd.selector will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getSelector() []string {
	if rd.selector == nil {
		s := rd.getParam(fieldTag)
		if len(s) != 0 {
			rd.selector = strings.Split(s, ",")
		}
	}
	return rd.selector
}

func (rd *reqDynamic) getOffset() int64 {
	s := rd.getParam(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.offset = r
	}
	return int64(rd.offset)
}

func (rd *reqDynamic) getLimit() int64 {
	s := rd.getParam(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.limit = r
	}
	return int64(rd.limit)
}

func (rd *reqDynamic) getExtra() operator.M {
	raw := rd.getParam(extraTag)
	if raw == "" {
		return nil
	}

	var extra operator.M
	lib.NewExtra(raw).Unmarshal(&extra)
	return extra
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
			r = operator.NewBranchCondition(operator.And, r, operator.NewLeafCondition(operator.In, features))
		}
		rd.condition = r
	}

	return rd.condition
}

func (rd *reqDynamic) generateFilter() error {
	if !rd.isGen {
		if err := codec.DecJson(rd.getQueryParamJSON(), &(rd.filter)); err != nil {
			return err
		}
		rd.isGen = true
	}
	return nil
}

func (rd *reqDynamic) queryDynamic() ([]operator.M, error) {
	if err := rd.generateFilter(); err != nil {
		return nil, err
	}
	return rd.get(rd.getFeat())
}

func (rd *reqDynamic) get(condition *operator.Condition) ([]operator.M, error) {
	getOption := &lib.StoreGetOption{
		Fields: rd.getSelector(),
		Offset: rd.getOffset(),
		Limit:  rd.getLimit(),
		Cond:   condition,
	}

	mList, err := rd.store.Get(rd.req.Request.Context(), rd.getTable(), getOption)
	if err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return nil, fmt.Errorf("failed to query. err: %v", err)
	}

	return mList, nil
}

func getQueryJSON(s url.Values) (p []byte) {
	r := make(map[string]string)
	for k, v := range s {
		if len(v) > 0 {
			r[k] = v[0]
		}
	}
	_ = codec.EncJson(r, &p)
	return
}

func urlPath(oldURL string) string {
	return urlPrefix + oldURL
}

func fetchNamespace(r []operator.M, result []string) []string {
	if result == nil {
		result = make([]string, 0)
	}

	nsMap := make(map[string]bool, 0)
	for _, ns := range result {
		nsMap[ns] = false
	}

	for _, item := range r {
		ns, ok := item[namespaceTag]
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
