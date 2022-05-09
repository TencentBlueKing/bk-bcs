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

from backend.resources.namespace import getitems
from backend.tests.conftest import TEST_SHARED_CLUSTER_ID

pytestmark = pytest.mark.django_db


class TestNamespace:
    """资源视图 命名空间 相关API单元测试"""

    def test_list(self, api_client, project_id, cluster_id):
        """测试获取资源列表接口"""
        response = api_client.get(f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/namespaces/')
        assert response.json()['code'] == 0

    def test_list_shared_cluster_ns(self, api_client, project_id, shared_cluster_ns_mgr):
        """获取共享集群中项目拥有的命名空间"""
        response = api_client.get(
            f'/api/dashboard/projects/{project_id}/clusters/{TEST_SHARED_CLUSTER_ID}/namespaces/'
        )
        namespaces = getitems(response.json(), 'data.manifest.items')
        assert len(namespaces) >= 1
        for ns in namespaces:
            assert getitems(ns, 'metadata.name').startswith('unittest-proj')
