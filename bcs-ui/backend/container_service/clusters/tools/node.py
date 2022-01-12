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
import logging
from dataclasses import dataclass
from typing import Dict, List

from backend.components.base import ComponentAuth
from backend.components.cluster_manager import ClusterManagerClient
from backend.container_service.clusters import constants as node_constants
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.constants import NodeConditionStatus
from backend.resources.node.client import Node

logger = logging.getLogger(__name__)


def query_cluster_nodes(ctx_cluster: CtxCluster, exclude_master: bool = True) -> Dict:
    """查询节点数据
    包含标签、污点、状态等供前端展示数据
    """
    # 获取集群中的节点列表
    # NOTE: 现阶段会有两个agent，新版agent上报集群信息到bcs api中，可能会有时延，导致bcs api侧找不到集群信息；处理方式:
    # 1. 初始化流程调整，创建集群时，注册一次集群信息
    # 2. 应用侧，兼容处理异常
    try:
        cluster_node_list = Node(ctx_cluster).list(is_format=False)
    except Exception as e:  # 兼容处理现阶段kube-agent没有注册时，连接不上集群的异常
        logger.error("query cluster nodes error, %s", e)
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


def query_nodes_from_cm(ctx_cluster: CtxCluster) -> Dict:
    """通过 cluster manager 查询节点数据
    目的是展示初始化中、初始化失败、删除中、删除失败的节点
    """
    client = ClusterManagerClient(ComponentAuth(access_token=ctx_cluster.context.auth.access_token))
    try:
        node_list = client.get_nodes(ctx_cluster.id)
    except Exception as e:
        logger.error("通过 cluster manager 查询节点数据异常，%s", e)
        return {}
    return {
        node["innerIP"]: {"inner_ip": node["innerIP"], "cluster_id": node["clusterID"], "status": node["status"]}
        for node in node_list
    }


def transform_status(cluster_node_status: str, unschedulable: bool, cm_node_status: str = None) -> str:
    """转换节点状态"""
    NODE_STATRUS = node_constants.ClusterManagerNodeStatus
    # 节点处于初始化中、初始化失败、删除中、删除失败时，任务需要继续处理agent、dns等，因此，需要展示bcs cc中的状态
    if cm_node_status in [
        NODE_STATRUS.INITIALIZATION,
        NODE_STATRUS.ADDFAILURE,
        NODE_STATRUS.DELETING,
        NODE_STATRUS.REMOVEFAILURE,
    ]:
        return cm_node_status

    # 如果集群中节点为非正常状态，则返回not_ready
    if cluster_node_status == NodeConditionStatus.NotReady:
        return NODE_STATRUS.NOTREADY

    # 如果集群中节点为正常状态，根据是否允许调度，转换状态
    if cluster_node_status == NodeConditionStatus.Ready:
        if unschedulable:
            return NODE_STATRUS.REMOVABLE
        else:
            return NODE_STATRUS.RUNNING

    return NODE_STATRUS.UNKNOWN


@dataclass
class NodesData:
    cm_nodes: Dict  # cluster manager 中存储的节点数据
    cluster_nodes: Dict  # 集群中实际存在的节点数据
    cluster_id: str
    cluster_name: str

    def nodes(self) -> List:
        """组装节点数据"""
        # 1. 集群中不存在的节点，并且在cluster manager中状态处于初始化中、初始化失败、移除中、移除失败状态时，需要展示cluster manager中数据
        # 2. 集群中存在的节点，则以集群中为准，注意状态的转换
        # 把cluster manager中非正常状态节点放到数组的前面，方便用户查看
        node_list = self._compose_data_by_cm_nodes()
        node_list.extend(self._compose_data_by_cluster_nodes())
        return node_list

    def _compose_data_by_cm_nodes(self) -> List:
        # 处理在 cluster manager 中的节点，但是状态为非正常状态数据
        node_list = []
        for inner_ip, node in self.cm_nodes.items():
            if inner_ip in self.cluster_nodes or node["status"] in [
                node_constants.ClusterManagerNodeStatus.RUNNING,
                node_constants.ClusterManagerNodeStatus.REMOVABLE,
            ]:
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
            # 如果 cluster manager 中存在节点信息，则从 cluster manager 中获取节点的额外数据
            if inner_ip in self.cm_nodes:
                _node = self.cm_nodes[inner_ip].copy()
                _node.update(node)
                _node["status"] = transform_status(
                    node["status"], node["unschedulable"], self.cm_nodes[inner_ip]["status"]
                )
                node_list.append(_node)
            else:
                node["cluster_id"] = self.cluster_id
                node["status"] = transform_status(node["status"], node["unschedulable"])
                node_list.append(node)
        return node_list
