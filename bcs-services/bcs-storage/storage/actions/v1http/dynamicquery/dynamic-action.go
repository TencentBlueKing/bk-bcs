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

package dynamicquery

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
	urlPrefix        = "/query"
	clusterIDTag     = "clusterId"
	extraTag         = "extra"
	fieldTag         = "field"
	offsetTag        = "offset"
	limitTag         = "limit"
	namespaceTag     = "namespace"
	usedTag          = "used"
	timeLayout       = "2006-01-02 15:04:05"
	timestampsLayout = "timestamps"
	// NestedTimeLayout time format
	NestedTimeLayout = "2006-01-02T15:04:05-0700"
	updateTimeTag    = "updateTime"
	createTimeTag    = "createTime"
)

// Use Mongodb for storage.
const dbConfig = "mongodb/dynamic"

// doQuery queries the dynamic database with the given filter and name.
func doQuery(req *restful.Request, resp *restful.Response, filter qFilter, name string) error {
	// Create a new dynamic request.
	request := newReqDynamic(req, filter, name)
	// Query the dynamic database.
	r, err := request.QueryDynamic()
	if err != nil {
		// If failed to query, log the error and return an error response.
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail,
			Message: common.BcsErrStorageListResourceFailStr})
		return err
	}
	// Return the query result.
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})

	return nil
}

// GetNameSpace get namespace 获取命名空间
func GetNameSpace(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetNameSpace"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := getMesosNamespaceResource(req, resp); err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetTaskGroup get taskgroup 获取taskgroup
func GetTaskGroup(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetTaskGroup"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &TaskGroupFilter{}, "taskgroup")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetApplication get application 获取Application
func GetApplication(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetApplication"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ApplicationFilter{Kind: ",application"}, "application")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetProcess get process 获取process
func GetProcess(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetProcess"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ProcessFilter{Kind: "process"}, "application")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetDeployment get deployment 获取deployment工作负载
func GetDeployment(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetDeployment"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &DeploymentFilter{}, "deployment")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetService get service 获取Service
func GetService(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetService"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ServiceFilter{}, "service")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetConfigMap get configmap 获取configmap
func GetConfigMap(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetConfigMap"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ConfigMapFilter{}, "configmap")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetSecret get secret 获取secret
func GetSecret(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetSecret"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &SecretFilter{}, "secret")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetEndpoints get endpoints 获取endpoints
func GetEndpoints(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetEndpoints"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &EndpointsFilter{}, "endpoint")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetExportService get export service 获取ExportService
func GetExportService(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetExportService"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ExportServiceFilter{}, "exportservice")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetPod get pod 获取Pod
func GetPod(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetPod"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &PodFilter{}, "Pod")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetReplicaSet get replicaset 获取replicaset
func GetReplicaSet(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetReplicaSet"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ReplicaSetFilter{}, "ReplicaSet")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetDeploymentK8s get deployment k8s 获取k8s Deployment
func GetDeploymentK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetDeploymentK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &DeploymentK8sFilter{}, "Deployment")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetServiceK8s get service k8s 获取k8s service
func GetServiceK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetServiceK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ServiceK8sFilter{}, "Service")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetConfigMapK8s get configmap k8s 获取k8s service
func GetConfigMapK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetConfigMapK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &ConfigMapK8sFilter{}, "ConfigMap")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetSecretK8s get secret k8s 获取k8s service
func GetSecretK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetSecretK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &SecretK8sFilter{}, "Secret")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetEndpointsK8s get endpoints k8s 获取k8s endpoints
func GetEndpointsK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetEndpointsK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &EndpointsK8sFilter{}, "Endpoints")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetIngress get ingress 获取ingress
func GetIngress(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetIngress"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &IngressFilter{}, "Ingress")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetNameSpaceK8s get namespaces k8s 获取k8s 命名空间
func GetNameSpaceK8s(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetNameSpaceK8s"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if req.QueryParameter(usedTag) == "1" {
		err := getK8sNamespaceResource(req, resp)
		if err != nil {
			utils.SetSpanLogTagError(span, err)
		}
		return
	}
	err := doQuery(req, resp, &NameSpaceFilter{}, "Namespace")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetNode get node 获取node
func GetNode(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetNode"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &NodeFilter{}, "Node")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetDaemonSet get daemonset 获取daemonset
func GetDaemonSet(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetDaemonSet"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &DaemonSetFilter{}, "DaemonSet")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetJob get job 获取job
func GetJob(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetJob"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &JobFilter{}, "Job")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetStatefulSet get statefulset 获取statefulset
func GetStatefulSet(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetStatefulSet"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &StatefulSetFilter{}, "StatefulSet")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetIPPoolStatic query netservice ip pool static resource data.
func GetIPPoolStatic(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetIPPoolStatic"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &IPPoolStaticFilter{}, "IPPoolStatic")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// GetIPPoolStaticDetail query netservice ip pool static resource detail data.
func GetIPPoolStaticDetail(req *restful.Request, resp *restful.Response) {
	// Define a constant named handler
	const (
		handler = "GetIPPoolStaticDetail"
	)
	// Create a span to trace the execution of the function
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	err := doQuery(req, resp, &IPPoolStaticDetailFilter{}, "IPPoolStaticDetail")
	if err != nil {
		utils.SetSpanLogTagError(span, err)
	}
}

// NOCC: golint/funlen(设计如此:)
// nolint
func init() {
	// GET
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/taskgroup"),
		Params: nil, Handler: lib.MarkProcess(GetTaskGroup)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/application"),
		Params: nil, Handler: lib.MarkProcess(GetApplication)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/process"),
		Params: nil, Handler: lib.MarkProcess(GetProcess)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/deployment"),
		Params: nil, Handler: lib.MarkProcess(GetDeployment)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/service"),
		Params: nil, Handler: lib.MarkProcess(GetService)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/configmap"),
		Params: nil, Handler: lib.MarkProcess(GetConfigMap)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/secret"),
		Params: nil, Handler: lib.MarkProcess(GetSecret)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/endpoints"),
		Params: nil, Handler: lib.MarkProcess(GetEndpoints)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/exportservice"),
		Params: nil, Handler: lib.MarkProcess(GetExportService)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/namespace"),
		Params: nil, Handler: lib.MarkProcess(GetNameSpace)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/ippoolstatic"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStatic)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/ippoolstaticdetail"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStaticDetail)})

	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/pod"),
		Params: nil, Handler: lib.MarkProcess(GetPod)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/replicaset"),
		Params: nil, Handler: lib.MarkProcess(GetReplicaSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/deployment"),
		Params: nil, Handler: lib.MarkProcess(GetDeploymentK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/service"),
		Params: nil, Handler: lib.MarkProcess(GetServiceK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/configmap"),
		Params: nil, Handler: lib.MarkProcess(GetConfigMapK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/secret"),
		Params: nil, Handler: lib.MarkProcess(GetSecretK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/endpoints"),
		Params: nil, Handler: lib.MarkProcess(GetEndpointsK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ingress"),
		Params: nil, Handler: lib.MarkProcess(GetIngress)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/namespace"),
		Params: nil, Handler: lib.MarkProcess(GetNameSpaceK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/node"),
		Params: nil, Handler: lib.MarkProcess(GetNode)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/daemonset"),
		Params: nil, Handler: lib.MarkProcess(GetDaemonSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/job"),
		Params: nil, Handler: lib.MarkProcess(GetJob)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/statefulset"),
		Params: nil, Handler: lib.MarkProcess(GetStatefulSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ippoolstatic"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStatic)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ippoolstaticdetail"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStaticDetail)})

	// POST
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/taskgroup"),
		Params: nil, Handler: lib.MarkProcess(GetTaskGroup)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/application"),
		Params: nil, Handler: lib.MarkProcess(GetApplication)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/process"),
		Params: nil, Handler: lib.MarkProcess(GetProcess)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/deployment"),
		Params: nil, Handler: lib.MarkProcess(GetDeployment)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/service"),
		Params: nil, Handler: lib.MarkProcess(GetService)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/configmap"),
		Params: nil, Handler: lib.MarkProcess(GetConfigMap)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/secret"),
		Params: nil, Handler: lib.MarkProcess(GetSecret)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/endpoints"),
		Params: nil, Handler: lib.MarkProcess(GetEndpoints)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/exportservice"),
		Params: nil, Handler: lib.MarkProcess(GetExportService)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/namespace"),
		Params: nil, Handler: lib.MarkProcess(GetNameSpace)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/ippoolstatic"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStatic)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/mesos/dynamic/clusters/{clusterId}/ippoolstaticdetail"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStaticDetail)})

	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/pod"),
		Params: nil, Handler: lib.MarkProcess(GetPod)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/replicaset"),
		Params: nil, Handler: lib.MarkProcess(GetReplicaSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/deployment"),
		Params: nil, Handler: lib.MarkProcess(GetDeploymentK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/service"),
		Params: nil, Handler: lib.MarkProcess(GetServiceK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/configmap"),
		Params: nil, Handler: lib.MarkProcess(GetConfigMapK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/secret"),
		Params: nil, Handler: lib.MarkProcess(GetSecretK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/endpoints"),
		Params: nil, Handler: lib.MarkProcess(GetEndpointsK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ingress"),
		Params: nil, Handler: lib.MarkProcess(GetIngress)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/namespace"),
		Params: nil, Handler: lib.MarkProcess(GetNameSpaceK8s)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/node"),
		Params: nil, Handler: lib.MarkProcess(GetNode)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/daemonset"),
		Params: nil, Handler: lib.MarkProcess(GetDaemonSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/job"),
		Params: nil, Handler: lib.MarkProcess(GetJob)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/statefulset"),
		Params: nil, Handler: lib.MarkProcess(GetStatefulSet)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ippoolstatic"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStatic)})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: urlPath("/k8s/dynamic/clusters/{clusterId}/ippoolstaticdetail"),
		Params: nil, Handler: lib.MarkProcess(GetIPPoolStaticDetail)})
}
