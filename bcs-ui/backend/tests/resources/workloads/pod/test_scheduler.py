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
from backend.resources.node.client import Node
from backend.resources.workloads.pod.scheduler import PodsRescheduler
from backend.tests.testing_utils.base import generate_random_string
from backend.utils.basic import getitems


class TestPodsRescheduler:
    def test_get_pods(self, ctx_cluster):
        # 获取集群中的一个 Node IP
        host_ips = []
        node = Node(ctx_cluster).list()[0]
        ip_addres = getitems(node, ["data", "status", "addresses"], default=[])
        for info in ip_addres:
            if info.get("type") == "InternalIP":
                host_ips.append(info["address"])
                break
        assert len(host_ips) > 0

        # 获取 pods
        pods = PodsRescheduler(ctx_cluster).list_pods_by_nodes(host_ips)
        assert len(pods) > 0

    def test_reschedule_pods(self, ctx_cluster):
        pods = [
            {"name": generate_random_string(6), "namespace": "default"},
            {"name": generate_random_string(6), "namespace": "default"},
            {"name": generate_random_string(6), "namespace": generate_random_string(6)},
        ]
        results = PodsRescheduler(ctx_cluster).reschedule_pods(ctx_cluster, pods)
        assert len(results) == len(pods)
