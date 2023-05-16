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
from backend.dashboard.custom_object_v2.constants import SHARED_CLUSTER_ENABLED_CRDS
from backend.utils.basic import getitems


class AccessCustomObjectsPermission(BasePermission):
    """检查是否可获取自定义资源"""

    message = '在该共享集群中，您没有查看或操作当前命名空间或该自定义资源的权限'

    def has_permission(self, request, view):
        # 普通独立集群无需检查
        if get_cluster_type(view.kwargs['cluster_id']) == ClusterType.SINGLE:
            return True

        # 共享集群等暂时只允许查询部分自定义资源
        if view.kwargs['crd_name'] not in SHARED_CLUSTER_ENABLED_CRDS:
            return False

        # 检查命名空间是否属于项目且在共享集群中
        # list, retrieve, destroy 方法使用路径参数中的 namespace，create, update 方法需要解析 request.data
        if view.action == 'create':
            request_ns = getitems(request.data, 'manifest.metadata.namespace')
        elif view.action == 'update':
            request_ns = request.data.get('namespace')
        else:
            request_ns = request.query_params.get('namespace')

        return is_proj_ns_in_shared_cluster(request.ctx_cluster, request_ns, request.project.english_name)
