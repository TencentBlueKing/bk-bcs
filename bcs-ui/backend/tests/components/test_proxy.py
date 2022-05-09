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
from requests_mock import ANY

from backend.components.proxy import ProxyClient, ProxyConfig
from backend.utils import FancyDict

FAKE_PREFIX_PATH = "api/cluster_manager/proxy/"
FAKE_RESPONSE = {"code": 0, "data": {"foo": "bar"}, "result": True, "message": ""}
FAKE_SERVER_HOST = "http://127.0.0.2"

FAKE_PROXY_CONFIG = ProxyConfig(
    host=FAKE_SERVER_HOST,
    request=FancyDict(
        method="GET",
        query_params={"test": "tet"},
        path=f"{FAKE_PREFIX_PATH}/test",
        data={},
    ),
    prefix_path=FAKE_PREFIX_PATH,
)


class TestProxyClient:
    def test_get_source_url(self, requests_mock):
        client = ProxyClient(FAKE_PROXY_CONFIG)
        assert client.source_url == f"{FAKE_SERVER_HOST}/test"

    def test_get_proxy(self, requests_mock):
        requests_mock.get(ANY, json=FAKE_RESPONSE)
        client = ProxyClient(FAKE_PROXY_CONFIG)
        resp_json = client.proxy()
        assert resp_json["code"] == 0
        # key不会变动
        for key in ["code", "data", "result", "message"]:
            assert key in resp_json
