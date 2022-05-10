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
from unittest import mock

import pytest

from backend.resources.namespace.namespace_quota import NamespaceQuota
from backend.tests.conftest import TEST_NAMESPACE

from ..conftest import FakeBcsKubeConfigurationService


class TestNamespaceQuota:
    @pytest.fixture(autouse=True)
    def use_faked_configuration(self):
        """Replace ConfigurationService with fake object"""
        with mock.patch(
            'backend.resources.utils.kube_client.BcsKubeConfigurationService',
            new=FakeBcsKubeConfigurationService,
        ):
            yield

    @pytest.fixture
    def client_obj(self, project_id, cluster_id):
        return NamespaceQuota('token', project_id, cluster_id)

    def test_create_namespace_quota(self, client_obj):
        assert not client_obj.get_namespace_quota(TEST_NAMESPACE)
        client_obj.create_namespace_quota(TEST_NAMESPACE, {'cpu': '1000m'})

    def test_get_namespace_quota(self, client_obj):
        """测试获取 单个 NamespaceQuota"""
        quota = client_obj.get_namespace_quota(TEST_NAMESPACE)
        assert isinstance(quota, dict)
        assert 'hard' in quota

    def test_list_namespace_quota(self, client_obj):
        """测试获取 NamespaceQuota 列表"""
        results = client_obj.list_namespace_quota(TEST_NAMESPACE)
        assert len(results) > 0

    def test_update_or_create_namespace_quota(self, client_obj):
        """
        测试 NamespaceQuota 的 更新或创建
        TODO create_namespace_quota 与 update_or_create_namespace_quota 逻辑相同，后续考虑废弃一个
        """
        client_obj.update_or_create_namespace_quota(TEST_NAMESPACE, {'cpu': '2000m'})

        quota = client_obj.get_namespace_quota(TEST_NAMESPACE)
        assert isinstance(quota, dict)

    def test_delete_namespace_quota(self, client_obj):
        client_obj.delete_namespace_quota(TEST_NAMESPACE)
        assert not client_obj.get_namespace_quota(TEST_NAMESPACE)
