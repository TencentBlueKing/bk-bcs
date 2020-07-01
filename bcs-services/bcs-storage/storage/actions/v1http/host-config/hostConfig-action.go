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

package hostConfig

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	clusterIdTag  = "clusterId"
	dataTag       = "data"
	ipTag         = "ip"
	tableName     = "host"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
	timeLayout    = "2006-01-02 15:04:05"
)

var needTimeFormatList = [...]string{updateTimeTag, createTimeTag}
var hostFeatTags = []string{ipTag}
var hostQueryFeatTags = []string{clusterIdTag}
var indexKeys = []string{ipTag}

// Use Mongodb for storage.
const dbConfig = "host"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func GetHost(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	r, err := request.getHost()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func PutHost(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	if err := request.putHost(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteHost(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	if err := request.removeHost(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func ListHost(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	r, err := request.queryHost()

	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func PostClusterRelation(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	if err := request.doRelation(false); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func PutClusterRelation(req *restful.Request, resp *restful.Response) {
	request := newReqHost(req)
	defer request.exit()
	if err := request.doRelation(true); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	hostPath := urlPath("/host/{ip}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: hostPath, Params: nil, Handler: lib.MarkProcess(GetHost)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: hostPath, Params: nil, Handler: lib.MarkProcess(PutHost)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: hostPath, Params: nil, Handler: lib.MarkProcess(DeleteHost)})

	listMetricPath := urlPath("/host/clusters/{clusterId}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(ListHost)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(PutClusterRelation)})
	actions.RegisterV1Action(actions.Action{Verb: "POST", Path: listMetricPath, Params: nil, Handler: lib.MarkProcess(PostClusterRelation)})
}
