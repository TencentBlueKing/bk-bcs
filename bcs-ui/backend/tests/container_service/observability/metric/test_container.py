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


# 通用的API请求参数
pod_name = 'deployment-1'
mock_api_params = {
    'start_at': '2021-01-01 10:00:00',
    'end_at': '2021-01-01 11:00:00',
}


class TestContainerMetric:
    """容器指标相关接口单元测试"""

    common_prefix = '/api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/{pod_name}/containers'.format(
        project_id=TEST_PROJECT_ID, cluster_id=TEST_CLUSTER_ID, pod_name=pod_name
    )

    def test_cpu_limit(self, api_client, container_metric_api_patch):
        """测试获取 CPU 限制 接口"""
        response = api_client.post(f'{self.common_prefix}/cpu_limit/', mock_api_params)
        assert response.json()['code'] == 0

    def test_cpu_usage(self, api_client, container_metric_api_patch):
        """测试获取 CPU 使用情况 接口"""
        response = api_client.post(f'{self.common_prefix}/cpu_usage/', mock_api_params)
        assert response.json()['code'] == 0

    def test_memory_limit(self, api_client, container_metric_api_patch):
        """测试获取 内存限制 接口"""
        response = api_client.post(f'{self.common_prefix}/memory_limit/', mock_api_params)
        assert response.json()['code'] == 0

    def test_memory_usage(self, api_client, container_metric_api_patch):
        """测试获取 内存使用情况 接口"""
        response = api_client.post(f'{self.common_prefix}/memory_usage/', mock_api_params)
        assert response.json()['code'] == 0

    def test_disk_read(self, api_client, container_metric_api_patch):
        """测试获取 磁盘读情况 接口"""
        response = api_client.post(f'{self.common_prefix}/disk_read/', mock_api_params)
        assert response.json()['code'] == 0

    def test_disk_write(self, api_client, container_metric_api_patch):
        """测试获取 磁盘写情况 接口"""
        response = api_client.post(f'{self.common_prefix}/disk_write/', mock_api_params)
        assert response.json()['code'] == 0
