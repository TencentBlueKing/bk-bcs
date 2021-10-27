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

from backend.components import sops

fake_bk_biz_id = 0
fake_template_id = 1
fake_task_id = 1


class TestSopsClient:
    def test_create_task(self, requests_mock):
        requests_mock.post(ANY, json={"result": True, "data": {"foo": "bar"}})

        client = sops.SopsClient()
        data = sops.CreateTaskParams(name="test")
        resp_data = client.create_task(fake_bk_biz_id, fake_template_id, data)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.called

    def test_start_task(self, requests_mock):
        requests_mock.post(ANY, json={"result": True, "data": {"foo": "bar"}})

        client = sops.SopsClient()
        resp_data = client.start_task(fake_bk_biz_id, fake_task_id)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.request_history[0].method == "POST"

    def test_get_task_status(self, requests_mock):
        requests_mock.get(ANY, json={"result": True, "data": {"foo": "bar"}})

        client = sops.SopsClient()
        resp_data = client.get_task_status(fake_bk_biz_id, fake_task_id)
        assert resp_data == {"foo": "bar"}
        assert requests_mock.request_history[0].method == "GET"

    def test_get_task_node_data(self, requests_mock):
        requests_mock.get(ANY, json={"result": True, "data": {"foo": "bar"}})

        client = sops.SopsClient()
        resp_data = client.get_task_node_data(fake_bk_biz_id, fake_task_id, node_id="test")
        assert resp_data == {"foo": "bar"}
        assert requests_mock.request_history[0].method == "GET"
