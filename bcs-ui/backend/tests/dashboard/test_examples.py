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

from backend.dashboard.examples.constants import RES_KIND_WITH_DEMO_MANIFEST
from backend.tests.conftest import TEST_CLUSTER_ID, TEST_PROJECT_ID
from backend.utils.basic import getitems

pytestmark = pytest.mark.django_db


class TestResourceExample:
    """ 测试 资源模版 相关接口 """

    common_prefix = f'/api/dashboard/projects/{TEST_PROJECT_ID}/clusters/{TEST_CLUSTER_ID}/examples'

    @pytest.mark.parametrize('resource_kind', RES_KIND_WITH_DEMO_MANIFEST)
    def test_fetch_demo_manifest(self, resource_kind, api_client):
        """ 测试获取资源列表接口 """
        response = api_client.get(f'{self.common_prefix}/manifests/?kind={resource_kind}')
        assert resource_kind == getitems(response.json(), 'data.kind')
        assert response.json()['code'] == 0
