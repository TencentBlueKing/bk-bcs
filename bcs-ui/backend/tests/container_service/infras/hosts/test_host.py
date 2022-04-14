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

from backend.container_service.infras.hosts import host
from backend.container_service.infras.hosts.perms import can_use_hosts
from backend.container_service.infras.hosts.terraform.engines import sops
from backend.container_service.infras.hosts.terraform.engines.sops import HostData
from backend.tests.bcs_mocks.fake_sops import FakeSopsMod, sops_json

fake_cc_host_ok_results = [{"bk_host_innerip": "127.0.0.1,127.0.0.3"}, {"bk_host_innerip": "127.0.0.2"}]
fake_cc_host_null_results = []
fake_cc_host_not_match_results = [{"bk_host_innerip": "127.0.0.1"}]
expect_used_ip_list = ["127.0.0.1", "127.0.0.2"]


class TestCheckUseHost:
    @patch("backend.container_service.infras.hosts.perms.get_has_perm_hosts", return_value=fake_cc_host_ok_results)
    def test_ok(self, biz_id, username):
        assert can_use_hosts(biz_id, username, expect_used_ip_list)

    @patch("backend.container_service.infras.hosts.perms.get_has_perm_hosts", return_value=fake_cc_host_null_results)
    def test_null_resp_failed(self, biz_id, username):
        assert not can_use_hosts(biz_id, username, expect_used_ip_list)

    @patch(
        "backend.container_service.infras.hosts.perms.get_has_perm_hosts", return_value=fake_cc_host_not_match_results
    )
    def test_not_match_failed(self, biz_id, username):
        assert not can_use_hosts(biz_id, username, expect_used_ip_list)


class TestGetAgentStatus:
    @patch(
        "backend.container_service.infras.hosts.host.gse.get_agent_status",
        return_value=[
            {"ip": "127.0.0.1", "bk_cloud_id": 0, "bk_agent_alive": 1},
            {"ip": "127.0.0.2", "bk_cloud_id": 0, "bk_agent_alive": 1},
            {"ip": "127.0.0.3", "bk_cloud_id": 0, "bk_agent_alive": 1},
        ],
    )
    def test_get_agent_status(self, mocker):
        host_list = [
            host.HostData(inner_ip="127.0.0.1", bk_cloud_id=0),
            host.HostData(inner_ip="127.0.0.2,127.0.0.3", bk_cloud_id=0),
        ]
        agent_data = host.get_agent_status("admin", host_list)
        # 因为有一个主机两个网卡: 127.0.0.2, 127.0.0.3
        assert len(agent_data) == 3
        assert {"ip": "127.0.0.3", "bk_cloud_id": 0, "bk_agent_alive": 1} in agent_data


try:
    from .test_host_ext import *  # noqa
except ImportError as e:
    pass


class TestApplyHostApi:
    fake_params = {
        "cc_app_id": "1",
        "username": "admin",
        "region": "ap-nanjing",
        "cvm_type": "cvm_type",
        "disk_size": 100,
        "replicas": 1,
        "vpc_name": "vpc_test",
    }

    @patch(
        "backend.container_service.infras.hosts.terraform.engines.sops.sops.SopsClient",
        new=FakeSopsMod,
    )
    def test_create_and_start_host_application(self):
        username = "admin"
        cc_app_id = "1"
        host_data = HostData(
            region="ap-nanjing",
            vpc_name="vpc_test",
            cvm_type="test",
            disk_type='test',
            disk_size=100,
            replicas=1,
            zone_id='ap-nanjing-1',
        )
        task_id, task_url = sops.create_and_start_host_application(username, cc_app_id, host_data)
        assert task_id == sops_json.fake_task_id
        assert task_url == sops_json.fake_task_url

    @patch("backend.container_service.infras.hosts.terraform.engines.sops.sops.SopsClient", new=FakeSopsMod)
    def test_get_task_state_and_steps(self):
        status_and_steps = sops.get_task_state_and_steps(sops_json.fake_task_id)
        assert status_and_steps["state"] == "RUNNING"
        assert status_and_steps["steps"]["申请CVM服务器"]["state"] == "FINISHED"
        assert "<class 'pipeline.core.flow.event.EmptyStartEvent'>" not in status_and_steps["steps"]
        assert "<class 'pipeline.core.flow.event.EmptyEndEvent'>" not in status_and_steps["steps"]
