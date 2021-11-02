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
import mock
import pytest

from backend.container_service.clusters.mgr.cluster.master import BcsClusterMaster
from backend.resources.node.client import Node
from backend.tests.container_service.clusters.test_cc_host import fake_fetch_all_hosts, fake_get_agent_status
from backend.tests.testing_utils.base import generate_random_string

fake_name = generate_random_string(8)
fake_inner_ip = "127.0.0.1"


@pytest.fixture
def create_and_delete_master(ctx_cluster):
    client = Node(ctx_cluster)
    client.update_or_create(
        body={
            "apiVersion": "v1",
            "kind": "Node",
            "metadata": {"name": fake_name, "labels": {"node-role.kubernetes.io/master": "true"}},
            "spec": {},
            "status": {
                "addresses": [
                    {"address": fake_inner_ip, "type": "InternalIP"},
                ],
                "conditions": [
                    {
                        "lastHeartbeatTime": "2021-10-25T04:13:48Z",
                        "lastTransitionTime": "2020-10-25T05:24:53Z",
                        "message": "kubelet is posting ready status",
                        "reason": "KubeletReady",
                        "status": "True",
                        "type": "Ready",
                    }
                ],
            },
        },
        name=fake_name,
    )
    yield
    client.delete_wait_finished(fake_name)


@pytest.fixture
def master_client(ctx_cluster):
    return BcsClusterMaster(ctx_cluster=ctx_cluster, biz_id=1)


class TestBcsClusterMaster:
    @mock.patch("backend.components.cc.HostQueryService.fetch_all", new=fake_fetch_all_hosts)
    @mock.patch("backend.components.gse.get_agent_status", new=fake_get_agent_status)
    def test_get_masters(self, master_client, create_and_delete_master):
        masters = master_client.get_masters()
        # 判断 ip 存在返回的数据中
        ip_exist = False
        for master in masters:
            if master["inner_ip"] == fake_inner_ip:
                ip_exist = True
        assert ip_exist

    def test_get_cluster_master(self, master_client, create_and_delete_master):
        cluster_masters = master_client._get_cluster_masters()
        assert fake_inner_ip in cluster_masters
        assert cluster_masters[fake_inner_ip]["host_name"] == fake_name

    @mock.patch("backend.components.cc.HostQueryService.fetch_all", new=fake_fetch_all_hosts)
    def test_get_cc_hosts(self, master_client):
        cc_hosts = master_client._get_cc_hosts_by_ip([fake_inner_ip])
        assert fake_inner_ip in cc_hosts
        # 判断下面field必须存在
        for field_name in ["inner_ip", "idc", "rack", "device_class", "bk_cloud_id"]:
            assert field_name in cc_hosts[fake_inner_ip]

    @mock.patch("backend.components.gse.get_agent_status", new=fake_get_agent_status)
    def test_get_agent_status(self, master_client):
        agent_status = master_client.get_agent_status_by_ip([{"inner_ip": fake_inner_ip, "bk_cloud_id": 0}])
        assert fake_inner_ip in agent_status
        assert "agent" in agent_status[fake_inner_ip]
