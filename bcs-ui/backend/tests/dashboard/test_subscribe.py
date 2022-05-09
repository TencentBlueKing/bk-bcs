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

from backend.tests.conftest import TEST_CLUSTER_ID, TEST_NAMESPACE, TEST_PROJECT_ID, TEST_SHARED_CLUSTER_ID
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestSubscribe:
    """测试事件订阅接口"""

    subscribe_api_path = '/api/dashboard/projects/{p_id}/clusters/{c_id}/subscribe/'

    def test_watch_namespace_scope_resource(self, api_client):
        """测试 命名空间维度 资源事件 获取接口"""
        url = self.subscribe_api_path.format(p_id=TEST_PROJECT_ID, c_id=TEST_CLUSTER_ID)
        params = {'kind': 'Deployment', 'resource_version': 0}
        response = api_client.get(url, data=params)
        # ValidationError
        assert response.json()['code'] == 400

        params['namespace'] = TEST_NAMESPACE
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0
        response_key = response.json()['data'].keys()
        assert 'events' in response_key
        assert 'latest_rv' in response_key

    def test_watch_cluster_scope_resource(self, api_client):
        """测试 集群维度 资源事件 获取接口"""
        url = self.subscribe_api_path.format(p_id=TEST_PROJECT_ID, c_id=TEST_CLUSTER_ID)
        params = {'kind': 'PersistentVolume', 'resource_version': 0}
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0

    @pytest.mark.parametrize(
        'res_kind',
        [
            'PersistentVolume',
            'PersistentVolumeClaim',
            'StorageClass',
            'CustomResourceDefinition',
            'CustomObject',
            'ServiceAccount',
            'HorizontalPodAutoscaler',
        ],
    )
    def test_watch_shared_cluster_disabled_resource(self, api_client, res_kind):
        """测试获取共享集群禁用资源事件"""
        url = self.subscribe_api_path.format(p_id=TEST_PROJECT_ID, c_id=TEST_SHARED_CLUSTER_ID)
        response = api_client.get(url, {'kind': res_kind, 'resource_version': 0})
        # PermissionDenied
        assert response.json()['code'] == 400

    def test_watch_shared_cluster_deployment(self, api_client, shared_cluster_ns_mgr):
        """测试获取共享集群 Deployment 事件"""
        url = self.subscribe_api_path.format(p_id=TEST_PROJECT_ID, c_id=TEST_SHARED_CLUSTER_ID)
        params = {'kind': 'Deployment', 'resource_version': 0}

        params['namespace'] = 'default'
        response = api_client.get(url, data=params)
        # PermissionDenied
        assert response.json()['code'] == 400

        params['namespace'] = shared_cluster_ns_mgr
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0

    def test_watch_shared_cluster_namespace(self, api_client, shared_cluster_ns_mgr):
        url = self.subscribe_api_path.format(p_id=TEST_PROJECT_ID, c_id=TEST_SHARED_CLUSTER_ID)
        params = {'kind': 'Namespace', 'resource_version': 0}

        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0
        for event in response.json()['data']['events']:
            assert getitems(event, 'manifest.metadata.name').startswith('unittest-proj')
