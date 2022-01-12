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
import functools
from typing import Dict, List, Optional

from backend.container_service.clusters.base.models import CtxCluster
from backend.utils.async_run import AsyncResult, async_run

from . import Pod


class PodsRescheduler:
    # 限制 pod 并发的数量为100
    MAX_TASK_POD_NUM = 100

    def __init__(self, ctx_cluster: CtxCluster):
        self.ctx_cluster = ctx_cluster

    def reschedule_by_nodes(self, host_ips: List[str], skip_namespaces: Optional[List[str]] = None):
        """重新调度节点上的 pods
        1. 查询集群下所有命名空间的 pods, 过滤出需要重新调度的节点上的 pods; 包含名称和命名空间
        2. 100个pod并行调度，减少对 apiserver 的压力和等待时间
        """
        # 获取指定节点下的 pods
        pods = self.list_pods_by_nodes(host_ips, skip_namespaces)
        # 后台任务处理，删除节点，完成节点的重新调度
        self.reschedule_pods(self.ctx_cluster, pods)

    def list_pods_by_nodes(
        self, host_ips: List[str], skip_namespaces: Optional[List[str]] = None
    ) -> List[Dict[str, str]]:
        """查询节点上的 pods"""
        client = Pod(self.ctx_cluster)
        pods = client.list(is_format=False)
        # 过滤 pod 名称、所属命名空间
        pod_list = []
        for pod in pods.items:
            # 异常处理，避免 pod 还没有分配时，获取不到 status 报错
            try:
                if pod.data.status.hostIP not in host_ips:
                    continue
            except Exception:
                continue
            # 跳过指定的命名空间
            if pod.metadata["namespace"] in (skip_namespaces or []):
                continue
            # 获取pod的名称和命名空间
            pod_list.append(
                {
                    "name": pod.metadata["name"],
                    "namespace": pod.metadata["namespace"],
                }
            )
        return pod_list

    def reschedule_pods(self, ctx_cluster: CtxCluster, pods: List[Dict[str, str]]) -> List[AsyncResult]:
        task_groups = []
        client = Pod(ctx_cluster)
        # 组装任务
        for i in range(0, len(pods), self.MAX_TASK_POD_NUM):
            # 记录组内任务，用于并行处理
            tasks = []
            for pod in pods[i : i + self.MAX_TASK_POD_NUM]:
                tasks.append(functools.partial(client.delete_ignore_nonexistent, pod["name"], pod["namespace"]))
            task_groups.append(tasks)
        # 执行任务
        results = []
        for t in task_groups:
            results.extend(async_run(t, raise_exception=False))
        return results
