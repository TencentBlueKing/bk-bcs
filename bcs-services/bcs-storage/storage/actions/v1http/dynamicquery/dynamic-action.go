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

package dynamicquery

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"

	"github.com/emicklei/go-restful"
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
	nestedTimeLayout = "2006-01-02T15:04:05-0700"
	updateTimeTag    = "updateTime"
	createTimeTag    = "createTime"
)

var needTimeFormatList = [...]string{updateTimeTag, createTimeTag}

// Use Mongodb for storage.
const dbConfig = "dynamic"

func doQuery(req *restful.Request, resp *restful.Response, filter qFilter, name string) {
	request := newReqDynamic(req, filter, name)
	r, err := request.queryDynamic()
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r})
}

func grepNamespace(req *restful.Request, filter qFilter, name string, origin []string) ([]string, error) {
	request := newReqDynamic(req, filter, name)
	r, err := request.queryDynamic()
	if err != nil {
		return nil, err
	}
	return fetchNamespace(r, origin), nil
}

// GetNameSpace get namespace
func GetNameSpace(req *restful.Request, resp *restful.Response) {
	// init Form
	req.Request.FormValue("")
	req.Request.Form[fieldTag] = []string{namespaceTag}
	var err error
	result := make([]string, 0)

	// grep application
	if result, err = grepNamespace(req, &ApplicationFilter{Kind: ",application"}, "application", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep process
	if result, err = grepNamespace(req, &ProcessFilter{Kind: "process"}, "application", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep deployment
	if result, err = grepNamespace(req, &DeploymentFilter{}, "deployment", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep service
	if result, err = grepNamespace(req, &ServiceFilter{}, "service", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep configMap
	if result, err = grepNamespace(req, &ConfigMapFilter{}, "configmap", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep secret
	if result, err = grepNamespace(req, &SecretFilter{}, "secret", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: result})
}

// GetTaskGroup get taskgroup
func GetTaskGroup(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &TaskGroupFilter{}, "taskgroup")
}

// GetApplication get application
func GetApplication(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ApplicationFilter{Kind: ",application"}, "application")
}

// GetProcess get process
func GetProcess(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ProcessFilter{Kind: "process"}, "application")
}

// GetDeployment get deployment
func GetDeployment(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DeploymentFilter{}, "deployment")
}

// GetService get service
func GetService(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ServiceFilter{}, "service")
}

// GetConfigMap get configmap
func GetConfigMap(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ConfigMapFilter{}, "configmap")
}

// GetSecret get secret
func GetSecret(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &SecretFilter{}, "secret")
}

// GetEndpoints get endpoints
func GetEndpoints(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &EndpointsFilter{}, "endpoint")
}

// GetExportService get export service
func GetExportService(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ExportServiceFilter{}, "exportservice")
}

// GetNameSpaceK8sUsed get namespace k8s used
func GetNameSpaceK8sUsed(req *restful.Request, resp *restful.Response) {
	// init Form
	req.Request.FormValue("")
	req.Request.Form[fieldTag] = []string{namespaceTag}
	var err error
	result := make([]string, 0)

	// grep replicaSet
	if result, err = grepNamespace(req, &ReplicaSetFilter{}, "ReplicaSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep deployment
	if result, err = grepNamespace(req, &DeploymentK8sFilter{}, "Deployment", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep service
	if result, err = grepNamespace(req, &ServiceK8sFilter{}, "Service", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep configMap
	if result, err = grepNamespace(req, &ConfigMapK8sFilter{}, "ConfigMap", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep secret
	if result, err = grepNamespace(req, &SecretK8sFilter{}, "Secret", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep ingress
	if result, err = grepNamespace(req, &IngressFilter{}, "Ingress", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep daemonSet
	if result, err = grepNamespace(req, &DaemonSetFilter{}, "DaemonSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep job
	if result, err = grepNamespace(req, &JobFilter{}, "Job", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep statefulSet
	if result, err = grepNamespace(req, &StatefulSetFilter{}, "StatefulSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: result})
}

// GetPod get pod
func GetPod(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &PodFilter{}, "Pod")
}

// GetReplicaSet get replica
func GetReplicaSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ReplicaSetFilter{}, "ReplicaSet")
}

// GetDeploymentK8s get deployment k8s
func GetDeploymentK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DeploymentK8sFilter{}, "Deployment")
}

// GetServiceK8s get service k8s
func GetServiceK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ServiceK8sFilter{}, "Service")
}

// GetConfigMapK8s get configmap k8s
func GetConfigMapK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ConfigMapK8sFilter{}, "ConfigMap")
}

// GetSecretK8s get secret k8s
func GetSecretK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &SecretK8sFilter{}, "Secret")
}

// GetEndpointsK8s get endpoints k8s
func GetEndpointsK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &EndpointsK8sFilter{}, "EndPoints")
}

// GetIngress get ingress
func GetIngress(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &IngressFilter{}, "Ingress")
}

// GetNameSpaceK8s get namespaces k8s
func GetNameSpaceK8s(req *restful.Request, resp *restful.Response) {
	if req.QueryParameter(usedTag) == "1" {
		GetNameSpaceK8sUsed(req, resp)
		return
	}
	doQuery(req, resp, &NameSpaceFilter{}, "Namespace")
}

// GetNode get node
func GetNode(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &NodeFilter{}, "Node")
}

// GetDaemonSet get daemonset
func GetDaemonSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DaemonSetFilter{}, "DaemonSet")
}

// GetJob get job
func GetJob(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &JobFilter{}, "Job")
}

// GetStatefulSet get statefulset
func GetStatefulSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &StatefulSetFilter{}, "StatefulSet")
}

// GetIPPoolStatic query netservice ip pool static resource data.
func GetIPPoolStatic(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &IPPoolStaticFilter{}, "IPPoolStatic")
}

// GetIPPoolStaticDetail query netservice ip pool static resource detail data.
func GetIPPoolStaticDetail(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &IPPoolStaticDetailFilter{}, "IPPoolStaticDetail")
}

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
		Params: nil, Handler: lib.MarkProcess(GetNameSpace)})
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
