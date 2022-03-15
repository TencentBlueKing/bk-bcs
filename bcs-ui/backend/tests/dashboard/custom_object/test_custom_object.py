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
from backend.tests.dashboard.conftest import DASHBOARD_API_URL_COMMON_PREFIX as DAU_PREFIX
from backend.utils.basic import getitems

from .conftest import cobj_manifest, crd_manifest

pytestmark = pytest.mark.django_db


class TestCustomObject:
    """ 测试 CustomObject 相关接口 """

    crd_name = getitems(crd_manifest, 'metadata.name')
    cobj_name = getitems(cobj_manifest, 'metadata.name')
    batch_url = f'{DAU_PREFIX}/crds/v2/{crd_name}/custom_objects/'
    detail_url = f'{DAU_PREFIX}/crds/v2/{crd_name}/custom_objects/{cobj_name}/'

    def test_create(self, api_client):
        """ 测试创建资源接口 """
        response = api_client.post(self.batch_url, data={'manifest': cobj_manifest})
        assert response.json()['code'] == 0

    def test_list(self, api_client):
        """ 测试获取资源列表接口 """
        response = api_client.get(self.batch_url, data={'namespace': 'default'})
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'CronTab4TestList'

    def test_update(self, api_client):
        """ 测试更新资源接口 """
        # 修改 cronSpec
        cobj_manifest['spec']['cronSpec'] = '* * * * */5'
        response = api_client.put(self.detail_url, data={'manifest': cobj_manifest, 'namespace': 'default'})
        assert response.json()['code'] == 0

    def test_retrieve(self, api_client):
        """ 测试获取单个资源接口 """
        response = api_client.get(self.detail_url, data={'namespace': 'default'})
        assert response.json()['code'] == 0
        assert response.data['manifest']['kind'] == 'CronTab4Test'
        assert getitems(response.data, 'manifest.spec.cronSpec') == '* * * * */5'

    def test_destroy(self, api_client):
        """ 测试删除单个资源 """
        response = api_client.delete(self.detail_url + '?namespace=default')
        assert response.json()['code'] == 0

    def test_list_shared_cluster_cobj(self, api_client, project_id):
        """ 获取共享集群 cobj，预期是被拦截（PermissionDenied） """
        url = f'/api/dashboard/projects/{project_id}/clusters/{TEST_SHARED_CLUSTER_ID}/crds/v2/{self.crd_name}/custom_objects/'  # noqa
        assert api_client.get(url).json()['code'] == 400
