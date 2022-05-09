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
from backend.tests.conftest import TEST_SHARED_CLUSTER_ID
from backend.tests.dashboard.conftest import DASHBOARD_API_URL_COMMON_PREFIX as DAU_PREFIX
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestStorageClass:
    """测试 StorageClass 相关接口"""

    manifest = load_demo_manifest('storages/simple_storage_class')
    batch_url = f'{DAU_PREFIX}/storages/storage_classes/'
    detail_url = f"{batch_url}{getitems(manifest, 'metadata.name')}/"

    def test_create(self, api_client):
        """测试创建资源接口"""
        response = api_client.post(self.batch_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_list(self, api_client):
        """测试获取资源列表接口"""
        response = api_client.get(self.batch_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'StorageClassList'

    def test_update(self, api_client):
        """测试更新资源接口"""
        self.manifest['metadata']['annotations'] = {'t_key': 't_val'}
        response = api_client.put(self.detail_url, data={'manifest': self.manifest})
        assert response.json()['code'] == 0

    def test_retrieve(self, api_client):
        """测试获取单个资源接口"""
        response = api_client.get(self.detail_url)
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'StorageClass'
        assert getitems(response.data, 'manifest.metadata.annotations.t_key') == 't_val'

    def test_destroy(self, api_client):
        """测试删除单个资源"""
        response = api_client.delete(self.detail_url)
        assert response.json()['code'] == 0

    def test_list_shared_cluster_sc(self, api_client, project_id):
        """获取共享集群 SC，预期是被拦截（PermissionDenied）"""
        url = f'/api/dashboard/projects/{project_id}/clusters/{TEST_SHARED_CLUSTER_ID}/storages/storage_classes/'
        assert api_client.get(url).json()['code'] == 400
