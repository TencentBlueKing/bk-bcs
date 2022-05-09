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


class AccessSvcMonitorNamespacePerm(BasePermission):
    """对于普通集群不做检查，对于共享集群需要检查命名空间是否属于指定项目"""

    message = '在该共享集群中，您没有权限 操作或查看 当前命名空间的 ServiceMonitor'

    def has_permission(self, request, view):
        cluster_type = get_cluster_type(view.kwargs['cluster_id'])
        if cluster_type == ClusterType.SINGLE:
            return True

        # create 使用 request.data 中的 namespace
        if view.action == 'create':
            request_ns = request.data.get('namespace')
        # retrieve, update, destroy 方法使用路径参数中的 namespace
        else:
            request_ns = view.kwargs.get('namespace')

        return is_proj_ns_in_shared_cluster(request.ctx_cluster, request_ns, request.project.english_name)
