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

package clusterConfig

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	urlPrefix       = "/cluster_config"
	clusterIdTag    = "clusterId"
	serviceTag      = "service"
	clusterIdNotTag = "clusterIdNot"
	dataTag         = "data"
	versionTag      = "version"

	tableSvc      = "services"
	tableCls      = "clusters"
	tableVer      = "stableVersion"
	tableTpl      = "clusterTemplate"
	createTimeTag = "createTime"
	updateTimeTag = "updateTime"
)

// Use Mongodb for storage.
const dbConfig = "clusterConfig"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)
var indexKeys = []string{clusterIdTag}

func GetClusterConfig(req *restful.Request, resp *restful.Response) {
	request := newReqConfig(req)
	defer request.exit()
	r, err := request.generateData(request.getCls)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func PutClusterConfig(req *restful.Request, resp *restful.Response) {
	request := newReqConfig(req)
	defer request.exit()
	if err := request.putClsConfig(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	request.reset()
	r, err := request.generateData(request.getCls)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func GetServiceConfig(req *restful.Request, resp *restful.Response) {
	request := newReqConfig(req)
	defer request.exit()
	r, err := request.generateData(request.getMultiCls)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func GetStableVersion(req *restful.Request, resp *restful.Response) {
	request := newReqConfig(req)
	defer request.exit()
	version, err := request.getStableVersion()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: version})
}

func PutStableVersion(req *restful.Request, resp *restful.Response) {
	request := newReqConfig(req)
	defer request.exit()
	if err := request.putStableVersion(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	clusterUrl := urlPath("/clusters/{clusterId}/")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: clusterUrl, Params: nil, Handler: lib.MarkProcess(GetClusterConfig)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: clusterUrl, Params: nil, Handler: lib.MarkProcess(PutClusterConfig)})

	serviceUrl := urlPath("/services/{service}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: serviceUrl, Params: nil, Handler: lib.MarkProcess(GetServiceConfig)})

	versionUrl := urlPath("/versions/{service}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: versionUrl, Params: nil, Handler: lib.MarkProcess(GetStableVersion)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: versionUrl, Params: nil, Handler: lib.MarkProcess(PutStableVersion)})
}
