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

from backend.accounts import bcs_perm
from backend.container_service.clusters.constants import ClusterType
from backend.container_service.clusters.permissions import AccessClusterPermission  # noqa
from backend.container_service.clusters.utils import get_cluster_type, get_shared_cluster_project_namespaces
from backend.utils.basic import getitems


def validate_cluster_perm(request, project_id: str, cluster_id: str, raise_exception: bool = True) -> bool:
    """ 检查用户是否有操作集群权限 """
    if request.user.is_superuser:
        return True
    perm = bcs_perm.Cluster(request, project_id, cluster_id)
    return perm.can_use(raise_exception=raise_exception)


class AccessNamespacePermission(BasePermission):
    """ 对于普通集群不做检查，对于公共集群需要检查命名空间是否属于指定项目 """

    def has_permission(self, request, view):
        cluster_type = get_cluster_type(view.kwargs['cluster_id'])
        if cluster_type == ClusterType.SINGLE:
            return True

        # list, retrieve, update, destroy 方法使用路径参数中的 namespace，create 方法需要解析 request.data
        if view.action == 'create':
            request_ns = getitems(request.data, 'manifest.metadata.namespace')
        else:
            request_ns = view.kwargs.get('namespace') or request.query_params.get('namespace')

        project_namespaces = get_shared_cluster_project_namespaces(
            request.ctx_cluster.project_id,
            request.project.english_name,
            request.ctx_cluster.id,
            request.user.token.access_token,
        )
        return request_ns in project_namespaces
