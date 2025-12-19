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

// Package clusterconfig xxx
package clusterconfig

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
const dbConfig = "mongodb/clusterConfig"

// GetClusterConfig get cluster config
func GetClusterConfig(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetClusterConfig"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := generateData(req, getCls)
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

// PutClusterConfig put cluster config
func PutClusterConfig(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutClusterConfig"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putClsConfig(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	r, err := generateData(req, getCls)
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

// GetServiceConfig get service config
func GetServiceConfig(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetServiceConfig"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := generateData(req, getMultiCls)
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

// GetStableVersion get stable version
func GetStableVersion(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetStableVersion"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	// ctx
	ctx := req.Request.Context()
	// option
	opt := &lib.StoreGetOption{
		Cond: getSvcCondition(req),
	}

	version, err := GetStableSvcVersion(ctx, opt)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: version})
}

// PutStableVersion put stable version
func PutStableVersion(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutStableVersion"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putStableVersion(req); err != nil {
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
	clusterURL := urlPath("/clusters/{clusterId}/")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    clusterURL,
		Params:  nil,
		Handler: lib.MarkProcess(GetClusterConfig)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    clusterURL,
		Params:  nil,
		Handler: lib.MarkProcess(PutClusterConfig)})

	serviceURL := urlPath("/services/{service}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    serviceURL,
		Params:  nil,
		Handler: lib.MarkProcess(GetServiceConfig)})

	versionURL := urlPath("/versions/{service}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    versionURL,
		Params:  nil,
		Handler: lib.MarkProcess(GetStableVersion)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    versionURL,
		Params:  nil,
		Handler: lib.MarkProcess(PutStableVersion)})
}
