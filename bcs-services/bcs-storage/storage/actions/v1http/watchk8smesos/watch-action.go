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

package watchk8smesos

import (
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

const (
	k8sEnv         = "k8s"
	urlK8SPrefix   = "/" + k8sEnv
	mesosEnv       = "mesos"
	urlMesosPrefix = "/" + mesosEnv

	clusterIDTag  = "clusterId"
	tableTag      = "resourceType"
	namespaceTag  = "namespace"
	nameTag       = "resourceName"
	updateTimeTag = "updateTime"
)

// Use Zookeeper for storage.
const dbConfig = "watch"

// GetWatchResource get watch resource
func GetWatchResource(req *restful.Request, resp *restful.Response) {
	r, err := get(req)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// PutWatchResource put watch resource
func PutWatchResource(req *restful.Request, resp *restful.Response) {
	if err := put(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageRestRequestDataIsNotJsonStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageRestRequestDataIsNotJson,
			Message: common.BcsErrStorageRestRequestDataIsNotJsonStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteWatchResource delete watch resource
func DeleteWatchResource(req *restful.Request, resp *restful.Response) {
	if err := remove(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// ListWatchResource list watch resource
func ListWatchResource(req *restful.Request, resp *restful.Response) {
	r, err := list(req)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			Data:    []string{},
			ErrCode: common.BcsErrStorageListResourceFail,
			Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func init() {
	urlK8S := urlK8SPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlK8S,
		Params: nil, Handler: lib.MarkProcess(GetWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: urlK8S, Params: nil, Handler: lib.MarkProcess(PutWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: urlK8S, Params: nil, Handler: lib.MarkProcess(DeleteWatchResource)})

	listK8SURL := urlMesosPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listK8SURL, Params: nil, Handler: lib.MarkProcess(ListWatchResource)})

	urlMesos := urlMesosPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlMesos,
		Params: nil, Handler: lib.MarkProcess(GetWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: urlMesos, Params: nil, Handler: lib.MarkProcess(PutWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: urlMesos, Params: nil, Handler: lib.MarkProcess(DeleteWatchResource)})

	listMesosURL := urlMesosPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listMesosURL, Params: nil, Handler: lib.MarkProcess(ListWatchResource)})
}
