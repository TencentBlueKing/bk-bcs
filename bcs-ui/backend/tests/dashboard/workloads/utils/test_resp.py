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

from backend.dashboard.workloads.utils.resp import ContainerRespBuilder
from backend.tests.dashboard.conftest import gen_mock_pod_manifest

pytestmark = pytest.mark.django_db

# 测试用数据
pod_manifest = gen_mock_pod_manifest()
container_name = 'echoserver'


class TestContainerRespBuilder:
    """测试容器信息构造逻辑"""

    def test_build_list(self):
        """测试组装 Container 列表信息方法"""
        ret = ContainerRespBuilder(pod_manifest).build_list()
        assert ret == [
            {
                'container_id': '651e64xxxxx',
                'image': 'k8s.gcr.io/echoserver:1.4',
                'name': 'echoserver',
                'status': 'running',
                'message': 'running',
                'reason': 'running',
            }
        ]

    def test_build(self):
        """测试组装单个 Container 信息方法"""
        ret = ContainerRespBuilder(pod_manifest, container_name).build()
        assert ret == {
            'host_name': 'minikube',
            'host_ip': '127.xxx.xxx.xxx',
            'container_ip': '127.xxx.xxx.xxx',
            'container_id': '651e64xxxxx',
            'container_name': 'echoserver',
            'image': 'k8s.gcr.io/echoserver:1.4',
            'network_mode': 'ClusterFirst',
            'ports': [],
            'command': {
                'command': '',
                'args': '',
            },
            'volumes': [
                {
                    'host_path': 'default-token-kvb6t',
                    'mount_path': '/var/run/secrets/kubernetes.io/serviceaccount',
                    'readonly': True,
                }
            ],
            'labels': [
                {
                    'key': 'app',
                    'val': 'balanced',
                },
                {
                    'key': 'pod-template-hash',
                    'val': '5744b548b4',
                },
            ],
            'resources': {},
        }
