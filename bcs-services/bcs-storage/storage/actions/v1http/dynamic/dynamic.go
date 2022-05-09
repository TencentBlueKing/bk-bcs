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
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/clean"

	"github.com/emicklei/go-restful"
)

// GetNamespaceResources get namespaced resources
func GetNamespaceResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetNamespaceResources"
	)

	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := getNamespaceResources(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		err := fmt.Errorf("resource does not exist")
		utils.SetSpanLogTagError(span, err)
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
	const (
		handler = "PutNamespaceResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putNamespaceResources(req); err != nil {
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

// DeleteNamespaceResources delete namespaced resources
func DeleteNamespaceResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteNamespaceResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteNamespaceResources(req); err != nil {
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

// GetClusterResources get cluster resources
func GetClusterResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetClusterResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := getClusterResources(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}

	if len(r) == 0 {
		err := fmt.Errorf("resource does not exist")
		utils.SetSpanLogTagError(span, err)
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
	const (
		handler = "PutClusterResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putClusterResources(req); err != nil {
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

// DeleteClusterResources delete cluster resources
func DeleteClusterResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteClusterResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteClusterResources(req); err != nil {
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

// ListNamespaceResources list namespaced resources
func ListNamespaceResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListNamespaceResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := listNamespaceResources(req)
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

// ListClusterResources list cluster resources
func ListClusterResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListClusterResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, err := listClusterResources(req)
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

// DeleteBatchNamespaceResource delete multiple namespaced resources
func DeleteBatchNamespaceResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteBatchNamespaceResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteBatchNamespaceResource(req); err != nil {
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

// DeleteBatchClusterResource delete multiple cluster resources
func DeleteBatchClusterResource(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteBatchClusterResource"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteClusterNamespaceResource(req); err != nil {
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

// GetCustomResources get custom resources
func GetCustomResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "GetCustomResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, extra, err := getCustomResources(req)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageGetResourceFail,
			Message: common.BcsErrStorageGetResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r, Extra: extra})
}

// PutCustomResources put custom resources
func PutCustomResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutCustomResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putCustomResources(req); err != nil {
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

// DeleteCustomResources delete custom resources
func DeleteCustomResources(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteCustomResources"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteCustomResources(req); err != nil {
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

// CreateCustomResourceIndex create custom resource's index
func CreateCustomResourcesIndex(req *restful.Request, resp *restful.Response) {
	const (
		handler = "CreateCustomResourcesIndex"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := createCustomResourcesIndex(req); err != nil {
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

// DeleteCustomResourceIndex delete custom resource's index
func DeleteCustomResourcesIndex(req *restful.Request, resp *restful.Response) {
	const (
		handler = "CreateCustomResourcesIndex"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteCustomResourcesIndex(req); err != nil {
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

func CleanDynamic() {
	dynamicDBClient := apiserver.GetAPIResource().GetDBClient(dbConfig)
	tables, err := dynamicDBClient.ListTableNames(context.TODO())
	if err != nil {
		blog.Errorf("list table name failed, err: %v", err)
		return
	}
	for _, table := range tables {
		cleaner := clean.NewDBCleaner(apiserver.GetAPIResource().GetDBClient(dbConfig), table, time.Hour)
		go cleaner.Run(context.TODO())
	}
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

	// Custom resources OPs
	customResourcePath := "/dynamic/customresources/{resourceType}"
	actions.RegisterV1Action(actions.Action{
		Verb:    "GET",
		Path:    customResourcePath,
		Params:  nil,
		Handler: lib.MarkProcess(GetCustomResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    customResourcePath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteCustomResources)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    customResourcePath,
		Params:  nil,
		Handler: lib.MarkProcess(PutCustomResources)})

	// Custom resource index
	customResourceIndexPath := "/dynamic/customresources/{resourceType}/index/{indexName}"
	actions.RegisterV1Action(actions.Action{
		Verb:    "PUT",
		Path:    customResourceIndexPath,
		Params:  nil,
		Handler: lib.MarkProcess(CreateCustomResourcesIndex)})
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    customResourceIndexPath,
		Params:  nil,
		Handler: lib.MarkProcess(DeleteCustomResourcesIndex)})

	actions.RegisterDaemonFunc(CleanDynamic)
}
