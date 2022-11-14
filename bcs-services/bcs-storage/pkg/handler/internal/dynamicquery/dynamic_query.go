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
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamicquery"
	"google.golang.org/protobuf/types/known/structpb"
)

// 方法名称
const (
	handlerNamespaceK8SRequestTag = "HandlerNamespaceK8SRequest"
)

// filter
var (
	IPPoolStaticFilter       = &dynamicquery.IPPoolStaticFilter{}
	IPPoolStaticDetailFilter = &dynamicquery.IPPoolStaticDetailFilter{}
	podFilter                = &dynamicquery.PodFilter{}
	replicaSetFilter         = &dynamicquery.ReplicaSetFilter{}
	deploymentK8sFilter      = &dynamicquery.DeploymentK8sFilter{}
	serviceK8sFilter         = &dynamicquery.ServiceK8sFilter{}
	configMapK8sFilter       = &dynamicquery.ConfigMapK8sFilter{}
	secretK8sFilter          = &dynamicquery.SecretK8sFilter{}
	endpointsK8sFilter       = &dynamicquery.EndpointsK8sFilter{}
	ingressFilter            = &dynamicquery.IngressFilter{}
	namespaceFilter          = &dynamicquery.NameSpaceFilter{}
	nodeFilter               = &dynamicquery.NodeFilter{}
	daemonSetFilter          = &dynamicquery.DaemonSetFilter{}
	jobFilter                = &dynamicquery.JobFilter{}
	statefulSetFilter        = &dynamicquery.StatefulSetFilter{}
)

// tags
var (
	IPPoolStaticTags       = util.GetStructTags(IPPoolStaticFilter)
	IPPoolStaticDetailTags = util.GetStructTags(IPPoolStaticDetailFilter)
	PodTags                = util.GetStructTags(podFilter)
	ReplicaSetTags         = util.GetStructTags(replicaSetFilter)
	DeploymentK8STags      = util.GetStructTags(deploymentK8sFilter)
	ServiceK8STags         = util.GetStructTags(serviceK8sFilter)
	ConfigMapK8STags       = util.GetStructTags(configMapK8sFilter)
	SecretK8STags          = util.GetStructTags(secretK8sFilter)
	EndpointsK8STags       = util.GetStructTags(endpointsK8sFilter)
	IngressTags            = util.GetStructTags(ingressFilter)
	NamespaceTags          = util.GetStructTags(namespaceFilter)
	NodeTags               = util.GetStructTags(nodeFilter)
	DaemonSetTags          = util.GetStructTags(daemonSetFilter)
	JobTags                = util.GetStructTags(jobFilter)
	StatefulSetTags        = util.GetStructTags(statefulSetFilter)
)

// HandlerIPPoolStaticRequest 获取IPPoolStatic
func HandlerIPPoolStaticRequest(ctx context.Context, req *storage.IPPoolStaticRequest) ([]operator.M, error) {
	raw := util.RequestToMap(IPPoolStaticTags, req)
	query := dynamicquery.NewDynamic(ctx, IPPoolStaticFilter, raw, constants.IPPoolStatic, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerIPPoolStaticDetailRequest 获取IPPoolStaticDetail
func HandlerIPPoolStaticDetailRequest(ctx context.Context, req *storage.IPPoolStaticDetailRequest) ([]operator.M, error) {
	raw := util.RequestToMap(IPPoolStaticDetailTags, req)
	query := dynamicquery.NewDynamic(ctx, IPPoolStaticDetailFilter, raw, constants.IPPoolStaticDetail, req.Field,
		int(req.Offset), int(req.Limit))

	return query.QueryDynamic()
}

// HandlerPodRequest 获取Pod
func HandlerPodRequest(ctx context.Context, req *storage.PodRequest) ([]operator.M, error) {
	raw := util.RequestToMap(PodTags, req)
	query := dynamicquery.NewDynamic(ctx, podFilter, raw, constants.Pod, req.Field, int(req.Offset), int(req.Limit))

	return query.QueryDynamic()
}

// HandlerReplicaSetRequest 获取ReplicaSet
func HandlerReplicaSetRequest(ctx context.Context, req *storage.ReplicaSetRequest) ([]operator.M, error) {
	raw := util.RequestToMap(ReplicaSetTags, req)
	query := dynamicquery.NewDynamic(ctx, replicaSetFilter, raw, constants.ReplicaSet, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerDeploymentK8SRequest 获取K8s Deployment
func HandlerDeploymentK8SRequest(ctx context.Context, req *storage.DeploymentK8SRequest) ([]operator.M, error) {
	raw := util.RequestToMap(DeploymentK8STags, req)
	query := dynamicquery.NewDynamic(ctx, deploymentK8sFilter, raw, constants.DeploymentK8S, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerServiceK8SRequest 获取 K8S Service
func HandlerServiceK8SRequest(ctx context.Context, req *storage.ServiceK8SRequest) ([]operator.M, error) {
	raw := util.RequestToMap(ServiceK8STags, req)
	query := dynamicquery.NewDynamic(ctx, serviceK8sFilter, raw, constants.ServiceK8S, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerConfigMapK8SRequest 获取 K8S ConfigMap
func HandlerConfigMapK8SRequest(ctx context.Context, req *storage.ConfigMapK8SRequest) ([]operator.M, error) {
	raw := util.RequestToMap(ConfigMapK8STags, req)
	query := dynamicquery.NewDynamic(ctx, configMapK8sFilter, raw, constants.ConfigMapK8S, req.Field,
		int(req.Offset), int(req.Limit),
	)

	return query.QueryDynamic()
}

// HandlerSecretK8SRequest 获取 K8S Secret
func HandlerSecretK8SRequest(ctx context.Context, req *storage.SecretK8SRequest) ([]operator.M, error) {
	raw := util.RequestToMap(SecretK8STags, req)
	query := dynamicquery.NewDynamic(ctx, secretK8sFilter, raw, constants.SecretK8S, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerEndpointsK8SRequest 获取 Endpoints K8S
func HandlerEndpointsK8SRequest(ctx context.Context, req *storage.EndpointsK8SRequest) ([]operator.M, error) {
	raw := util.RequestToMap(EndpointsK8STags, req)
	query := dynamicquery.NewDynamic(ctx, endpointsK8sFilter, raw, constants.EndpointsK8S, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerIngressRequest 获取 Ingress
func HandlerIngressRequest(ctx context.Context, req *storage.IngressRequest) ([]operator.M, error) {
	raw := util.RequestToMap(IngressTags, req)
	query := dynamicquery.NewDynamic(ctx, ingressFilter, raw, constants.Ingress, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerNamespaceK8SRequest 获取 K8S Namespace
func HandlerNamespaceK8SRequest(ctx context.Context, req *storage.NamespaceK8SRequest,
	rsp *storage.NamespaceK8SResponse) {
	var err error
	var data interface{}
	raw := util.RequestToMap(NamespaceTags, req)

	if req.Used == "1" {
		raw[constants.FieldTag] = []string{constants.NamespaceTag}
		data, err = dynamicquery.GetK8sNamespaceResource(ctx, raw, int(req.Offset), int(req.Limit))
	} else {
		query := dynamicquery.NewDynamic(ctx, namespaceFilter, raw, constants.NamespaceK8S, req.Field, int(req.Offset),
			int(req.Limit))
		data, err = query.QueryDynamic()
	}

	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("%s %s | err: %v", handlerNamespaceK8SRequestTag, common.BcsErrStorageListResourceFailStr, err)
		return
	}

	if v, ok := data.([]string); ok && len(v) != 0 {
		for _, name := range v {
			v, _ := structpb.NewStruct(util.StructToMap(&storage.Namespace{ResourceName: name}))
			rsp.Data = append(rsp.Data, v)
		}
	} else if v, ok := data.([]operator.M); ok && len(v) != 0 {
		if err = util.ListMapToListStruct(v, &rsp.Data, "HandlerNamespaceK8SRequest"); err != nil {
			rsp.Result = false
			rsp.Code = common.BcsErrStorageReturnDataIsNotJson
			rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
			blog.Errorf("%s %s | err: %v", handlerNamespaceK8SRequestTag, common.BcsErrStorageReturnDataIsNotJsonStr, err)
			return
		}
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr
}

// HandlerNodeRequest 获取 Node
func HandlerNodeRequest(ctx context.Context, req *storage.NodeRequest) ([]operator.M, error) {
	raw := util.RequestToMap(NodeTags, req)
	query := dynamicquery.NewDynamic(ctx, nodeFilter, raw, constants.Node, req.Field, int(req.Offset), int(req.Limit))

	return query.QueryDynamic()
}

// HandlerDaemonSetRequest 获取 DaemonSet
func HandlerDaemonSetRequest(ctx context.Context, req *storage.DaemonSetRequest) ([]operator.M, error) {
	raw := util.RequestToMap(DaemonSetTags, req)
	query := dynamicquery.NewDynamic(ctx, daemonSetFilter, raw, constants.DaemonSet, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}

// HandlerJobRequest 获取 Job
func HandlerJobRequest(ctx context.Context, req *storage.JobRequest) ([]operator.M, error) {
	raw := util.RequestToMap(JobTags, req)
	query := dynamicquery.NewDynamic(ctx, jobFilter, raw, constants.Job, req.Field, int(req.Offset), int(req.Limit))

	return query.QueryDynamic()
}

// HandlerStatefulSetRequest 获取 StatefulSet
func HandlerStatefulSetRequest(ctx context.Context, req *storage.StatefulSetRequest) ([]operator.M, error) {
	raw := util.RequestToMap(StatefulSetTags, req)
	query := dynamicquery.NewDynamic(ctx, statefulSetFilter, raw, constants.StatefulSet, req.Field, int(req.Offset),
		int(req.Limit))

	return query.QueryDynamic()
}
