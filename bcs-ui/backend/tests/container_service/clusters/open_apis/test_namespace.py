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
from unittest.mock import patch

import pytest

from backend.tests.testing_utils.base import generate_random_string
from backend.utils.basic import getitems

RANDOM_NS_NAME = generate_random_string(8)
FAKE_DATA = {'id': 123, 'name': RANDOM_NS_NAME}
# project_id 与 cluster_id随机，防止项目的缓存，导致获取项目类型错误
BASE_URL = f'/apis/resources/projects/{generate_random_string(32)}/clusters/{generate_random_string(8)}'
pytestmark = pytest.mark.django_db


class TestNamespace:
    @pytest.fixture(autouse=True)
    def pre_patch(self):
        with patch(
            'backend.resources.namespace.client.get_namespaces_by_cluster_id', new=lambda *args, **kwargs: []
        ), patch('backend.resources.namespace.client.create_cc_namespace', new=lambda *args, **kwargs: FAKE_DATA):
            yield

    def test_create_namespace(self, api_client):
        """ 测试 open_api 创建命名空间 """
        url = f'{BASE_URL}/namespaces/'
        resp = api_client.post(url, data={'name': RANDOM_NS_NAME, 'labels': {'test_lk': 'test_lv'}})
        assert resp.json()['code'] == 0
        data = resp.json()['data']
        assert 'namespace_id' in data
        assert isinstance(data, dict)
        assert data['name'] == FAKE_DATA['name']

    def test_update_namespace(self, api_client):
        """ 测试 open_api 更新 namespace labels，annotations 信息 """
        url = f'{BASE_URL}/namespaces/{RANDOM_NS_NAME}/'
        resp = api_client.put(url, data={'labels': {'test_lk': 'test_lv1'}, 'annotations': {'test_ak': 'test_av'}})
        assert resp.json()['code'] == 0

    def test_retrieve_namespace(self, api_client):
        """ 测试 open_api 获取 namespace 详情 """
        url = f'{BASE_URL}/namespaces/{RANDOM_NS_NAME}/'
        resp_data = api_client.get(url).json()['data']
        assert getitems(resp_data, 'metadata.labels.test_lk') == 'test_lv1'
        assert getitems(resp_data, 'metadata.annotations.test_ak') == 'test_av'
