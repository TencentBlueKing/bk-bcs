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

from backend.tests.conftest import TEST_NAMESPACE

pytestmark = pytest.mark.django_db


class TestSubscribe:
    """ 测试事件订阅接口 """

    def test_watch_namespace_scope_resource(self, api_client, project_id, cluster_id):
        """ 测试 命名空间维度 资源事件 获取接口 """
        url = f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/subscribe/'  # noqa
        params = {'kind': 'Deployment', 'resource_version': 0}
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 400

        params['namespace'] = TEST_NAMESPACE
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0
        response_key = response.json()['data'].keys()
        assert 'events' in response_key
        assert 'latest_rv' in response_key

    def test_watch_cluster_scope_resource(self, api_client, project_id, cluster_id):
        """ 测试 集群维度 资源事件 获取接口 """
        url = f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/subscribe/'  # noqa
        params = {'kind': 'PersistentVolume', 'resource_version': 0}
        response = api_client.get(url, data=params)
        assert response.json()['code'] == 0
