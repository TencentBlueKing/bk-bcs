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

from rest_framework.decorators import action
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components import bcs_monitor as prom
from backend.container_service.observability.metric.constants import METRICS_DEFAULT_NAMESPACE
from backend.container_service.observability.metric.serializers import FetchPodMetricSLZ


class PodMetricViewSet(SystemViewSet):

    serializer_class = FetchPodMetricSLZ

    def _common_query_handler(self, query_metric_func: Callable, cluster_id: str) -> Dict:
        """
        查询Pod指标通用逻辑

        :param query_metric_func: 指标查询方法
        :param cluster_id: 集群ID
        :return: 指标查询结果
        """
        params = self.params_validate(self.serializer_class)
        return query_metric_func(
            cluster_id,
            params['namespace'],
            params['pod_name_list'],
            params['start_at'],
            params['end_at'],
            bk_biz_id=self.request.project.cc_app_id,
        )

    @action(methods=['POST'], url_path='cpu_usage', detail=False)
    def cpu_usage(self, request, project_id, cluster_id):
        """获取指定 Pod CPU 使用情况"""
        response_data = self._common_query_handler(prom.get_pod_cpu_usage_range, cluster_id)
        return Response(response_data)

    @action(methods=['POST'], url_path='memory_usage', detail=False)
    def memory_usage(self, request, project_id, cluster_id):
        """获取 Pod 内存使用情况"""
        response_data = self._common_query_handler(prom.get_pod_memory_usage_range, cluster_id)
        return Response(response_data)

    @action(methods=['POST'], url_path='network_receive', detail=False)
    def network_receive(self, request, project_id, cluster_id):
        """获取 网络入流量 情况"""
        response_data = self._common_query_handler(prom.get_pod_network_receive, cluster_id)
        return Response(response_data)

    @action(methods=['POST'], url_path='network_transmit', detail=False)
    def network_transmit(self, request, project_id, cluster_id):
        """获取 网络出流量 情况"""
        response_data = self._common_query_handler(prom.get_pod_network_transmit, cluster_id)
        return Response(response_data)
