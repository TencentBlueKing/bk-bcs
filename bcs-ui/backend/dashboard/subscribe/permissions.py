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
from rest_framework.permissions import BasePermission

from backend.container_service.clusters.base.utils import get_cluster_type, is_proj_ns_in_shared_cluster
from backend.container_service.clusters.constants import ClusterType
from backend.resources.constants import K8sResourceKind

from .constants import SHARED_CLUSTER_SUBSCRIBEABLE_RESOURCE_KINDS


class IsSubscribeable(BasePermission):
    """检查当前指定参数是否支持订阅"""

    def has_permission(self, request, view):
        project_id, cluster_id = view.kwargs['project_id'], view.kwargs['cluster_id']
        cluster_type = get_cluster_type(cluster_id)
        if cluster_type == ClusterType.SINGLE:
            return True

        # 只有指定的数类资源可以执行订阅功能
        res_kind = request.query_params.get('kind')
        if res_kind not in SHARED_CLUSTER_SUBSCRIBEABLE_RESOURCE_KINDS:
            return False

        # 命名空间可以直接查询，但是不属于项目的需要被过滤掉
        if res_kind == K8sResourceKind.Namespace.value:
            return True

        # 可以执行订阅功能的资源，也需要检查命名空间是否属于指定的项目
        request_ns = request.query_params.get('namespace')
        return is_proj_ns_in_shared_cluster(request.ctx_cluster, request_ns, request.project.english_name)
