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
from typing import List

from django.utils.translation import ugettext_lazy as _
from rest_framework.decorators import action
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components import bcs_monitor as prom
from backend.container_service.clusters.base.utils import get_cluster_nodes
from backend.container_service.observability.metric.constants import CLUSTER_DIMENSIONS_FUNC, MetricDimension
from backend.container_service.observability.metric.serializers import FetchMetricOverviewSLZ
from backend.utils.error_codes import error_codes


class ClusterMetricViewSet(SystemViewSet):
    """集群相关指标"""

    @action(methods=['POST'], url_path='overview', detail=False)
    def overview(self, request, project_id, cluster_id):
        """集群指标总览"""
        params = self.params_validate(FetchMetricOverviewSLZ)
        node_ip_list = self._get_cluster_node_ip_list(project_id, cluster_id)
        if not node_ip_list:
            return Response({})

        # 默认使用3个维度（不含磁盘IO），若指定则使用指定的维度
        dimensions = params.get('dimensions') or [dim for dim in MetricDimension if dim != MetricDimension.DiskIOUsage]

        response_data = {}
        for dimension in dimensions:
            if dimension not in CLUSTER_DIMENSIONS_FUNC:
                raise error_codes.APIError(_("节点指标维度 {} 不合法").format(dimension))

            dimension_func = CLUSTER_DIMENSIONS_FUNC[dimension]
            response_data[dimension] = dimension_func(cluster_id, node_ip_list, bk_biz_id=request.project.cc_app_id)

        return Response(response_data)

    @action(methods=['GET'], url_path='cpu_usage', detail=False)
    def cpu_usage(self, request, project_id, cluster_id):
        """集群 CPU 使用情况"""
        node_ip_list = self._get_cluster_node_ip_list(project_id, cluster_id)
        if not node_ip_list:
            return Response({})

        response_data = prom.get_cluster_cpu_usage_range(cluster_id, node_ip_list, bk_biz_id=request.project.cc_app_id)
        return Response(response_data)

    @action(methods=['GET'], url_path='memory_usage', detail=False)
    def memory_usage(self, request, project_id, cluster_id):
        """集群 内存 使用情况"""
        node_ip_list = self._get_cluster_node_ip_list(project_id, cluster_id)
        if not node_ip_list:
            return Response({})

        response_data = prom.get_cluster_memory_usage_range(
            cluster_id, node_ip_list, bk_biz_id=request.project.cc_app_id
        )
        return Response(response_data)

    @action(methods=['GET'], url_path='disk_usage', detail=False)
    def disk_usage(self, request, project_id, cluster_id):
        """集群 磁盘 使用情况"""
        node_ip_list = self._get_cluster_node_ip_list(project_id, cluster_id)
        if not node_ip_list:
            return Response({})

        response_data = prom.get_cluster_disk_usage_range(
            cluster_id, node_ip_list, bk_biz_id=request.project.cc_app_id
        )
        return Response(response_data)

    def _get_cluster_node_ip_list(self, project_id: str, cluster_id: str) -> List:
        """
        获取指定集群下属节点 IP 列表

        :param project_id: 项目 ID
        :param cluster_id: 集群 ID
        :return: Node IP 列表
        """
        node_list = get_cluster_nodes(self.request.user.token.access_token, project_id, cluster_id)
        return [node['inner_ip'] for node in node_list]
