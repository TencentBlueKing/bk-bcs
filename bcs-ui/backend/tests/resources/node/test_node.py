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
from unittest import mock

import pytest

from backend.resources.node.client import Node, NodeObj
from backend.tests.testing_utils.base import generate_random_string
from backend.utils import FancyDict

from ..conftest import FakeBcsKubeConfigurationService

fake_inner_ip = "127.0.0.1"
fake_node_name = generate_random_string(8)
fake_labels = {"bcs-test": "test"}
fake_taints = {"key": "test", "value": "tet", "effect": "NoSchedule"}


class TestNodeObj:
    @pytest.fixture(autouse=True)
    def fake_node_data(self):
        self.data = FancyDict(
            metadata=FancyDict(labels=FancyDict()),
            spec=FancyDict(taints=[]),
            status=FancyDict(
                addresses=[FancyDict(address=fake_inner_ip, type="InternalIP")],
                conditions=[FancyDict(status="True", type="Ready")],
            ),
        )

    def test_inner_ip(self):
        assert NodeObj(self.data).inner_ip == fake_inner_ip


class TestNode:
    @pytest.fixture(autouse=True)
    def use_faked_configuration(self):
        with mock.patch(
            'backend.resources.utils.kube_client.BcsKubeConfigurationService',
            new=FakeBcsKubeConfigurationService,
        ):
            yield

    @pytest.fixture
    def client(self, ctx_cluster):
        return Node(ctx_cluster)

    @pytest.fixture
    def create_and_delete_node(self, client):
        client.update_or_create(
            body={
                "apiVersion": "v1",
                "kind": "Node",
                "metadata": {"name": fake_node_name, "labels": fake_labels},
                "spec": {"taints": [fake_taints]},
                "status": {
                    "addresses": [
                        {"address": fake_inner_ip, "type": "InternalIP"},
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
            name=fake_node_name,
        )
        # 等待1s，是为了保证集群内可以查询到正确的节点
        time.sleep(1)
        yield
        client.delete_wait_finished(fake_node_name)

    def test_query_node(self, client, create_and_delete_node):
        nodes = client.list(is_format=False)
        assert len(nodes.data.items) > 0
        assert nodes.metadata
        assert fake_inner_ip in [node.inner_ip for node in nodes.items]

    @pytest.mark.parametrize(
        "field, node_id_field, expected_data",
        [
            ("name", "inner_ip", fake_node_name),
            ("labels", "name", fake_labels),
        ],
    )
    def test_query_nodes_field_data(self, field, node_id_field, expected_data, client, create_and_delete_node):
        data = client.filter_nodes_field_data(field, [fake_node_name], node_id_field=node_id_field)
        node_id = fake_inner_ip if node_id_field == "inner_ip" else fake_node_name
        node_field_data = data[node_id]
        assert node_field_data == expected_data

    @pytest.mark.parametrize(
        "labels, expected",
        [
            ({"bcs-test": "v1"}, {"bcs-test": "v1"}),
            ({"bcs-test": "v1", "bcs-test1": "v2"}, {"bcs-test": "v1", "bcs-test1": "v2"}),
            ({"bcs-test1": "v2"}, {"bcs-test1": "v2"}),
            ({}, {}),
        ],
    )
    def test_set_labels(self, labels, expected, client, create_and_delete_node):
        client.set_labels_for_multi_nodes([{"node_name": fake_node_name, "labels": labels}])
        node_labels = client.filter_nodes_field_data("labels", filter_node_names=[fake_node_name])
        labels = node_labels[fake_inner_ip]
        assert labels == expected

    @pytest.mark.parametrize(
        "taints, expected",
        [
            ([{"key": "test", "value": "", "effect": "NoSchedule"}], {"key": "test", "effect": "NoSchedule"}),
            (
                [{"key": "test", "value": "test", "effect": "NoSchedule"}],
                {"key": "test", "value": "test", "effect": "NoSchedule"},
            ),
        ],
    )
    def test_set_taints(self, taints, expected, client, create_and_delete_node):
        # NOTE: 因为节点通过api直接创建，实际是不正常的，当节点不正常时，k8s会主动设置`NoSchedule`等taint
        # 这样查询节点的taint时，返回的taint中包含期望值则认为正确
        client.set_taints_for_multi_nodes([{"node_name": fake_node_name, "taints": taints}])
        node_taints = client.filter_nodes_field_data("taints", filter_node_names=[fake_node_name])
        taints = node_taints[fake_inner_ip]
        assert expected in taints

    @pytest.mark.parametrize(
        "unschedulable, expected_unschedulable",
        [(True, True), (False, None)],
    )
    def test_set_nodes_schedule_status(self, unschedulable, expected_unschedulable, client, create_and_delete_node):
        client.set_nodes_schedule_status(unschedulable, [fake_node_name])
        nodes = client.list(is_format=False)
        # 查询节点所处的调度状态
        node_unschedulable = None
        for node in nodes.items:
            if node.name not in [fake_node_name]:
                continue
            node_unschedulable = node.data.spec.unschedulable
        assert node_unschedulable == expected_unschedulable
