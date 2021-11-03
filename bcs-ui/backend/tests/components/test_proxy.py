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

from backend.components.proxy import ProxyConfig, ProxyClient
from backend.utils import FancyDict

fake_prefix_path = "api/cluster_manager/proxy/"
fake_response = {"code": 0, "data": {"foo": "bar"}, "result": True, "message": ""}
fake_server_host = "http://127.0.0.2"

fake_proxy_config = ProxyConfig(
    host=fake_server_host,
    request=FancyDict(
        method="GET",
        query_params={"test": "tet"},
        path=f"{fake_prefix_path}/test",
        data={},
    ),
    prefix_path=fake_prefix_path,
)


class TestProxyClient:
    def test_get_source_url(self, requests_mock):
        client = ProxyClient(fake_proxy_config)
        assert client.source_url == f"{fake_server_host}/test"

    def test_get_proxy(self, requests_mock):
        requests_mock.get(ANY, json=fake_response)
        client = ProxyClient(fake_proxy_config)
        resp_json = client.proxy()
        assert resp_json["code"] == 0
        # key不会变动
        for key in ["code", "data", "result", "message"]:
            assert key in resp_json
