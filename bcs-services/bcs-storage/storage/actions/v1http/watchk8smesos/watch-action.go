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

package watchk8smesos

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
)

const (
	// K8sEnv  k8s环境
	K8sEnv       = "k8s"
	urlK8SPrefix = "/" + K8sEnv
	// MesosEnv mesos环境
	MesosEnv       = "mesos"
	urlMesosPrefix = "/" + MesosEnv

	clusterIDTag  = "clusterId"
	tableTag      = "resourceType"
	namespaceTag  = "namespace"
	nameTag       = "resourceName"
	updateTimeTag = "updateTime"
)

// Use Zookeeper for storage.
const dbConfig = "zk/watch"

// K8SGetWatchResource get watch resource
func K8SGetWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "K8SGetWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := get(req, K8sEnv)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// K8SPutWatchResource put watch resource
func K8SPutWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "K8SPutWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := put(req, K8sEnv); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageRestRequestDataIsNotJsonStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageRestRequestDataIsNotJson,
			Message: common.BcsErrStorageRestRequestDataIsNotJsonStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// K8SDeleteWatchResource delete watch resource
func K8SDeleteWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "K8SDeleteWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := remove(req, K8sEnv); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// K8SListWatchResource list watch resource
func K8SListWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "K8SListWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := list(req, K8sEnv)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
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

// MesosGetWatchResource get watch resource
func MesosGetWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "MesosGetWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := get(req, MesosEnv)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

// MesosPutWatchResource put watch resource
func MesosPutWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "MesosPutWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := put(req, MesosEnv); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageRestRequestDataIsNotJsonStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageRestRequestDataIsNotJson,
			Message: common.BcsErrStorageRestRequestDataIsNotJsonStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// MesosDeleteWatchResource delete watch resource
func MesosDeleteWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "MesosDeleteWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := remove(req, MesosEnv); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// MesosListWatchResource list watch resource
func MesosListWatchResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "MesosListWatchResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := list(req, MesosEnv)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
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
		Params: nil, Handler: lib.MarkProcess(K8SGetWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: urlK8S, Params: nil, Handler: lib.MarkProcess(K8SPutWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: urlK8S, Params: nil, Handler: lib.MarkProcess(K8SDeleteWatchResource)})

	listK8SURL := urlK8SPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listK8SURL, Params: nil, Handler: lib.MarkProcess(K8SListWatchResource)})

	urlMesos := urlMesosPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlMesos,
		Params: nil, Handler: lib.MarkProcess(MesosGetWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: urlMesos, Params: nil, Handler: lib.MarkProcess(MesosPutWatchResource)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: urlMesos, Params: nil, Handler: lib.MarkProcess(MesosDeleteWatchResource)})

	listMesosURL := urlMesosPath("/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: listMesosURL, Params: nil, Handler: lib.MarkProcess(MesosListWatchResource)})
}
