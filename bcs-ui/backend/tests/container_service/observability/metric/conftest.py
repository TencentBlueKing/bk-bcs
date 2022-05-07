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
import json

import mock
import pytest
from django.conf import settings

from backend.container_service.observability.metric.constants import MetricDimension
from backend.tests.conftest import TEST_CLUSTER_ID
from backend.tests.resources.formatter.conftest import NETWORK_CONFIG_DIR

# 指标相关配置 目录
METRIC_CONFIG_DIR = f'{settings.BASE_DIR}/backend/tests/container_service/observability/metric/contents/'


@pytest.fixture
def pod_metric_api_patch():
    with mock.patch(
        'backend.container_service.observability.metric.views.pod.PodMetricViewSet._common_query_handler',
        new=lambda *args, **kwargs: None,
    ):
        yield


@pytest.fixture
def container_metric_api_patch():
    with mock.patch(
        'backend.container_service.observability.metric.views.container.ContainerMetricViewSet._common_query_handler',
        new=lambda *args, **kwargs: None,
    ):
        yield


@pytest.fixture
def node_metric_api_patch():
    with mock.patch(
        'backend.container_service.observability.metric.views.node.NodeMetricViewSet._common_query_handler',
        new=lambda *args, **kwargs: None,
    ):
        yield


@pytest.fixture
def node_info_api_patch():
    with mock.patch(
        'backend.container_service.observability.metric.views.node.get_cluster_nodes',
        new=lambda *args, **kwargs: [{'inner_ip': '127.0.0.1'}],
    ), mock.patch(
        'backend.container_service.observability.metric.views.node.prom.get_node_info',
        new=lambda *args, **kwargs: {
            'result': [
                {
                    'metric': {
                        'dockerVersion': 'v1',
                    },
                },
                {
                    'metric': {
                        'osVersion': 'v2',
                    }
                },
                {'metric': {'metric_name': 'cpu_count'}, 'value': [None, '8']},
            ]
        },
    ):
        yield


@pytest.fixture
def node_overview_api_patch():
    MOCK_NODE_DIMENSIONS_FUNC = {
        MetricDimension.CpuUsage: lambda *args, **kwargs: None,
        MetricDimension.MemoryUsage: lambda *args, **kwargs: None,
        MetricDimension.DiskUsage: lambda *args, **kwargs: None,
        MetricDimension.DiskIOUsage: lambda *args, **kwargs: None,
    }
    with mock.patch(
        'backend.container_service.observability.metric.views.node.prom.get_container_pod_count',
        new=lambda *args, **kwargs: {'result': [{'metric': {'metric_name': 'pod_count'}, 'value': [None, '8']}]},
    ), mock.patch(
        'backend.container_service.observability.metric.views.node.constants.NODE_DIMENSIONS_FUNC',
        new=MOCK_NODE_DIMENSIONS_FUNC,
    ):
        yield


@pytest.fixture
def cluster_metric_api_patch():
    MOCK_CLUSTER_DIMENSIONS_FUNC = {
        MetricDimension.CpuUsage: lambda *args, **kwargs: None,
        MetricDimension.MemoryUsage: lambda *args, **kwargs: None,
        MetricDimension.DiskUsage: lambda *args, **kwargs: None,
    }
    with mock.patch(
        'backend.container_service.observability.metric.views.cluster.ClusterMetricViewSet._get_cluster_node_ip_list',
        new=lambda *args, **kwargs: ['127.0.0.1'],
    ), mock.patch(
        'backend.container_service.observability.metric.views.cluster.CLUSTER_DIMENSIONS_FUNC',
        new=MOCK_CLUSTER_DIMENSIONS_FUNC,
    ):
        yield


@pytest.fixture
def target_metric_api_patch():
    with mock.patch(
        'backend.container_service.observability.metric.views.target.get_targets',
        new=lambda *args, **kwargs: {'data': []},
    ):
        yield


@pytest.fixture
def sm_api_patch():
    common_prefix = 'backend.container_service.observability.metric.views.service_monitor'
    with mock.patch(
        f'{common_prefix}.ServiceMonitorMixin._activity_log', new=lambda *args, **kwargs: None
    ), mock.patch(
        f'{common_prefix}.ServiceMonitorMixin._get_cluster_map',
        new=lambda *args, **kwargs: {
            TEST_CLUSTER_ID: {'cluster_id': TEST_CLUSTER_ID, 'name': 'test-cluster', 'environment': 'k8s'}
        },
    ), mock.patch(
        f'{common_prefix}.ServiceMonitorMixin._get_namespace_map',
        new=lambda *args, **kwargs: {(TEST_CLUSTER_ID, 'default'): 1},
    ), mock.patch(
        f'{common_prefix}.ServiceMonitorMixin._single_service_monitor_operate_handler',
        new=lambda *args, **kwargs: None,
    ), mock.patch(
        f'{common_prefix}.ServiceMonitorDetailViewSet._update_manifest', new=lambda _, manifest, params: manifest
    ), mock.patch(
        f'{common_prefix}.ServiceMonitorMixin._validate_namespace_use_perm', new=lambda *args, **kwargs: None
    ):
        yield


class FakeK8SClientForMetric:
    """指标相关 单元测试用的 K8SClient"""

    def __init__(self, *args, **kwargs):
        pass

    def get_service(self, params):
        """获取 Service 列表"""
        with open(f'{NETWORK_CONFIG_DIR}/service.json') as fr:
            configs = json.load(fr)
        return {'data': [{'data': configs['normal']}]}

    def list_service_monitor(self):
        """获取 ServiceMonitor 列表"""
        with open(f'{METRIC_CONFIG_DIR}/service_monitor.json') as fr:
            configs = json.load(fr)
        return {'items': [configs]}

    def get_service_monitor(self, namespace, name):
        """获取单个 ServiceMonitor 信息"""
        with open(f'{METRIC_CONFIG_DIR}/service_monitor.json') as fr:
            configs = json.load(fr)
        return configs


@pytest.fixture
def patch_k8s_client():
    with mock.patch(
        'backend.container_service.observability.metric.views.service.K8SClient', new=FakeK8SClientForMetric
    ), mock.patch(
        'backend.container_service.observability.metric.views.service_monitor.K8SClient', new=FakeK8SClientForMetric
    ):
        yield
