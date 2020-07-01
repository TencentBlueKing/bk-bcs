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
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	clusterIdTag  = "clusterId"
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
var metricFeatTags = []string{clusterIdTag, namespaceTag, typeTag, nameTag}
var queryFeatTags = []string{clusterIdTag}
var queryExtraTags = []string{namespaceTag, typeTag, nameTag}
var indexKeys = []string{clusterIdTag, namespaceTag, typeTag, nameTag}

// Use Mongodb for storage.
const dbConfig = "metric"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func GetMetric(req *restful.Request, resp *restful.Response) {
	request := newReqMetric(req)
	defer request.exit()
	r, err := request.getMetric()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		if err == storageErr.ResourceDoesNotExist {
			lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
			return
		}
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	if len(r) == 0 {
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r[0]})
}

func PutMetric(req *restful.Request, resp *restful.Response) {
	request := newReqMetric(req)
	defer request.exit()
	if err := request.put(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteMetric(req *restful.Request, resp *restful.Response) {
	request := newReqMetric(req)
	defer request.exit()
	if err := request.remove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		if err == storageErr.ResourceDoesNotExist {
			lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
			return
		}
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func QueryMetric(req *restful.Request, resp *restful.Response) {
	request := newReqMetric(req)
	defer request.exit()
	r, err := request.queryMetric()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func ListMetricTables(req *restful.Request, resp *restful.Response) {
	request := newReqMetric(req)
	defer request.exit()
	r, err := request.tables()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDecodeListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageDecodeListResourceFail, Message: common.BcsErrStorageDecodeListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func init() {
	metricPath := urlPath("/metric/clusters/{clusterId}/namespaces/{namespace}/{type}/{name}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: metricPath, Params: nil, Handler: lib.MarkProcess(GetMetric)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: metricPath, Params: nil, Handler: lib.MarkProcess(PutMetric)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: metricPath, Params: nil, Handler: lib.MarkProcess(DeleteMetric)})

	listMetricPath := urlPath("/metric/clusters/{clusterId}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(QueryMetric)})

	listMetricTablePath := urlPath("/metric/clusters")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listMetricTablePath, Params: nil, Handler: lib.MarkProcess(ListMetricTables)})
}
