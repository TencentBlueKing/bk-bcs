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

package dynamic

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	urlPrefix       = "/mesos"
	clusterIdTag    = "clusterId"
	namespaceTag    = "namespace"
	resourceTypeTag = "resourceType"
	resourceNameTag = "resourceName"

	tableTag      = resourceTypeTag
	dataTag       = "data"
	extraTag      = "extra"
	fieldTag      = "field"
	offsetTag     = "offset"
	limitTag      = "limit"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
	timeLayout    = "2006-01-02 15:04:05"

	applicationTypeName = "application"
	processTypeName     = "process"
	kindTag             = "data.kind"
)

var needTimeFormatList = [...]string{updateTimeTag, createTimeTag}
var nsFeatTags = []string{clusterIdTag, namespaceTag, resourceTypeTag, resourceNameTag}
var csFeatTags = []string{clusterIdTag, resourceTypeTag, resourceNameTag}
var nsListFeatTags = []string{clusterIdTag, namespaceTag, resourceTypeTag}
var csListFeatTags = []string{clusterIdTag, resourceTypeTag}
var indexKeys = []string{resourceNameTag, namespaceTag}

// Use Mongodb for storage.
const dbConfig = "dynamic"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func GetNamespaceResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	r, err := request.nsGet()
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

func PutNamespaceResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.nsPut(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteNamespaceResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.nsRemove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func GetClusterResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	r, err := request.csGet()
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

func PutClusterResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.csPut(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteClusterResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.csRemove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func ListNamespaceResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	r, err := request.nsList()

	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func ListClusterResources(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	r, err := request.csList()

	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func DeleteBatchNamespaceResource(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.nsBatchRemove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func DeleteBatchClusterResource(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req)
	defer request.exit()
	if err := request.csBatchRemove(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStorageDeleteResourceFail, Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	// Namespace resources.
	namespaceResourcesPath := urlPath("/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: namespaceResourcesPath, Params: nil, Handler: lib.MarkProcess(GetNamespaceResources)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: namespaceResourcesPath, Params: nil, Handler: lib.MarkProcess(PutNamespaceResources)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: namespaceResourcesPath, Params: nil, Handler: lib.MarkProcess(DeleteNamespaceResources)})

	listNamespaceResourcesPath := urlPath("/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listNamespaceResourcesPath, Params: nil, Handler: lib.MarkProcess(ListNamespaceResources)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: listNamespaceResourcesPath, Params: nil, Handler: lib.MarkProcess(DeleteBatchNamespaceResource)})

	// Cluster resources.
	clusterResourcesPath := urlPath("/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: clusterResourcesPath, Params: nil, Handler: lib.MarkProcess(GetClusterResources)})
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: clusterResourcesPath, Params: nil, Handler: lib.MarkProcess(PutClusterResources)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: clusterResourcesPath, Params: nil, Handler: lib.MarkProcess(DeleteClusterResources)})

	listClusterResourcesPath := urlPath("/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: listClusterResourcesPath, Params: nil, Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: listClusterResourcesPath, Params: nil, Handler: lib.MarkProcess(DeleteBatchClusterResource)})

	// All Ops.
	allResourcesPath := urlPath("/dynamic/all_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: allResourcesPath, Params: nil, Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{Verb: "DELETE", Path: allResourcesPath, Params: nil, Handler: lib.MarkProcess(DeleteBatchClusterResource)})
}
