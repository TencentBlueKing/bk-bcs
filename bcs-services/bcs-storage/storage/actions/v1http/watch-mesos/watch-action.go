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

package watch

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	env       = "mesos"
	urlPrefix = "/" + env

	clusterIdTag  = "clusterId"
	tableTag      = "resourceType"
	namespaceTag  = "namespace"
	nameTag       = "resourceName"
	updateTimeTag = "updateTime"
)

// Use Zookeeper for storage.
const dbConfig = "watch"

var getNewTank operator.GetNewTank = lib.GetZookeeperTank(dbConfig)

func GetWatchResource(req *restful.Request, resp *restful.Response) {
	request := newReqWatch(req)
	defer request.exit()
	r, err := request.get()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageGetResourceFail, Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	if len(r) == 0 {
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageResourceNotExist, Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r[0]})
}

func PutWatchResource(req *restful.Request, resp *restful.Response) {
	request := newReqWatch(req)
	defer request.exit()
	if err := request.put(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageRestRequestDataIsNotJsonStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageRestRequestDataIsNotJson, Message: common.BcsErrStorageRestRequestDataIsNotJsonStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteWatchResource(req *restful.Request, resp *restful.Response) {
	request := newReqWatch(req)
	defer request.exit()
	if err := request.remove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func ListWatchResource(req *restful.Request, resp *restful.Response) {
	request := newReqWatch(req)
	defer request.exit()
	r, err := request.list()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func init() {
	url := urlPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: url, Params: nil, Handler: lib.MarkProcess(GetWatchResource)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: url, Params: nil, Handler: lib.MarkProcess(PutWatchResource)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: url, Params: nil, Handler: lib.MarkProcess(DeleteWatchResource)})

	listUrl := urlPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listUrl, Params: nil, Handler: lib.MarkProcess(ListWatchResource)})
}
