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
	"fmt"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
)

const (
	clusterIDTag  = "clusterId"
	namespaceTag  = "namespace"
	typeTag       = "type"
	nameTag       = "name"
	dataTag       = "data"
	extraTag      = "extra"
	fieldTag      = "field"
	offsetTag     = "offset"
	limitTag      = "limit"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
	timeLayout    = "2006-01-02 15:04:05"
)

var needTimeFormatList = [...]string{updateTimeTag, createTimeTag}
var metricFeatTags = []string{clusterIDTag, namespaceTag, typeTag, nameTag}
var queryFeatTags = []string{clusterIDTag}
var queryExtraTags = []string{namespaceTag, typeTag, nameTag}
var indexKeys = []string{clusterIDTag, namespaceTag, typeTag, nameTag}

// Use Mongodb for storage.
const dbConfig = "mongodb/metric"

// GetMetric get metric
func GetMetric(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetMetric"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := getMetric(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		if err == storageErr.ResourceDoesNotExist {
			lib.ReturnRest(&lib.RestResponse{
				Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist,
				Message: common.BcsErrStorageResourceNotExistStr})
			return
		}
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	if len(r) == 0 {
		err := fmt.Errorf("resource does not exist")
		utils.SetSpanLogTagError(span, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist,
			Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r[0]})
}

// PutMetric put metric
func PutMetric(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutMetric"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := put(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteMetric delete metric
func DeleteMetric(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteMetric"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := remove(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		if err == storageErr.ResourceDoesNotExist {
			lib.ReturnRest(&lib.RestResponse{
				Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist,
				Message: common.BcsErrStorageResourceNotExistStr})
			return
		}
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// QueryMetric query metric
func QueryMetric(req *restful.Request, resp *restful.Response) {
	const (
		handler = "QueryMetric"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := queryMetric(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail,
			Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// ListMetricTables list metric tables
func ListMetricTables(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListMetricTables"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := tables(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDecodeListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageDecodeListResourceFail,
			Message: common.BcsErrStorageDecodeListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func init() {
	metricPath := "/metric/clusters/{clusterId}/namespaces/{namespace}/{type}/{name}"
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: metricPath, Params: nil, Handler: lib.MarkProcess(GetMetric)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: metricPath, Params: nil, Handler: lib.MarkProcess(PutMetric)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: metricPath, Params: nil, Handler: lib.MarkProcess(DeleteMetric)})

	listMetricPath := "/metric/clusters/{clusterId}"
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(QueryMetric)})

	listMetricTablePath := "/metric/clusters"
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listMetricTablePath, Params: nil, Handler: lib.MarkProcess(ListMetricTables)})
}
