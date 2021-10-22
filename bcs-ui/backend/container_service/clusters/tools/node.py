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
from typing import Dict, List

from kubernetes.client import ApiException

from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters import constants as node_constants
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.models import NodeStatus
from backend.resources.constants import NodeConditionStatus
from backend.resources.node.client import Node


def query_cluster_nodes(ctx_cluster: CtxCluster, exclude_master: bool = True) -> Dict:
    """查询节点数据
    包含标签、污点、状态等供前端展示数据
    """
    # 获取集群中的节点列表
    node_client = Node(ctx_cluster)
    try:
        cluster_node_list = node_client.list(is_format=False)
    except ApiException:
        # 查询集群内节点异常，返回空字典
        return {}

    nodes = {}
    for node in cluster_node_list.items:
        labels = node.labels
        # 现阶段节点页面展示及操作，需要排除master
        if exclude_master and labels.get(node_constants.K8S_NODE_ROLE_MASTER) == "true":
            continue

        # 使用inner_ip作为key，主要是方便匹配及获取值
        nodes[node.inner_ip] = {
            "inner_ip": node.inner_ip,
            "name": node.name,
            "status": node.node_status,
            "labels": labels,
            "taints": node.taints,
            "unschedulable": node.data.spec.unschedulable or False,
        }
    return nodes


def query_bcs_cc_nodes(ctx_cluster: CtxCluster) -> Dict:
    """查询bcs cc中的节点数据"""
    client = PaaSCCClient(ComponentAuth(access_token=ctx_cluster.context.auth.access_token))
    node_data = client.get_node_list(ctx_cluster.project_id, ctx_cluster.id)
    return {
        node["inner_ip"]: node
        for node in (node_data.get("results") or [])
        if node["status"] not in [NodeStatus.Removed]
    }


def transform_status(cluster_node_status: str, unschedulable: bool, bcs_cc_node_status: str = None) -> str:
    """转换节点状态"""
    # 如果集群中节点为非正常状态，则返回not_ready
    if cluster_node_status == NodeConditionStatus.NotReady:
        return node_constants.BcsCCNodeStatus.NotReady

    # 如果集群中节点为正常状态，根据是否允许调度，转换状态
    if cluster_node_status == NodeConditionStatus.Ready:
        if unschedulable:
            if bcs_cc_node_status == node_constants.BcsCCNodeStatus.ToRemoved:
                return node_constants.BcsCCNodeStatus.ToRemoved
            return node_constants.BcsCCNodeStatus.Removable
        else:
            return node_constants.BcsCCNodeStatus.Normal

    return node_constants.BcsCCNodeStatus.Unknown


@dataclass
class NodesData:
    bcs_cc_nodes: Dict  # bcs cc中存储的节点数据
    cluster_nodes: Dict  # 集群中实际存在的节点数据
    cluster_id: str
    cluster_name: str

    @property
    def _normal_status(self) -> List:
        return [
            node_constants.BcsCCNodeStatus.Normal,
            node_constants.BcsCCNodeStatus.ToRemoved,
            node_constants.BcsCCNodeStatus.Removable,
        ]

    def nodes(self) -> List:
        """组装节点数据"""
        # 1. 集群中不存在的节点，并且bcs cc中状态处于初始化中、初始化失败、移除中、移除失败状态时，需要展示bcs cc中数据
        # 2. 集群中存在的节点，则以集群中为准，注意状态的转换
        # 把bcs cc中非正常状态节点放到数组的前面，方便用户查看
        node_list = self._compose_data_by_bcs_cc_nodes()
        node_list.extend(self._compose_data_by_cluster_nodes())
        return node_list

    def _compose_data_by_bcs_cc_nodes(self) -> List:
        # 处理在bcs cc中的节点，但是状态为非正常状态数据
        node_list = []
        for inner_ip in self.bcs_cc_nodes:
            node = self.bcs_cc_nodes[inner_ip]
            if (inner_ip in self.cluster_nodes) or (node["status"] in self._normal_status):
                continue
            node["cluster_name"] = self.cluster_name
            node_list.append(node)
        return node_list

    def _compose_data_by_cluster_nodes(self) -> List:
        node_list = []
        # 以集群中数据为准
        for inner_ip, node in self.cluster_nodes.items():
            # 添加集群名称
            node["cluster_name"] = self.cluster_name
            # 如果bcs cc中存在节点信息，则从bcs cc获取节点的额外数据
            if inner_ip in self.bcs_cc_nodes:
                _node = self.bcs_cc_nodes[inner_ip].copy()
                _node.update(node)
                _node["status"] = transform_status(
                    node["status"], node["unschedulable"], self.bcs_cc_nodes[inner_ip]["status"]
                )
                node_list.append(_node)
            else:
                node["cluster_id"] = self.cluster_id
                node["status"] = transform_status(node["status"], node["unschedulable"])
                node_list.append(node)
        return node_list
