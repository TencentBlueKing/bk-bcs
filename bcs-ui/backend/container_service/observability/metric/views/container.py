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
from backend.container_service.observability.metric.constants import (
    METRICS_DEFAULT_CONTAINER_LIST,
    METRICS_DEFAULT_NAMESPACE,
    METRICS_DEFAULT_POD_NAME,
)
from backend.container_service.observability.metric.serializers import FetchContainerMetricSLZ
from backend.utils.url_slug import URL_DEFAULT_PLACEHOLDER


class ContainerMetricViewSet(SystemViewSet):

    serializer_class = FetchContainerMetricSLZ

    def _common_query_handler(
        self, query_metric_func: Callable, cluster_id: str, pod_name: str, need_time_range: bool = True
    ) -> Dict:
        """
        查询容器指标通用逻辑

        :param query_metric_func: 指标查询方法
        :param cluster_id: 集群ID
        :param pod_name: Pod 名称
        :param need_time_range: 是否需要指定时间范围
        :return: 指标查询结果
        """
        params = self.params_validate(self.serializer_class)
        query_params = {
            'cluster_id': cluster_id,
            'namespace': params['namespace'],
            'pod_name': pod_name,
            'container_name': params.get('container_name') or ".*",
        }
        # 部分指标如 Limit 不需要时间范围
        if need_time_range:
            query_params.update(
                {
                    'start': params['start_at'],
                    'end': params['end_at'],
                }
            )

        # 添加业务ID
        query_params['bk_biz_id'] = self.request.project.cc_app_id

        return query_metric_func(**query_params)

    @action(methods=['POST'], url_path='cpu_limit', detail=False)
    def cpu_limit(self, request, project_id, cluster_id, pod_name):
        response_data = self._common_query_handler(
            prom.get_container_cpu_limit, cluster_id, pod_name, need_time_range=False
        )
        return Response(response_data)

    @action(methods=['POST'], url_path='cpu_usage', detail=False)
    def cpu_usage(self, request, project_id, cluster_id, pod_name):
        """获取指定 容器 CPU 使用情况"""
        response_data = self._common_query_handler(prom.get_container_cpu_usage_range, cluster_id, pod_name)
        return Response(response_data)

    @action(methods=['POST'], url_path='memory_limit', detail=False)
    def memory_limit(self, request, project_id, cluster_id, pod_name):
        response_data = self._common_query_handler(
            prom.get_container_memory_limit, cluster_id, pod_name, need_time_range=False
        )
        return Response(response_data)

    @action(methods=['POST'], url_path='memory_usage', detail=False)
    def memory_usage(self, request, project_id, cluster_id, pod_name):
        """获取 容器内存 使用情况"""
        response_data = self._common_query_handler(prom.get_container_memory_usage_range, cluster_id, pod_name)
        return Response(response_data)

    @action(methods=['POST'], url_path='disk_read', detail=False)
    def disk_read(self, request, project_id, cluster_id, pod_name):
        """获取 磁盘读 情况"""
        response_data = self._common_query_handler(prom.get_container_disk_read, cluster_id, pod_name)
        return Response(response_data)

    @action(methods=['POST'], url_path='disk_write', detail=False)
    def disk_write(self, request, project_id, cluster_id, pod_name):
        """获取 磁盘写 情况"""
        response_data = self._common_query_handler(prom.get_container_disk_write, cluster_id, pod_name)
        return Response(response_data)
