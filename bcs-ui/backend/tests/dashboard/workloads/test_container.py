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

pytestmark = pytest.mark.django_db

namespace, pod_name, container_name = 'default', 'test_pod_name', 'echoserver'


class TestContainer:
    """测试 Container 相关接口"""

    def test_list(self, api_client, project_id, cluster_id, dashboard_container_api_patch):
        """测试获取资源列表接口"""
        response = api_client.get(
            f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/'
            + f'namespaces/{namespace}/workloads/pods/{pod_name}/containers/'
        )
        assert response.json()['code'] == 0
        ret = response.json()['data'][0]
        assert set(ret.keys()) == {'container_id', 'image', 'name', 'status', 'message', 'reason'}

    def test_retrieve(self, api_client, project_id, cluster_id, dashboard_container_api_patch):
        """测试获取单个容器信息"""
        response = api_client.get(
            f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/'
            + f'namespaces/{namespace}/workloads/pods/{pod_name}/containers/{container_name}/'
        )
        assert response.json()['code'] == 0
        assert set(response.json()['data'].keys()) == {
            'host_name',
            'host_ip',
            'container_ip',
            'container_id',
            'container_name',
            'image',
            'network_mode',
            'ports',
            'command',
            'volumes',
            'labels',
            'resources',
        }

    def test_fetch_env_info(self, api_client, project_id, cluster_id, dashboard_container_api_patch):
        """测试获取单个容器环境变量配置信息"""
        response = api_client.get(
            f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}'
            + f'/workloads/pods/{pod_name}/containers/{container_name}/env_info/'
        )
        assert response.json()['code'] == 0
        assert response.json()['data'] == [
            {'name': 'env1', 'value': 'xxx'},
            {'name': 'env2', 'value': 'xxx'},
            {'name': 'env3', 'value': 'xxx'},
        ]
