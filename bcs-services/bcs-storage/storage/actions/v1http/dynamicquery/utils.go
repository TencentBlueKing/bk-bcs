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

package dynamicquery

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

var (
	k8sNamespaceGrepNames = []string{
		"ReplicaSet", "Deployment", "Service", "ConfigMap", "Secret", "Ingress", "DaemonSet", "Job", "StatefulSet",
	}
	k8sNamespaceGrepQFilters = []qFilter{
		&ReplicaSetFilter{}, &DeploymentK8sFilter{}, &ServiceK8sFilter{}, &ConfigMapK8sFilter{}, &SecretK8sFilter{},
		&IngressFilter{}, &DaemonSetFilter{}, &JobFilter{}, &StatefulSetFilter{},
	}

	mesosNamespaceGrepNames = []string{
		"application", "application", "deployment", "service", "configmap", "secret",
	}
	mesosNamespaceGrepQFilters = []qFilter{
		&ApplicationFilter{Kind: ",application"}, &ProcessFilter{Kind: "process"}, &DeploymentFilter{}, &ServiceFilter{},
		&ConfigMapFilter{}, &SecretFilter{},
	}
)

// DynamicParams dynamic params
type DynamicParams struct {
	req       *restful.Request
	filter    qFilter
	name      string
	isGen     bool
	isPost    bool
	ctx       context.Context
	clusterId string

	offset    int
	limit     int
	table     string
	selector  []string
	condition *operator.Condition
	features  operator.M // nolint
	data      operator.M // nolint
	body      operator.M
}

// newReqDynamic xxx
// get a new instance of reqDynamic, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqDynamic(req *restful.Request, filter qFilter, name string) *DynamicParams {

	return &DynamicParams{
		req:    req,
		name:   name,
		ctx:    req.Request.Context(),
		filter: filter,
		offset: 0,
		limit:  0,
		isPost: req.Request.Method == "POST",
	}
}

// NewDynamic 创建 动态参数，兼容当前的使用
// data中必须有clusterID、field、offset、limit相关字段
func NewDynamic(ctx context.Context, filter qFilter, data operator.M, name string, selector []string, offset, limit int,
) *DynamicParams {
	return &DynamicParams{
		req:       nil,
		isPost:    true,
		ctx:       ctx,
		clusterId: data[constants.ClusterIDTag].(string), // 集群id

		filter:   filter,   // 模型
		name:     name,     // 表名
		table:    name,     // 表名
		body:     data,     // 数据
		selector: selector, // 查询字段
		// 分页
		offset: offset,
		limit:  limit,
	}
}

// reset clean the condition, data etc. so that the DynamicParams can be ready for
// next op.
// nolint
func (rd *DynamicParams) reset() {
	rd.condition = nil
	rd.selector = nil
	rd.data = nil
	rd.table = ""
}

func (rd *DynamicParams) getBody() operator.M {
	if rd.body == nil {
		if err := codec.DecJsonReader(rd.req.Request.Body, &rd.body); err != nil {
			blog.Errorf("get body failed: %v", err)
		}
	}
	return rd.body
}

func (rd *DynamicParams) getClusterId() string {
	if rd.clusterId == "" {
		rd.clusterId = rd.req.PathParameter(clusterIDTag)
	}
	return rd.clusterId
}

func (rd *DynamicParams) getParam(key string) (r string) {
	if rd.isPost {
		body := rd.getBody()
		r, _ = body[key].(string)
	} else {
		r = rd.req.QueryParameter(key)
	}
	return r
}

func (rd *DynamicParams) getQueryParamJSON() []byte {
	if rd.isPost {
		body := rd.getBody()

		body[clusterIDTag] = rd.getClusterId()
		var r []byte
		_ = codec.EncJson(body, &r)
		return r
	}

	query := rd.req.Request.URL.Query()
	query[clusterIDTag] = []string{rd.req.PathParameter(clusterIDTag)}
	return getQueryJSON(query)
}

// getTable xxx
// dynamic data table is like: "clusterId_Type", for instance: "BCS-K8S-10001_Deployment".
// rd.table will be save since first call, so reset() should be called if doing another op.
func (rd *DynamicParams) getTable() string {
	if rd.table == "" {
		rd.table = rd.name
	}
	return rd.table
}

// getSelector return a slice of string contains select key for db query.
// rd.selector will be save since first call, so reset() should be called if doing another op.
func (rd *DynamicParams) getSelector() []string {
	if rd.selector == nil {
		s := rd.getParam(fieldTag)
		if len(s) != 0 {
			rd.selector = strings.Split(s, ",")
		}
	}
	return rd.selector
}

func (rd *DynamicParams) getOffset() int64 {
	s := rd.getParam(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.offset = r
	}
	return int64(rd.offset)
}

func (rd *DynamicParams) getLimit() int64 {
	s := rd.getParam(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.limit = r
	}
	return int64(rd.limit)
}

func (rd *DynamicParams) getExtra() operator.M {
	raw := rd.getParam(extraTag)
	if raw == "" {
		return nil
	}

	var extra operator.M
	_ = lib.NewExtra(raw).Unmarshal(&extra)
	return extra
}

func (rd *DynamicParams) getFeat() *operator.Condition {
	if rd.condition == nil {
		r := rd.filter.GetCondition()

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

func (rd *DynamicParams) generateFilter() error {
	if !rd.isGen {
		if err := codec.DecJson(rd.getQueryParamJSON(), &(rd.filter)); err != nil {
			return err
		}
		rd.isGen = true
	}
	return nil
}

// QueryDynamic query dynamic
func (rd *DynamicParams) QueryDynamic() ([]operator.M, error) {
	// 生成过滤条件
	if err := rd.generateFilter(); err != nil {
		return nil, err
	}
	// 构建查询条件 和 执行查询
	return rd.get(rd.getFeat())
}

func (rd *DynamicParams) get(condition *operator.Condition) (mList []operator.M, err error) {
	// option
	opt := &lib.StoreGetOption{
		Fields: rd.getSelector(),
		Offset: rd.getOffset(),
		Limit:  rd.getLimit(),
		Cond:   condition,
	}
	// blog.Infof("opt: %v", util.PrettyStruct(opt))
	// blog.Infof("filter: %v", util.PrettyStruct(rd.filter))

	if mList, err = GetData(rd.ctx, rd.getTable(), opt); err != nil {
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
	return p
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

func grepNamespace(req *restful.Request, filter qFilter, name string, origin []string) ([]string, error) {
	request := newReqDynamic(req, filter, name)
	r, err := request.QueryDynamic()
	if err != nil {
		return nil, err
	}
	return fetchNamespace(r, origin), nil
}

func getMesosNamespaceResource(req *restful.Request, resp *restful.Response) (err error) {
	// init Form
	req.Request.FormValue("")
	req.Request.Form[fieldTag] = []string{namespaceTag}
	result := make([]string, 0)

	for i, name := range mesosNamespaceGrepNames {
		// grep replicaSet
		if result, err = grepNamespace(req, mesosNamespaceGrepQFilters[i], name, result); err != nil {
			blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
			lib.ReturnRest(&lib.RestResponse{
				Resp: resp, Data: []string{},
				ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
			return err
		}
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: result})
	return nil
}

// getK8sNamespaceResource get namespace k8s used
func getK8sNamespaceResource(req *restful.Request, resp *restful.Response) error {
	// init Form
	req.Request.FormValue("")
	req.Request.Form[fieldTag] = []string{namespaceTag}
	var err error
	result := make([]string, 0)

	for i, name := range k8sNamespaceGrepNames {
		// grep replicaSet
		if result, err = grepNamespace(req, k8sNamespaceGrepQFilters[i], name, result); err != nil {
			blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
			lib.ReturnRest(&lib.RestResponse{
				Resp: resp, Data: []string{},
				ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
			return err
		}
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: result})
	return nil
}

// grepNamespaceResource
func grepNamespaceResource(request *DynamicParams, origin []string) ([]string, error) {
	r, err := request.QueryDynamic()
	if err != nil {
		return nil, err
	}
	return fetchNamespace(r, origin), nil
}

// GetK8sNamespaceResource get namespace k8s used
func GetK8sNamespaceResource(ctx context.Context, raw operator.M, offset, limit int) (result []string, err error) {
	result = make([]string, 0)

	req := &DynamicParams{
		req:       nil,
		isPost:    true,
		ctx:       ctx,
		clusterId: raw[constants.ClusterIDTag].(string), // 集群id
		body:      raw,                                  // 数据
		selector:  raw[fieldTag].([]string),             // 查询字段
		// 分页
		offset: offset,
		limit:  limit,
	}

	for i, name := range k8sNamespaceGrepNames {
		req.name = name
		req.table = name
		req.filter = k8sNamespaceGrepQFilters[i]
		if result, err = grepNamespaceResource(req, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetMesosNamespaceResource get namespace mesos used
func GetMesosNamespaceResource(ctx context.Context, raw operator.M, offset, limit int) (result []string, err error) {
	result = make([]string, 0)

	req := &DynamicParams{
		req:       nil,
		isPost:    true,
		ctx:       ctx,
		clusterId: raw[constants.ClusterIDTag].(string), // 集群id
		body:      raw,                                  // 数据
		selector:  raw[fieldTag].([]string),             // 查询字段
		// 分页
		offset: offset,
		limit:  limit,
	}

	for i, name := range mesosNamespaceGrepNames {
		req.name = name
		req.table = name
		req.filter = mesosNamespaceGrepQFilters[i]
		if result, err = grepNamespaceResource(req, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
