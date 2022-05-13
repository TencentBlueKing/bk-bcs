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


class TestService:
    """指标：Service 相关测试"""

    def test_list(self, api_client, patch_k8s_client):
        """测试获取 集群指标总览 接口"""
        response = api_client.get(f'/api/metrics/projects/{TEST_PROJECT_ID}/clusters/{TEST_CLUSTER_ID}/services/')
        assert response.json()['code'] == 0
        assert response.json()['data'] == [
            {
                'data': {
                    'apiVersion': 'v1',
                    'kind': 'Service',
                    'metadata': {
                        'creationTimestamp': '2021-04-13T09:12:22Z',
                        'labels': {'app': 'balanced'},
                        'name': 'balanced',
                        'namespace': 'default',
                    },
                    'spec': {
                        'clusterIP': '127.xxx.xxx.1',
                        'clusterIPs': ['127.xxx.xxx.1'],
                        'externalTrafficPolicy': 'Cluster',
                        'ports': [{'nodePort': 30608, 'port': 8080, 'protocol': 'TCP', 'targetPort': 8080}],
                        'selector': {'app': 'balanced'},
                        'sessionAffinity': 'None',
                        'type': 'LoadBalancer',
                    },
                    'status': {'loadBalancer': {'ingress': [{'ip': '127.xxx.xxx.xx9'}, {'hostname': 'localhost'}]}},
                }
            }
        ]
