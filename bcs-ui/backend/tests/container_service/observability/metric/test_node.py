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
import pytest

from backend.tests.conftest import TEST_CLUSTER_ID, TEST_PROJECT_ID

pytestmark = pytest.mark.django_db


class TestNodeMetric:
    """ 节点指标相关接口单元测试 """

    common_prefix = '/api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}'.format(
        project_id=TEST_PROJECT_ID, cluster_id=TEST_CLUSTER_ID, node_ip='127.0.0.1'
    )

    def test_overview(self, api_client, node_overview_api_patch):
        """ 测试获取 节点指标总览 接口 """
        response = api_client.post(
            f'{self.common_prefix}/overview/', data={'dimensions': ['cpu_usage', 'memory_usage']}
        )
        assert response.json()['code'] == 0
        assert response.json()['data'] == {
            'container_count': '0',
            'pod_count': '8',
            'cpu_usage': None,
            'memory_usage': None,
        }

    def test_info(self, api_client, node_info_api_patch):
        """ 测试获取 节点信息 接口 """
        response = api_client.get(f'{self.common_prefix}/info/')
        assert response.json()['code'] == 0
        assert response.json()['data'] == {
            'provider': 'Prometheus',
            'dockerVersion': 'v1',
            'osVersion': 'v2',
            'cpu_count': '8',
        }

    def test_cpu_usage(self, api_client, node_metric_api_patch):
        """ 测试获取 CPU 使用情况 接口 """
        response = api_client.get(f'{self.common_prefix}/cpu_usage/')
        assert response.json()['code'] == 0

    def test_memory_usage(self, api_client, node_metric_api_patch):
        """ 测试获取 内存使用情况 接口 """
        response = api_client.get(f'{self.common_prefix}/memory_usage/')
        assert response.json()['code'] == 0

    def test_network_receive(self, api_client, node_metric_api_patch):
        """ 测试获取 网络入流量 接口 """
        response = api_client.get(f'{self.common_prefix}/network_receive/')
        assert response.json()['code'] == 0

    def test_network_transmit(self, api_client, node_metric_api_patch):
        """ 测试获取 网络出流量 接口 """
        response = api_client.get(f'{self.common_prefix}/network_transmit/')
        assert response.json()['code'] == 0

    def test_diskio_usage(self, api_client, node_metric_api_patch):
        """ 测试获取 磁盘读写情况 接口 """
        response = api_client.get(f'{self.common_prefix}/diskio_usage/')
        assert response.json()['code'] == 0
