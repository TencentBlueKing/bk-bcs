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
import time

import mock
import pytest

from backend.dashboard.examples.utils import load_demo_manifest
from backend.tests.conftest import TEST_NAMESPACE
from backend.tests.dashboard.conftest import DASHBOARD_API_URL_COMMON_PREFIX as DAU_PREFIX
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestPod:
    """ 测试 Pod 相关接口 """

    manifest = load_demo_manifest('workloads/simple_pod')
    create_url = f'{DAU_PREFIX}/workloads/pods/'
    list_url = f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/workloads/pods/'
    inst_url = f"{list_url}{getitems(manifest, 'metadata.name')}/"

    def test_create(self, api_client):
        """ 测试创建资源接口 """
        response = api_client.post(self.create_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_list(self, api_client):
        """ 测试获取资源列表接口 """
        response = api_client.get(self.list_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'PodList'

    def test_retrieve(self, api_client):
        """ 测试获取单个资源接口 """
        response = api_client.get(self.inst_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'Pod'

    def test_destroy(self, api_client):
        """ 测试删除单个资源 """
        response = api_client.delete(self.inst_url)
        assert response.json()['code'] == 0

    def test_list_pod_pvcs(self, api_client, patch_pod_client):
        """ 测试获取 Pod 关联 PersistentVolumeClaim """
        response = api_client.get(f'{self.inst_url}pvcs/')
        assert response.json()['code'] == 0

    def test_list_pod_configmaps(self, api_client, patch_pod_client):
        """ 测试获取 Pod 关联 ConfigMap """
        response = api_client.get(f'{self.inst_url}configmaps/')
        assert response.json()['code'] == 0

    def test_list_pod_secrets(self, api_client, patch_pod_client):
        """ 测试获取单个资源接口 """
        response = api_client.get(f'{self.inst_url}secrets/')
        assert response.json()['code'] == 0

    def test_reschedule(self, api_client):
        """
        测试重新调度 Pod
        TODO 可考虑 mock 掉下发集群操作，仅验证接口功能
        """
        # 创建有父级资源的 Pod，测试重新调度
        deploy_manifest = load_demo_manifest('workloads/simple_deployment')
        deploy_name = deploy_manifest['metadata']['name']
        api_client.post(f'{DAU_PREFIX}/workloads/deployments/', data={'manifest': deploy_manifest})
        # 等待 Deployment 下属 Pod 创建
        time.sleep(3)
        # 找到 Deployment 下属的 第一个 Pod Name
        resp = api_client.get(
            f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/workloads/pods/',
            data={'label_selector': 'app=nginx', 'owner_kind': 'Deployment', 'owner_name': deploy_name},
        )
        pods = getitems(resp.json(), 'data.manifest.items', [])
        pod_name = getitems(pods[0], 'metadata.name')
        resp = api_client.put(f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/workloads/pods/{pod_name}/reschedule/')
        assert resp.json()['code'] == 0
        assert getitems(resp.json(), 'data.metadata.name') == pod_name
        # 清理测试用的资源
        resp = api_client.delete(f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/workloads/deployments/{deploy_name}/')
        assert resp.json()['code'] == 0
