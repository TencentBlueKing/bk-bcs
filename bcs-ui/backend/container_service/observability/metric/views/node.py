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
from typing import Callable, Dict

from django.utils.translation import ugettext_lazy as _
from rest_framework.decorators import action
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components import bcs_monitor as prom
from backend.container_service.clusters.base.utils import get_cluster_nodes
from backend.container_service.observability.metric import constants
from backend.container_service.observability.metric.serializers import BaseMetricSLZ, FetchMetricOverviewSLZ
from backend.utils.error_codes import error_codes
from backend.utils.url_slug import IPV4_REGEX


class NodeMetricViewSet(SystemViewSet):

    lookup_field = 'node_ip'
    # 指定匹配 IPV4 地址
    lookup_value_regex = IPV4_REGEX

    @action(methods=['POST'], url_path='overview', detail=True)
    def overview(self, request, project_id, cluster_id, node_ip):
        """节点指标总览"""
        params = self.params_validate(FetchMetricOverviewSLZ)

        # 默认包含 container_count, pod_count
        response_data = {'container_count': '0', 'pod_count': '0'}

        container_pod_count = prom.get_container_pod_count(cluster_id, node_ip, bk_biz_id=request.project.cc_app_id)
        for count in container_pod_count.get('result') or []:
            for k, v in count['metric'].items():
                if k == 'metric_name' and count['value']:
                    response_data[v] = count['value'][1]

        # 默认使用全维度，若指定则使用指定的维度
        dimensions = params.get('dimensions') or [dim for dim in constants.MetricDimension]

        for dimension in dimensions:
            if dimension not in constants.NODE_DIMENSIONS_FUNC:
                raise error_codes.APIError(_("节点指标维度 {} 不合法").format(dimension))

            dimension_func = constants.NODE_DIMENSIONS_FUNC[dimension]
            response_data[dimension] = dimension_func(cluster_id, node_ip, bk_biz_id=request.project.cc_app_id)

        return Response(response_data)

    @action(methods=['GET'], url_path='info', detail=True)
    def info(self, request, project_id, cluster_id, node_ip):
        """节点基础指标信息"""
        node_list = get_cluster_nodes(request.user.token.access_token, project_id, cluster_id)
        node_ip_list = [node["inner_ip"] for node in node_list]

        if node_ip not in node_ip_list:
            raise error_codes.ValidateError(_('IP {} 不合法或不属于当前集群').format(node_ip))

        response_data = {'provider': 'Prometheus'}
        for info in prom.get_node_info(cluster_id, node_ip, bk_biz_id=request.project.cc_app_id).get('result') or []:
            for k, v in info['metric'].items():
                if k in constants.NODE_UNAME_METRIC:
                    response_data[k] = v
                elif k == 'metric_name' and v in constants.NODE_USAGE_METRIC:
                    response_data[v] = info['value'][1] if info['value'] else '0'

        return Response(response_data)

    @action(methods=['GET'], url_path='cpu_usage', detail=True)
    def cpu_usage(self, request, project_id, cluster_id, node_ip):
        """节点 CPU 使用率"""
        response_data = self._common_query_handler(prom.get_node_cpu_usage_range, cluster_id, node_ip)
        return Response(response_data)

    @action(methods=['GET'], url_path='memory_usage', detail=True)
    def memory_usage(self, request, project_id, cluster_id, node_ip):
        """节点 内存 使用率"""
        response_data = self._common_query_handler(prom.get_node_memory_usage_range, cluster_id, node_ip)
        return Response(response_data)

    @action(methods=['GET'], url_path='network_receive', detail=True)
    def network_receive(self, request, project_id, cluster_id, node_ip):
        """节点 网络入流量"""
        response_data = self._common_query_handler(prom.get_node_network_receive, cluster_id, node_ip)
        return Response(response_data)

    @action(methods=['GET'], url_path='network_transmit', detail=True)
    def network_transmit(self, request, project_id, cluster_id, node_ip):
        """节点 网络出流量"""
        response_data = self._common_query_handler(prom.get_node_network_transmit, cluster_id, node_ip)
        return Response(response_data)

    @action(methods=['GET'], url_path='diskio_usage', detail=True)
    def diskio_usage(self, request, project_id, cluster_id, node_ip):
        """磁盘 IO 使用情况"""
        response_data = self._common_query_handler(prom.get_node_diskio_usage_range, cluster_id, node_ip)
        return Response(response_data)

    def _common_query_handler(self, query_metric_func: Callable, cluster_id: str, node_ip) -> Dict:
        """
        查询 Node 指标通用逻辑

        :param query_metric_func: 指标查询方法
        :param cluster_id: 集群 ID
        :param node_ip: 节点 IP
        :return: 指标查询结果
        """
        params = self.params_validate(BaseMetricSLZ)
        return query_metric_func(
            cluster_id, node_ip, params['start_at'], params['end_at'], bk_biz_id=self.request.project.cc_app_id
        )
