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
from typing import Dict

import pytest

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.workloads.deployment import Deployment
from backend.tests.conftest import TEST_CLUSTER_ID, TEST_NAMESPACE, TEST_PROJECT_ID
from backend.tests.testing_utils.base import generate_random_string

pytestmark = pytest.mark.django_db


class TestDeployment:
    """Deployment OpenAPI 相关接口测试"""

    deployment_name = 'deployment-for-test-{}'.format(generate_random_string(8))
    common_prefix = '/apis/resources/projects/{p_id}/clusters/{c_id}/namespaces/{ns}/deployments'.format(
        p_id=TEST_PROJECT_ID, c_id=TEST_CLUSTER_ID, ns=TEST_NAMESPACE
    )

    @pytest.fixture(autouse=True)
    def common_patch(self):
        ctx_cluster = CtxCluster.create(TEST_CLUSTER_ID, TEST_PROJECT_ID, token='token')
        Deployment(ctx_cluster).update_or_create(
            namespace=TEST_NAMESPACE, name=self.deployment_name, body=gen_deployment_body(self.deployment_name)
        )
        yield
        Deployment(ctx_cluster).delete(namespace=TEST_NAMESPACE, name=self.deployment_name)

    def test_list_by_namespace(self, api_client):
        """测试获取指定命名空间下的 Deployment"""
        response = api_client.get(f'{self.common_prefix}/')
        assert response.json()['code'] == 0
        assert isinstance(response.json()['data'], list)

    def test_list_pods_by_deployment(self, api_client):
        """测试获取指定 Deployment 下属 Pod"""
        response = api_client.get(f'{self.common_prefix}/{self.deployment_name}/pods/')
        assert response.json()['code'] == 0
        response_data = response.json()['data']
        assert isinstance(response_data, list)
        # 验证获取到的 Pod 属于指定 Deployment（名称前缀相同）
        assert set(response_data[0].keys()) == {
            'data',
            'clusterId',
            'resourceType',
            'resourceName',
            'namespace',
            'createTime',
            'updateTime',
        }
        # 确保所有的 Pod 都是属于指定的 Deployment 的
        for pod in response_data:
            assert pod['resourceName'].startswith(self.deployment_name)


def gen_deployment_body(name: str) -> Dict:
    """生成用于测试的 Deployment 配置 TODO 后续接入 load_demo_manifest"""
    return {
        'apiVersion': 'apps/v1',
        'kind': 'Deployment',
        'metadata': {'name': name, 'labels': {'app': 'nginx'}},
        'spec': {
            'replicas': 3,
            'selector': {'matchLabels': {'app': 'nginx', 'deploy_name': name}},
            'template': {
                'metadata': {'labels': {'app': 'nginx', 'deploy_name': name}},
                'spec': {'containers': [{'name': 'nginx', 'image': 'nginx:1.14.2', 'ports': [{'containerPort': 80}]}]},
            },
        },
    }
