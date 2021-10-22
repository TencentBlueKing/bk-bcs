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

from backend.container_service.clusters.constants import BcsCCNodeStatus
from backend.container_service.clusters.tools import node as node_tools
from backend.resources.constants import NodeConditionStatus

fake_inner_ip = "127.0.0.1"
fake_node_name = "bcs-test-node"


def test_query_cluster_nodes(client, create_and_delete_node, ctx_cluster):
    cluster_nodes = node_tools.query_cluster_nodes(ctx_cluster)
    assert fake_inner_ip in cluster_nodes
    assert cluster_nodes[fake_inner_ip]["name"] == fake_node_name
    assert cluster_nodes[fake_inner_ip]["status"] == NodeConditionStatus.Ready
    assert not cluster_nodes[fake_inner_ip]["unschedulable"]


@pytest.mark.parametrize(
    "cluster_node_status,unschedulable,bcs_cc_node_status,expected_status",
    [
        (NodeConditionStatus.Ready, False, BcsCCNodeStatus.Normal, BcsCCNodeStatus.Normal),
        (NodeConditionStatus.Ready, True, BcsCCNodeStatus.Normal, BcsCCNodeStatus.Removable),
        (NodeConditionStatus.Ready, True, BcsCCNodeStatus.ToRemoved, BcsCCNodeStatus.ToRemoved),
        (NodeConditionStatus.NotReady, True, BcsCCNodeStatus.NotReady, BcsCCNodeStatus.NotReady),
        (NodeConditionStatus.NotReady, True, BcsCCNodeStatus.Removable, BcsCCNodeStatus.NotReady),
        (NodeConditionStatus.Unknown, True, BcsCCNodeStatus.Removable, BcsCCNodeStatus.Unknown),
    ],
)
def test_transform_status(cluster_node_status, unschedulable, bcs_cc_node_status, expected_status):
    assert expected_status == node_tools.transform_status(cluster_node_status, unschedulable, bcs_cc_node_status)


@pytest.fixture
def cluster_name():
    return "cluster_name"


class TestNodesData:
    def test_compose_data_by_bcs_cc_nodes(self, bcs_cc_nodes, cluster_nodes, cluster_id, cluster_name):
        client = node_tools.NodesData(
            bcs_cc_nodes=bcs_cc_nodes, cluster_nodes=cluster_nodes, cluster_id=cluster_id, cluster_name=cluster_name
        )
        node_data = client._compose_data_by_bcs_cc_nodes()
        assert len(node_data) == len(
            [node for inner_ip, node in bcs_cc_nodes.items() if node["status"] != BcsCCNodeStatus.Normal]
        )
        assert node_data[0]["cluster_name"] == cluster_name

    def test_compose_data_by_cluster_nodes(self, bcs_cc_nodes, cluster_nodes, cluster_id):
        client = node_tools.NodesData(
            bcs_cc_nodes=bcs_cc_nodes, cluster_nodes=cluster_nodes, cluster_id=cluster_id, cluster_name="cluster_name"
        )
        node_data = client._compose_data_by_cluster_nodes()
        assert len(node_data) == len(cluster_nodes)
        assert node_data[0]["status"] == BcsCCNodeStatus.Normal
