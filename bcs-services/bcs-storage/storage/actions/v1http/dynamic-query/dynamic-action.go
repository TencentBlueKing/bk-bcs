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

package dynamicQuery

import (
	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-storage/storage/actions"
	"bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	urlPrefix        = "/query"
	clusterIdTag     = "clusterId"
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

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func doQuery(req *restful.Request, resp *restful.Response, filter qFilter, name string) {
	request := newReqDynamic(req, filter, name)
	defer request.exit()
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
	defer request.exit()
	r, err := request.queryDynamic()
	if err != nil {
		return nil, err
	}
	return fetchNamespace(r, origin), nil
}

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

func GetTaskGroup(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &TaskGroupFilter{}, "taskgroup")
}

func GetApplication(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ApplicationFilter{Kind: ",application"}, "application")
}

func GetProcess(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ProcessFilter{Kind: "process"}, "application")
}

func GetDeployment(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DeploymentFilter{}, "deployment")
}

func GetService(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ServiceFilter{}, "service")
}

func GetConfigMap(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ConfigMapFilter{}, "configmap")
}

func GetSecret(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &SecretFilter{}, "secret")
}

func GetEndpoints(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &EndpointsFilter{}, "endpoint")
}

func GetExportService(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ExportServiceFilter{}, "exportservice")
}

func GetNameSpaceK8sUsed(req *restful.Request, resp *restful.Response) {
	// init Form
	req.Request.FormValue("")
	req.Request.Form[fieldTag] = []string{namespaceTag}
	var err error
	result := make([]string, 0)

	// grep replicaSet
	if result, err = grepNamespace(req, &ReplicaSetFilter{}, "ReplicaSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep deployment
	if result, err = grepNamespace(req, &DeploymentK8sFilter{}, "Deployment", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep service
	if result, err = grepNamespace(req, &ServiceK8sFilter{}, "Service", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep configMap
	if result, err = grepNamespace(req, &ConfigMapK8sFilter{}, "ConfigMap", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep secret
	if result, err = grepNamespace(req, &SecretK8sFilter{}, "Secret", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep ingress
	if result, err = grepNamespace(req, &IngressFilter{}, "Ingress", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep daemonSet
	if result, err = grepNamespace(req, &DaemonSetFilter{}, "DaemonSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep job
	if result, err = grepNamespace(req, &JobFilter{}, "Job", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	// grep statefulSet
	if result, err = grepNamespace(req, &StatefulSetFilter{}, "StatefulSet", result); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr})
		return
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: result})
}

func GetPod(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &PodFilter{}, "Pod")
}

func GetReplicaSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ReplicaSetFilter{}, "ReplicaSet")
}

func GetDeploymentK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DeploymentK8sFilter{}, "Deployment")
}

func GetServiceK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ServiceK8sFilter{}, "Service")
}

func GetConfigMapK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &ConfigMapK8sFilter{}, "ConfigMap")
}

func GetSecretK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &SecretK8sFilter{}, "Secret")
}

func GetEndpointsK8s(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &EndpointsK8sFilter{}, "EndPoints")
}

func GetIngress(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &IngressFilter{}, "Ingress")
}

func GetNameSpaceK8s(req *restful.Request, resp *restful.Response) {
	if req.QueryParameter(usedTag) == "1" {
		GetNameSpaceK8sUsed(req, resp)
		return
	}
	doQuery(req, resp, &NameSpaceFilter{}, "Namespace")
}

func GetNode(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &NodeFilter{}, "Node")
}

func GetDaemonSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &DaemonSetFilter{}, "DaemonSet")
}

func GetJob(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &JobFilter{}, "Job")
}

func GetStatefulSet(req *restful.Request, resp *restful.Response) {
	doQuery(req, resp, &StatefulSetFilter{}, "StatefulSet")
}

func init() {
	// GET
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/taskgroup"), nil, lib.MarkProcess(GetTaskGroup)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/application"), nil, lib.MarkProcess(GetApplication)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/process"), nil, lib.MarkProcess(GetProcess)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/deployment"), nil, lib.MarkProcess(GetDeployment)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/service"), nil, lib.MarkProcess(GetService)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/configmap"), nil, lib.MarkProcess(GetConfigMap)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/secret"), nil, lib.MarkProcess(GetSecret)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/endpoints"), nil, lib.MarkProcess(GetEndpoints)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/exportservice"), nil, lib.MarkProcess(GetExportService)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/mesos/dynamic/clusters/{clusterId}/namespace"), nil, lib.MarkProcess(GetNameSpace)})

	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/pod"), nil, lib.MarkProcess(GetPod)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/replicaset"), nil, lib.MarkProcess(GetReplicaSet)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/deployment"), nil, lib.MarkProcess(GetDeploymentK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/service"), nil, lib.MarkProcess(GetServiceK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/configmap"), nil, lib.MarkProcess(GetConfigMapK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/secret"), nil, lib.MarkProcess(GetSecretK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/endpoints"), nil, lib.MarkProcess(GetEndpointsK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/ingress"), nil, lib.MarkProcess(GetIngress)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/namespace"), nil, lib.MarkProcess(GetNameSpaceK8s)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/node"), nil, lib.MarkProcess(GetNode)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/daemonset"), nil, lib.MarkProcess(GetDaemonSet)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/job"), nil, lib.MarkProcess(GetJob)})
	actions.RegisterV1Action(actions.Action{"GET", urlPath("/k8s/dynamic/clusters/{clusterId}/statefulset"), nil, lib.MarkProcess(GetStatefulSet)})

	// POST
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/taskgroup"), nil, lib.MarkProcess(GetTaskGroup)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/application"), nil, lib.MarkProcess(GetApplication)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/process"), nil, lib.MarkProcess(GetProcess)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/deployment"), nil, lib.MarkProcess(GetDeployment)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/service"), nil, lib.MarkProcess(GetService)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/configmap"), nil, lib.MarkProcess(GetConfigMap)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/secret"), nil, lib.MarkProcess(GetSecret)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/endpoints"), nil, lib.MarkProcess(GetEndpoints)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/mesos/dynamic/clusters/{clusterId}/exportservice"), nil, lib.MarkProcess(GetExportService)})

	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/pod"), nil, lib.MarkProcess(GetPod)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/replicaset"), nil, lib.MarkProcess(GetReplicaSet)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/deployment"), nil, lib.MarkProcess(GetDeploymentK8s)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/service"), nil, lib.MarkProcess(GetServiceK8s)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/configmap"), nil, lib.MarkProcess(GetConfigMapK8s)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/secret"), nil, lib.MarkProcess(GetSecretK8s)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/endpoints"), nil, lib.MarkProcess(GetEndpointsK8s)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/ingress"), nil, lib.MarkProcess(GetIngress)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/namespace"), nil, lib.MarkProcess(GetNameSpace)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/node"), nil, lib.MarkProcess(GetNode)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/daemonset"), nil, lib.MarkProcess(GetDaemonSet)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/job"), nil, lib.MarkProcess(GetJob)})
	actions.RegisterV1Action(actions.Action{"POST", urlPath("/k8s/dynamic/clusters/{clusterId}/statefulset"), nil, lib.MarkProcess(GetStatefulSet)})
}
