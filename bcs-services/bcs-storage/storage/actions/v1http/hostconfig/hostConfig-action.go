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

package hostconfig

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"

	"github.com/emicklei/go-restful"
)

const (
	clusterIDTag  = "clusterId"
	dataTag       = "data"
	ipTag         = "ip"
	tableName     = "host"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
	timeLayout    = "2006-01-02 15:04:05"
)

var needTimeFormatList = [...]string{updateTimeTag, createTimeTag}
var hostFeatTags = []string{ipTag}
var hostQueryFeatTags = []string{clusterIDTag}
var indexKeys = []string{ipTag}

// Use Mongodb for storage.
const dbConfig = "mongdb/host"

// GetHost get host
func GetHost(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetHost"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := getHost(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		err := fmt.Errorf("resource does not exist")
		utils.SetSpanLogTagError(span, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// PutHost put host
func PutHost(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutHost"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putHost(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteHost delete host
func DeleteHost(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteHost"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := removeHost(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// ListHost list host
func ListHost(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListHost"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := queryHost(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// PostClusterRelation post cluster relation
func PostClusterRelation(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PostClusterRelation"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := doRelation(req, false); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// PutClusterRelation put cluster relation
func PutClusterRelation(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutClusterRelation"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := doRelation(req, true); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	hostPath := urlPath("/host/{ip}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: hostPath, Params: nil, Handler: lib.MarkProcess(GetHost)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: hostPath, Params: nil, Handler: lib.MarkProcess(PutHost)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: hostPath, Params: nil, Handler: lib.MarkProcess(DeleteHost)})

	listMetricPath := urlPath("/host/clusters/{clusterId}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(ListHost)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(PutClusterRelation)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(PostClusterRelation)})
}
