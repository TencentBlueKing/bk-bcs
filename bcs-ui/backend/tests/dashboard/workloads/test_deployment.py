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
from copy import deepcopy

import pytest

from backend.dashboard.examples.utils import load_demo_manifest
from backend.tests.conftest import TEST_NAMESPACE, TEST_PROJECT_ID, TEST_SHARED_CLUSTER_ID
from backend.tests.dashboard.conftest import DASHBOARD_API_URL_COMMON_PREFIX as DAU_PREFIX
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestDeployment:
    """测试 Deployment 相关接口"""

    manifest = load_demo_manifest('workloads/simple_deployment')
    create_url = f'{DAU_PREFIX}/workloads/deployments/'
    list_url = f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/workloads/deployments/'
    inst_url = f"{list_url}{getitems(manifest, 'metadata.name')}/"
    shared_cluster_url_prefix = f'/api/dashboard/projects/{TEST_PROJECT_ID}/clusters/{TEST_SHARED_CLUSTER_ID}'

    def test_create(self, api_client):
        """测试创建资源接口"""
        response = api_client.post(self.create_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_list(self, api_client):
        """测试获取资源列表接口"""
        response = api_client.get(self.list_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'DeploymentList'

    def test_update(self, api_client):
        """测试更新资源接口"""
        # 修改 replicas 数量
        self.manifest['spec']['replicas'] = 5
        response = api_client.put(self.inst_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_retrieve(self, api_client):
        """测试获取单个资源接口"""
        response = api_client.get(self.inst_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'Deployment'
        assert getitems(response.data, 'manifest.spec.replicas') == 5

    def test_destroy(self, api_client):
        """测试删除单个资源"""
        response = api_client.delete(self.inst_url)
        assert response.json()['code'] == 0

    def test_list_shared_cluster_deploys(self, api_client, shared_cluster_ns_mgr):
        """测试获取共享集群 Deploy"""
        shared_cluster_ns = shared_cluster_ns_mgr

        response = api_client.get(
            f'{self.shared_cluster_url_prefix}/namespaces/{shared_cluster_ns}/workloads/deployments/'
        )
        assert response.json()['code'] == 0

        # 获取不是项目拥有的共享集群命名空间，导致 PermissionDenied
        response = api_client.get(f'{self.shared_cluster_url_prefix}/namespaces/default/workloads/deployments/')
        assert response.json()['code'] == 400

    def test_operate_shared_cluster_deploys(self, api_client, shared_cluster_ns_mgr):
        """测试 创建 / 获取 / 删除 共享集群 Pod"""
        shared_cluster_ns = shared_cluster_ns_mgr

        pc_deploy_manifest = deepcopy(self.manifest)
        pc_deploy_manifest['metadata']['namespace'] = shared_cluster_ns

        pc_create_url = f'{self.shared_cluster_url_prefix}/workloads/deployments/'
        response = api_client.post(pc_create_url, data={'manifest': pc_deploy_manifest})
        assert response.json()['code'] == 0

        pc_inst_url = (
            f"{self.shared_cluster_url_prefix}/namespaces/{shared_cluster_ns}/"
            + f"workloads/deployments/{getitems(pc_deploy_manifest, 'metadata.name')}/"
        )

        # 修改 replicas 数量，测试 Update
        pc_deploy_manifest['spec']['replicas'] = 3
        response = api_client.put(pc_inst_url, data={'manifest': pc_deploy_manifest})
        assert response.json()['code'] == 0

        response = api_client.get(pc_inst_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'Deployment'
        # Retrieve 验证 Update 结果
        assert getitems(response.data, 'manifest.spec.replicas') == 3

        # 回收 Deploy 资源
        response = api_client.delete(pc_inst_url)
        assert response.json()['code'] == 0

    def test_operate_shared_cluster_no_perm_ns_deploys(self, api_client):
        """测试越权操作共享集群不属于项目的命名空间"""
        deploy_manifest = deepcopy(self.manifest)
        deploy_manifest['metadata']['namespace'] = 'default'

        pc_create_url = f'{self.shared_cluster_url_prefix}/workloads/deployments/'
        response = api_client.post(pc_create_url, data={'manifest': deploy_manifest})
        # PermissionDenied
        assert response.json()['code'] == 400

        pc_inst_url = (
            f"{self.shared_cluster_url_prefix}/namespaces/default/"
            + f"workloads/deployments/{getitems(deploy_manifest, 'metadata.name')}/"
        )

        assert api_client.get(pc_inst_url).json()['code'] == 400
        assert api_client.delete(pc_inst_url).json()['code'] == 400
