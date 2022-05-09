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
from backend.resources.workloads.pod import Pod
from backend.tests.conftest import TEST_CLUSTER_ID, TEST_NAMESPACE, TEST_PROJECT_ID
from backend.tests.testing_utils.base import generate_random_string

pytestmark = pytest.mark.django_db


class TestPod:
    """Pod OpenAPI 相关接口测试"""

    pod_name = 'pod-for-test-{}'.format(generate_random_string(8))

    @pytest.fixture(autouse=True)
    def common_patch(self):
        ctx_cluster = CtxCluster.create(TEST_CLUSTER_ID, TEST_PROJECT_ID, token='token')
        Pod(ctx_cluster).update_or_create(
            namespace=TEST_NAMESPACE, name=self.pod_name, body=gen_pod_body(self.pod_name)
        )
        yield
        Pod(ctx_cluster).delete(namespace=TEST_NAMESPACE, name=self.pod_name)

    def test_get_pod(self, api_client):
        """测试获取指定命名空间下的 Deployment"""
        response = api_client.get(
            '/apis/resources/projects/{p_id}/clusters/{c_id}/namespaces/{ns}/pods/{pod_name}/'.format(
                p_id=TEST_PROJECT_ID, c_id=TEST_CLUSTER_ID, ns=TEST_NAMESPACE, pod_name=self.pod_name
            )
        )
        assert response.json()['code'] == 0
        response_data = response.json()['data']
        assert isinstance(response_data, list)
        assert isinstance(response_data[0]['data'], dict)
        # 测试 ResourceDefaultFormatter 是否生效
        assert set(response_data[0].keys()) == {
            'data',
            'clusterId',
            'resourceType',
            'resourceName',
            'namespace',
            'createTime',
            'updateTime',
        }


def gen_pod_body(name: str) -> Dict:
    """生成用于测试的 Pod 配置 TODO 后续接入 load_demo_manifest"""
    return {
        'apiVersion': 'v1',
        'kind': 'Pod',
        'metadata': {'name': name},
        'spec': {'containers': [{'name': "main", 'image': "busybox"}]},
    }
