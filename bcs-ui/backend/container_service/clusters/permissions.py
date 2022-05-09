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

from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType


class AccessClusterPermission(BasePermission):
    """拦截所有共享集群相关的请求"""

    message = '当前请求的 API 在共享集群中不可用'

    def has_permission(self, request, view):
        cluster_id = view.kwargs.get('cluster_id') or request.query_params.get('cluster_id')
        return get_cluster_type(cluster_id) != ClusterType.SHARED


class AccessClusterPermMixin:
    """集群接口访问权限控制"""

    def get_permissions(self):
        # 禁用共享集群相关请求
        return [AccessClusterPermission(), *super().get_permissions()]
