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
import logging
from typing import Any, Dict, List, Type

from kubernetes.dynamic.resource import ResourceInstance

from backend.resources.constants import K8sResourceKind, NodeConditionStatus, NodeConditionType
from backend.resources.resource import ResourceClient, ResourceObj
from backend.utils.async_run import async_run
from backend.utils.basic import getitems

logger = logging.getLogger(__name__)


class NodeObj(ResourceObj):
    def __init__(self, data: ResourceInstance):
        super().__init__(data)
        # NOTE: 属性不存在时，返回None
        self.labels = dict(self.data.metadata.labels or {})
        self.taints = [dict(t) for t in self.data.spec.taints or []]

    @property
    def inner_ip(self) -> str:
        """获取inner ip"""
        addresses = self.data.status.addresses
        for addr in addresses:
            if addr.type == "InternalIP":
                return addr.address
        logger.warning("inner ip of addresses is null, address is %s", addresses)
        return ""

    @property
    def node_status(self) -> str:
        """获取节点状态
        ref: https://github.com/kubernetes/dashboard/blob/0de61860f8d24e5a268268b1fbadf327a9bb6013/src/app/backend/resource/node/list.go#L106  # noqa
        """
        for condition in self.data.status.conditions:
            if condition.type != NodeConditionType.Ready:
                continue
            # 正常可用状态
            if condition.status == "True":
                return NodeConditionStatus.Ready
            # 节点不健康而且不能接收 Pod
            return NodeConditionStatus.NotReady
        # 节点控制器在最近 node-monitor-grace-period 期间（默认 40 秒）没有收到节点的消息
        return NodeConditionStatus.Unknown


class Node(ResourceClient):
    """节点 client
    针对节点的查询、操作等
    """

    kind = K8sResourceKind.Node.value
    result_type: Type['ResourceObj'] = NodeObj

    def set_labels_for_multi_nodes(self, node_labels: List[Dict]):
        """设置标签

        :param node_labels: 要设置的标签信息，格式: [{"node_name": "", "labels": {"key": "val"}}]
        NOTE: 如果要删除某个label时，不建议使用replace，可以把要删除的label的值设置为None
        """
        filter_labels = self.filter_nodes_field_data(
            "labels", [label["node_name"] for label in node_labels], node_id_field="name", default_data={}
        )
        # 比对数据，当label在集群节点中存在，而变更的数据中不存在，则需要在变更的数据中设置为None
        for node in node_labels:
            labels = filter_labels.get(node["node_name"])
            # 设置要删除key的值为None
            for key in set(labels) - set(node["labels"]):
                node["labels"][key] = None

        # 下发的body格式: {"metadata": {"labels": {"demo": "demo"}}}
        tasks = [
            functools.partial(self.patch, {"metadata": {"labels": l["labels"]}}, l["node_name"]) for l in node_labels
        ]
        # 当有操作失败的，抛出异常
        async_run(tasks)

    def set_taints_for_multi_nodes(self, node_taints: List[Dict]):
        """设置污点

        :param node_taints: 要设置的污点信息，格式: [{"node_name": "", "taints": [{"key": "", "value": "", "effect": ""}]}]
        """
        # 下发的body格式: {"spec": {"taints": [{"key": xxx, "value": xxx, "effect": xxx}]}}
        tasks = [functools.partial(self.patch, {"spec": {"taints": t["taints"]}}, t["node_name"]) for t in node_taints]
        # 当有操作失败的，抛出异常
        async_run(tasks)

    def set_nodes_schedule_status(self, unschedulable: bool, node_names: List[str]):
        """设置节点调度状态

        unschedulable: 节点是否可以调度
        node_names: 节点名称, 允许多个, 格式[节点名称]
        """
        tasks = [functools.partial(self.patch, {"spec": {"unschedulable": unschedulable}}, n) for n in node_names]
        # 如有失败, 则抛出异常
        async_run(tasks)

    def filter_nodes_field_data(
        self,
        field: str,
        filter_node_names: List[str],
        node_id_field: str = "inner_ip",
        default_data: Any = None,
    ) -> Dict:
        """查询节点属性

        :param field: 查询的属性
        :param filter_node_names: 节点name列表
        :param node_id_field: 节点标识的属性名称，支持name和inner_ip，默认是inner_ip
        :returns: 返回节点的属性数据
        """
        nodes = self.list(is_format=False)
        data = {}
        for node in nodes.items:
            if node.name not in filter_node_names:
                continue
            # 因为field字段可控，先不添加异常处理
            node_id = getattr(node, node_id_field, "")
            data[node_id] = getattr(node, field, default_data)
        return data
