# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
from rest_framework.response import Response

from backend.bcs_web.viewsets import UserViewSet
from backend.container_service.clusters.permissions import AccessClusterPermMixin
from backend.resources.constants import K8sResourceKind
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.resources.utils.kube_client import make_labels_string
from backend.resources.workloads.deployment import Deployment
from backend.resources.workloads.pod import Pod
from backend.utils.basic import getitems


class DeploymentViewSet(AccessClusterPermMixin, UserViewSet):
    def list_by_namespace(self, request, project_id_or_code, cluster_id, namespace):
        # TODO 增加用户对层级资源project/cluster/namespace的权限校验
        deployments = Deployment(request.ctx_cluster).list(namespace=namespace, is_format=False)
        return Response(deployments.data.to_dict()['items'])

    def list_pods_by_deployment(self, request, project_id_or_code, cluster_id, namespace, deploy_name):
        # TODO 增加用户对层级资源project/cluster/namespace的权限校验(由于粒度没有细化到Deployment)
        deployment = Deployment(request.ctx_cluster).get(namespace=namespace, name=deploy_name, is_format=False)
        labels_string = make_labels_string(getitems(deployment.data.to_dict(), 'spec.selector.matchLabels', {}))
        pods = Pod(request.ctx_cluster).list(
            namespace=namespace,
            label_selector=labels_string,
            is_format=False,
            owner_kind=K8sResourceKind.Deployment.value,
            owner_name=deploy_name,
        )['items']
        return Response(ResourceDefaultFormatter().format_list(pods))
