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

	"github.com/emicklei/go-restful"
)

// GetNamespaceResources get namespaced resources
func GetNamespaceResources(req *restful.Request, resp *restful.Response) {
	r, err := getNamespaceResources(req)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageResourceNotExist,
			Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r[0]})
}

// PutNamespaceResources put namespaced resources
func PutNamespaceResources(req *restful.Request, resp *restful.Response) {
	if err := putNamespaceResources(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteNamespaceResources delete namespaced resources
func DeleteNamespaceResources(req *restful.Request, resp *restful.Response) {
	if err := deleteNamespaceResources(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// GetClusterResources get cluster resources
func GetClusterResources(req *restful.Request, resp *restful.Response) {
	r, err := getClusterResources(req)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageResourceNotExist,
			Message: common.BcsErrStorageResourceNotExistStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r[0]})
}

// PutClusterResources put cluster resources
func PutClusterResources(req *restful.Request, resp *restful.Response) {
	if err := putClusterResources(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteClusterResources delete cluster resources
func DeleteClusterResources(req *restful.Request, resp *restful.Response) {
	if err := deleteClusterResources(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// ListNamespaceResources list namespaced resources
func ListNamespaceResources(req *restful.Request, resp *restful.Response) {
	r, err := listNamespaceResources(req)
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

// ListClusterResources list cluster resources
func ListClusterResources(req *restful.Request, resp *restful.Response) {
	r, err := listClusterResources(req)
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

// DeleteBatchNamespaceResource delete multiple namespaced resources
func DeleteBatchNamespaceResource(req *restful.Request, resp *restful.Response) {
	if err := deleteBatchNamespaceResource(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteBatchClusterResource delete multiple cluster resources
func DeleteBatchClusterResource(req *restful.Request, resp *restful.Response) {
	if err := deleteClusterNamespaceResource(req); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	// for k8s
	// Namespace resources.
	k8sNamespaceResourcesPath := urlPathK8S(
		"/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    k8sNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(GetNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    k8sNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(PutNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    k8sNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteNamespaceResources)})

	k8sListNamespaceResourcesPath := urlPathK8S(
		"/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    k8sListNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    k8sListNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchNamespaceResource)})

	// Cluster resources.
	k8sClusterResourcesPath := urlPathK8S(
		"/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    k8sClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(GetClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    k8sClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(PutClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    k8sClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteClusterResources)})

	k8sListClusterResourcesPath := urlPathK8S(
		"/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    k8sListClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    k8sListClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchClusterResource)})

	// All Ops.
	k8sAllResourcesPath := urlPathK8S(
		"/dynamic/all_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    k8sAllResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    k8sAllResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchClusterResource)})

	// for mesos
	// Namespace resources.
	mesosNamespaceResourcesPath := urlPathMesos(
		"/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    mesosNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(GetNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    mesosNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(PutNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    mesosNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteNamespaceResources)})

	mesosListNamespaceResourcesPath := urlPathMesos(
		"/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    mesosListNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListNamespaceResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    mesosListNamespaceResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchNamespaceResource)})

	// Cluster resources.
	mesosClusterResourcesPath := urlPathMesos(
		"/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    mesosClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(GetClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    mesosClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(PutClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    mesosClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteClusterResources)})

	mesosListClusterResourcesPath := urlPathMesos(
		"/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    mesosListClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    mesosListClusterResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchClusterResource)})

	// All Ops.
	mesosAllResourcesPath := urlPathMesos(
		"/dynamic/all_resources/clusters/{clusterId}/{resourceType}")
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    mesosAllResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(ListClusterResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    mesosAllResourcesPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteBatchClusterResource)})
}
