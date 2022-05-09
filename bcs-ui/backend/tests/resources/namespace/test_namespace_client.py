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

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.namespace import Namespace
from backend.tests.conftest import TEST_CLUSTER_ID, TEST_PROJECT_ID
from backend.tests.testing_utils.base import generate_random_string


class TestNamespaceClient:

    namespace_for_test = 'ns-for-test-{}'.format(generate_random_string(8))

    @pytest.fixture
    def common_patch(self):
        with mock.patch(
            'backend.resources.namespace.client.get_namespaces_by_cluster_id',
            new=lambda *args, **kwargs: [{'id': 1, 'name': 'default'}],
        ), mock.patch(
            'backend.resources.namespace.client.create_cc_namespace',
            new=lambda *args, **kwargs: {'id': 2, 'name': self.namespace_for_test},
        ):
            yield

    def test_get_existed(self, common_patch):
        """测试获取已经存在的命名空间"""
        client = Namespace(CtxCluster.create(TEST_CLUSTER_ID, TEST_PROJECT_ID, token='token'))
        ret = client.get_or_create_cc_namespace('default', 'admin', labels={'test_key': 'test_val'})
        assert ret == {'name': 'default', 'namespace_id': 1}

    def test_create_nonexistent(self, common_patch):
        """测试获取不存在的命名空间（触发创建逻辑）"""
        client = Namespace(CtxCluster.create(TEST_CLUSTER_ID, TEST_PROJECT_ID, token='token'))
        ret = client.get_or_create_cc_namespace(self.namespace_for_test, 'admin', annotations={'test_key': 'test_val'})
        assert ret == {'name': self.namespace_for_test, 'namespace_id': 2}
        client.delete(name=self.namespace_for_test)
