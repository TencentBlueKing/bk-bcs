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
from dataclasses import dataclass
from typing import List

from mock import patch

from backend.resources.workloads.pod.scheduler import PodsRescheduler
from backend.tests.testing_utils.base import generate_random_string
from backend.utils import FancyDict


@dataclass
class NodePods:
    items: List[FancyDict]


FAKE_POD_NAME = generate_random_string(6)
FAKE_HOST_IP = "127.0.0.1"
FAKE_NAMESPACE = "default"
FAKE_PODS = NodePods(
    items=[
        FancyDict(
            data=FancyDict(status=FancyDict(hostIP=FAKE_HOST_IP)),
            metadata={"name": FAKE_POD_NAME, "namespace": FAKE_NAMESPACE},
        )
    ]
)


class TestPodsRescheduler:
    @patch("backend.resources.workloads.pod.scheduler.Pod.list", return_value=FAKE_PODS)
    def test_pods_by_existed_node(self, ctx_cluster):
        # 通过节点 IP，可以过滤到 pods
        pods = PodsRescheduler(ctx_cluster).list_pods_by_nodes([FAKE_HOST_IP])
        assert len(pods) == 1
        # 校验字段
        assert pods[0]["name"] == FAKE_POD_NAME
        assert pods[0]["namespace"] == FAKE_NAMESPACE

    @patch("backend.resources.workloads.pod.scheduler.Pod.list", return_value=FAKE_PODS)
    def test_pods_by_not_exist_node(self, ctx_cluster):
        # 通过节点 IP，过滤不到 pods
        pods = PodsRescheduler(ctx_cluster).list_pods_by_nodes(["127.0.0.2"])
        assert len(pods) == 0

    @patch("backend.resources.workloads.pod.scheduler.Pod.delete_ignore_nonexistent", return_value=None)
    def test_task_group(self, ctx_cluster):
        pods = [
            {"name": generate_random_string(6), "namespace": "default"},
            {"name": generate_random_string(6), "namespace": "default"},
            {"name": generate_random_string(6), "namespace": generate_random_string(6)},
        ]
        results = PodsRescheduler(ctx_cluster).reschedule_pods(ctx_cluster, pods)
        assert len(results) == len(pods)
