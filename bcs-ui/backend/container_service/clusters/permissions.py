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

from backend.resources.namespace import Namespace, getitems

from .constants import ClusterType
from .utils import get_cluster_type


class IsProjectNamespace(BasePermission):
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
        # TODO 若性能不佳可以考虑添加缓存
        namespaces = [
            getitems(ns, 'metadata.name')
            for ns in Namespace(request.ctx_cluster).list(
                is_format=False, cluster_type=cluster_type, project_code=request.project.english_name
            )['items']
        ]
        return request_ns in namespaces


class DisableCommonClusterRequest(BasePermission):
    """ 拦截所有公共集群相关的请求 """

    def has_permission(self, request, view):
        cluster_id = view.kwargs.get('cluster_id') or request.query_params.get('cluster_id')
        return get_cluster_type(cluster_id) != ClusterType.COMMON
