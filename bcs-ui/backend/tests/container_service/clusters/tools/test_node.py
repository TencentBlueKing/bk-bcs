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

from backend.container_service.clusters.constants import ClusterManagerNodeStatus
from backend.container_service.clusters.tools import node as node_tools
from backend.resources.constants import NodeConditionStatus
from backend.tests.container_service.clusters.test_cc_host import fake_fetch_all_hosts, fake_get_agent_status

FAKE_INNER_IP = "127.0.0.1"
FAKE_NODE_NAME = "bcs-test-node"


def test_query_cluster_nodes(client, create_and_delete_node, ctx_cluster):
    cluster_nodes = node_tools.query_cluster_nodes(ctx_cluster)
    assert FAKE_INNER_IP in cluster_nodes
    assert cluster_nodes[FAKE_INNER_IP]["name"] == FAKE_NODE_NAME
    assert cluster_nodes[FAKE_INNER_IP]["status"] == NodeConditionStatus.Ready
    assert not cluster_nodes[FAKE_INNER_IP]["unschedulable"]


@pytest.mark.parametrize(
    "cluster_node_status,unschedulable,cm_node_status,expected_status",
    [
        (NodeConditionStatus.Ready, False, ClusterManagerNodeStatus.RUNNING, ClusterManagerNodeStatus.RUNNING),
        (NodeConditionStatus.Ready, True, ClusterManagerNodeStatus.RUNNING, ClusterManagerNodeStatus.REMOVABLE),
        (NodeConditionStatus.Ready, True, ClusterManagerNodeStatus.REMOVABLE, ClusterManagerNodeStatus.REMOVABLE),
        (NodeConditionStatus.NotReady, True, ClusterManagerNodeStatus.NOTREADY, ClusterManagerNodeStatus.NOTREADY),
        (NodeConditionStatus.NotReady, True, ClusterManagerNodeStatus.REMOVABLE, ClusterManagerNodeStatus.NOTREADY),
        (NodeConditionStatus.Unknown, True, ClusterManagerNodeStatus.REMOVABLE, ClusterManagerNodeStatus.UNKNOWN),
        ("", False, ClusterManagerNodeStatus.INITIALIZATION, ClusterManagerNodeStatus.INITIALIZATION),
        ("", False, ClusterManagerNodeStatus.DELETING, ClusterManagerNodeStatus.DELETING),
        ("", False, ClusterManagerNodeStatus.ADDFAILURE, ClusterManagerNodeStatus.ADDFAILURE),
        ("", False, ClusterManagerNodeStatus.REMOVEFAILURE, ClusterManagerNodeStatus.REMOVEFAILURE),
    ],
)
def test_transform_status(cluster_node_status, unschedulable, cm_node_status, expected_status):
    assert expected_status == node_tools.transform_status(cluster_node_status, unschedulable, cm_node_status)


@pytest.fixture
def cluster_name():
    return "cluster_name"


class TestNodesData:
    def test_compose_data_by_cm_nodes(self, cm_nodes, cluster_nodes, cluster_id, cluster_name):
        client = node_tools.NodesData(
            cm_nodes=cm_nodes, cluster_nodes=cluster_nodes, cluster_id=cluster_id, cluster_name=cluster_name
        )
        node_data = client._compose_data_by_cm_nodes()
        assert len(node_data) == len(
            [node for inner_ip, node in cm_nodes.items() if node["status"] != ClusterManagerNodeStatus.RUNNING]
        )
        assert node_data[0]["cluster_name"] == cluster_name

    def test_compose_data_by_cluster_nodes(self, cm_nodes, cluster_nodes, cluster_id):
        client = node_tools.NodesData(
            cm_nodes=cm_nodes, cluster_nodes=cluster_nodes, cluster_id=cluster_id, cluster_name="cluster_name"
        )
        node_data = client._compose_data_by_cluster_nodes()
        assert len(node_data) == len(cluster_nodes)
        assert node_data[0]["status"] == ClusterManagerNodeStatus.RUNNING


@pytest.fixture
def master_client(ctx_cluster):
    return node_tools.BcsClusterMaster(ctx_cluster=ctx_cluster, biz_id=1)


class TestBcsClusterMaster:
    @mock.patch("backend.components.cc.HostQueryService.fetch_all", new=fake_fetch_all_hosts)
    @mock.patch("backend.components.gse.get_agent_status", new=fake_get_agent_status)
    def test_list_masters(self, master_client, create_and_delete_master):
        masters = master_client.list_masters()
        # 判断 ip 存在返回的数据中
        detail, is_exist = {}, False
        for master in masters:
            if master["inner_ip"] == FAKE_INNER_IP:
                detail, is_exist = master, True
                break
        assert is_exist
        # 判断包含对应的字段
        for field_name in ["inner_ip", "idc", "rack", "device_class", "bk_cloud_id", "agent"]:
            assert field_name in detail
