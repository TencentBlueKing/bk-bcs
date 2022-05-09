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

from backend.tests.conftest import TEST_SHARED_CLUSTER_ID
from backend.utils.basic import getitems

from .conftest import crd_manifest

pytestmark = pytest.mark.django_db


class TestCRD:
    """测试 CRD 相关接口"""

    crd_name = getitems(crd_manifest, 'metadata.name')

    def test_list(self, api_client, project_id, cluster_id):
        """测试获取资源列表接口"""
        response = api_client.get(f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/crds/v2/')
        assert response.json()['code'] == 0
        response_data = response.json()['data']
        crds = [getitems(item, 'metadata.name') for item in response_data['manifest']['items']]
        assert self.crd_name in crds

    def test_retrieve(self, api_client, project_id, cluster_id):
        """测试获取单个资源接口"""
        response = api_client.get(
            f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/crds/v2/{self.crd_name}/'
        )  # noqa
        assert response.json()['code'] == 0
        assert self.crd_name == getitems(response.json()['data'], 'manifest.metadata.name')

    def test_list_shared_cluster_crd(self, api_client, project_id):
        """获取共享集群 CRD，预期是被拦截（PermissionDenied）"""
        url = f'/api/dashboard/projects/{project_id}/clusters/{TEST_SHARED_CLUSTER_ID}/crds/v2/'
        assert api_client.get(url).json()['code'] == 400
