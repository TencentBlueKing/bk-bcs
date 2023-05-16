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

import pytest

from backend.container_service.clusters.constants import ClusterManagerNodeStatus
from backend.resources.node.client import Node
from backend.tests.resources.conftest import FakeBcsKubeConfigurationService

fake_inner_ip = "127.0.0.1"
fake_node_name = "bcs-test-node"
fake_labels = {"bcs-test": "test"}
fake_taints = {"key": "test", "value": "tet", "effect": "NoSchedule"}


@pytest.fixture
def node_name():
    return fake_node_name


@pytest.fixture
def cm_nodes():
    return {
        "127.0.0.1": {"inner_ip": "127.0.0.1", "status": ClusterManagerNodeStatus.INITIALIZATION},
        "127.0.0.2": {"inner_ip": "127.0.0.2", "status": ClusterManagerNodeStatus.RUNNING},
        "127.0.0.3": {"inner_ip": "127.0.0.3", "status": ClusterManagerNodeStatus.ADDFAILURE},
    }


@pytest.fixture
def cluster_nodes():
    return {
        "127.0.0.2": {"inner_ip": "127.0.0.2", "status": "Ready", "unschedulable": False, "node_name": "127.0.0.2"},
        "127.0.0.4": {"inner_ip": "127.0.0.3", "status": "Ready", "unschedulable": False, "node_name": "127.0.0.4"},
    }


@pytest.fixture(autouse=True)
def use_faked_configuration():
    with patch(
        'backend.resources.utils.kube_client.BcsKubeConfigurationService',
        new=FakeBcsKubeConfigurationService,
    ):
        yield


@pytest.fixture
def client(ctx_cluster):
    return Node(ctx_cluster)


@pytest.fixture
def create_and_delete_node(client):
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
    yield
    client.delete_wait_finished(fake_node_name)


@pytest.fixture
def create_and_delete_master(ctx_cluster):
    client = Node(ctx_cluster)
    client.update_or_create(
        body={
            "apiVersion": "v1",
            "kind": "Node",
            "metadata": {"name": fake_node_name, "labels": {"node-role.kubernetes.io/master": "true"}},
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
        name=fake_node_name,
    )
    yield
    client.delete_wait_finished(fake_node_name)
