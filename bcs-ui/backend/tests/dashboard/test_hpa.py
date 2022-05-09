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

from backend.dashboard.examples.utils import load_demo_manifest
from backend.tests.conftest import TEST_NAMESPACE, TEST_SHARED_CLUSTER_ID
from backend.tests.dashboard.conftest import DASHBOARD_API_URL_COMMON_PREFIX as DAU_PREFIX
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestHPA:
    """测试 HPA 相关接口"""

    manifest = load_demo_manifest('hpa/simple_hpa')
    create_url = f'{DAU_PREFIX}/hpa/'
    list_url = f'{DAU_PREFIX}/namespaces/{TEST_NAMESPACE}/hpa/'
    inst_url = f"{list_url}{getitems(manifest, 'metadata.name')}/"

    def test_create(self, api_client):
        """测试创建资源接口"""
        response = api_client.post(self.create_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_list(self, api_client):
        """测试获取资源列表接口"""
        response = api_client.get(self.list_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'HorizontalPodAutoscalerList'

    def test_update(self, api_client):
        """测试更新资源接口"""
        # 修改 minReplicas 数量
        self.manifest['spec']['minReplicas'] = 2
        response = api_client.put(self.inst_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_retrieve(self, api_client):
        """测试获取单个资源接口"""
        response = api_client.get(self.inst_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'HorizontalPodAutoscaler'
        assert getitems(response.data, 'manifest.spec.minReplicas') == 2

    def test_destroy(self, api_client):
        """测试删除单个资源"""
        response = api_client.delete(self.inst_url)
        assert response.json()['code'] == 0

    def test_list_shared_cluster_hpa(self, api_client, project_id):
        """获取共享集群 HPA，预期是被拦截（PermissionDenied）"""
        url = f'/api/dashboard/projects/{project_id}/clusters/{TEST_SHARED_CLUSTER_ID}/namespaces/{TEST_NAMESPACE}/hpa/'  # noqa
        assert api_client.get(url).json()['code'] == 400
