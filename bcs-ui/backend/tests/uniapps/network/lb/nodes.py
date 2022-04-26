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
import time

import pytest
from mock import patch

from backend.resources.node.client import Node
from backend.tests.testing_utils.base import generate_random_string
from backend.uniapps.network.constants import K8S_LB_LABEL
from backend.uniapps.network.views.lb.controller import LBController, convert_ip_used_data

FAKE_IP_ID = 1
FAKE_INNER_IP = "127.0.0.1"
FAKE_NODE_NAME = generate_random_string(8)


class TestLbController:
    @pytest.fixture
    def client(self, ctx_cluster):
        return Node(ctx_cluster)

    @pytest.fixture
    def create_and_delete_node(self, client):
        client.update_or_create(
            body={
                "apiVersion": "v1",
                "kind": "Node",
                "metadata": {"name": FAKE_NODE_NAME},
                "spec": {},
                "status": {
                    "addresses": [
                        {"address": FAKE_INNER_IP, "type": "InternalIP"},
                    ],
                    "conditions": [
                        {
                            "lastHeartbeatTime": "2021-07-07T04:13:48Z",
                            "lastTransitionTime": "2020-09-16T05:24:53Z",
                            "message": "kubelet is posting ready status",
                            "reason": "KubeletReady",
                            "status": "True",
                            "type": "Ready",
                        }
                    ],
                },
            },
            name=FAKE_NODE_NAME,
        )
        # 等待1s，是为了保证集群内可以查询到正确的节点
        time.sleep(5)
        yield
        client.delete_wait_finished(FAKE_NODE_NAME)

    def test_add_lb_labels(self, client, create_and_delete_node, ctx_cluster):
        LBController(ctx_cluster).add_labels(FAKE_INNER_IP)
        # 查询label
        node_labels = client.filter_nodes_field_data("labels", filter_node_names=[FAKE_NODE_NAME])
        labels = node_labels[FAKE_INNER_IP]
        # 查询lb需要的label已经存在
        for key, val in K8S_LB_LABEL.items():
            assert key in labels
            assert val == labels[key]

    def test_del_lb_labels(self, client, create_and_delete_node, ctx_cluster):
        # 先添加lb的标签
        LBController(ctx_cluster).add_labels(FAKE_INNER_IP)
        # 检查标签存在
        node_labels = client.filter_nodes_field_data("labels", filter_node_names=[FAKE_NODE_NAME])
        labels = node_labels[FAKE_INNER_IP]
        assert set(K8S_LB_LABEL.keys()).issubset(set(labels.keys()))
        # 删除标签
        LBController(ctx_cluster).delete_labels(FAKE_INNER_IP)
        # 检查标签已经删除
        node_labels = client.filter_nodes_field_data("labels", filter_node_names=[FAKE_NODE_NAME])
        labels = node_labels[FAKE_INNER_IP]
        assert not set(K8S_LB_LABEL.keys()).issubset(set(labels.keys()))


@pytest.mark.parametrize(
    "ip_used_data, converted_ip_used_data",
    [({str(FAKE_IP_ID): True}, {FAKE_INNER_IP: True}), ({FAKE_INNER_IP: True}, {FAKE_INNER_IP: True})],
)
def test_convert_ip_used_data(request_user, project_id, cluster_id, ip_used_data, converted_ip_used_data):
    with patch(
        "backend.uniapps.network.views.lb.nodes.PaaSCCClient.get_node_list",
        return_value={"results": [{"id": 1, "inner_ip": FAKE_INNER_IP}]},
    ):
        assert (
            convert_ip_used_data(request_user.token.access_token, project_id, cluster_id, ip_used_data)
            == converted_ip_used_data
        )
