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

# 单元测试用数据
namespace = 'default'
sm_name = 'test-service-monitor'
create_update_params = {
    "name": "test-metric",
    "namespace": "default",
    "port": 9090,
    "path": "/metric",
    "interval": 30,
    "sample_limit": 50,
    "selector": {"a": "b"},
    "service_name": "default",
}


class TestServiceMonitor:
    """指标：ServiceMonitor 相关测试"""

    common_prefix = '/api/metrics/projects/{project_id}/clusters/{cluster_id}/service_monitors'.format(
        project_id=TEST_PROJECT_ID, cluster_id=TEST_CLUSTER_ID
    )

    def test_list(self, api_client, sm_api_patch, patch_k8s_client):
        """测试获取列表接口"""
        response = api_client.get(f'{self.common_prefix}/?with_perms=false')
        assert response.json()['code'] == 0
        assert set(response.json()['data'][0].keys()) == {
            'namespace_id',
            'metadata',
            'cluster_id',
            'service_name',
            'create_time',
            'spec',
            'cluster_name',
            'environment',
            'status',
            'namespace',
            'name',
            'instance_id',
            'iam_ns_id',
            'is_system',
        }

    def test_create(self, api_client, sm_api_patch):
        """测试创建接口"""
        response = api_client.post(f'{self.common_prefix}/', data=create_update_params)
        assert response.json()['code'] == 0

    def test_batch_delete(self, api_client, sm_api_patch):
        """测试批量删除接口"""
        params = {'service_monitors': [{'namespace': namespace, 'name': sm_name}]}
        response = api_client.delete(f'{self.common_prefix}/batch/', data=params)
        assert response.json()['code'] == 0

    def test_retrieve(self, api_client, sm_api_patch, patch_k8s_client):
        """测试获取单个接口"""
        response = api_client.get(f'{self.common_prefix}/{namespace}/{sm_name}/')
        assert response.json()['code'] == 0

    def test_destroy(self, api_client, sm_api_patch):
        """测试删除接口"""
        response = api_client.delete(f'{self.common_prefix}/{namespace}/{sm_name}/')
        assert response.json()['code'] == 0

    def test_update(self, api_client, sm_api_patch):
        response = api_client.put(f'{self.common_prefix}/{namespace}/{sm_name}/', data=create_update_params)
        assert response.json()['code'] == 0
