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
from requests_mock import ANY

from backend.components.base import ComponentAuth
from backend.components.bcs_api import BcsApiClient

BCS_APIGW_TOKEN = 'example-auth-token'


@pytest.fixture(autouse=True)
def setup_token(settings):
    settings.BCS_APIGW_TOKEN = BCS_APIGW_TOKEN


class TestBcsApiClient:
    def test_get_cluster_simple(self, project_id, cluster_id, requests_mock):
        requests_mock.get(ANY, json={'id': 'foo-id'})

        client = BcsApiClient(ComponentAuth('fake_token'))
        result = client.query_cluster_id('stag', project_id, cluster_id)
        assert result == 'foo-id'

        req_history = requests_mock.request_history[0]
        # Assert token was in request headers and access_token was in query string
        assert req_history.headers.get('Authorization') == f"Bearer {BCS_APIGW_TOKEN}"
        assert 'access_token=fake_token' in req_history.url

    def test_get_cluster_credentials(self, requests_mock):
        requests_mock.get(ANY, json={'name': 'foo'})

        client = BcsApiClient(ComponentAuth('fake_token'))
        resp = client.get_cluster_credentials('stag', 'fake-bcs-cluster-foo')
        assert resp == {'name': 'foo'}
