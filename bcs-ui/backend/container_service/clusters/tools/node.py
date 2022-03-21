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

from django.conf import settings

from backend.components import cc, gse
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
    client = ClusterManagerClient(ctx_cluster.context.auth.access_token)
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


class BcsClusterMaster:
    def __init__(self, ctx_cluster: CtxCluster, biz_id: str, username: str = settings.ADMIN_USERNAME):
        self.ctx_cluster = ctx_cluster
        self.biz_id = biz_id
        self.username = username

    def list_masters(self) -> List[Dict]:
        """获取master信息
        1. 查询集群中的ip和name
        2. 通过ip查询主机所在的机房、机架、机型
        3. 通过ip查询主机的agent信息
        4. 组装数据，添加master对应的机房、agent等信息
        """
        cluster_masters = self._get_cluster_masters()
        master_ips = [m["inner_ip"] for m in cluster_masters]
        ip_map_hosts = self._get_ip_map_hosts(master_ips)
        ip_map_agent_status = self._get_ip_map_agent_status(list(ip_map_hosts.values()))
        # 组装数据，追加前端展示需要的机房、机架、机型及agent信息
        for master in cluster_masters:
            inner_ip = master["inner_ip"]
            master.update(ip_map_hosts.get(inner_ip, {}), **ip_map_agent_status.get(inner_ip, {}))
        return cluster_masters

    def _get_cluster_masters(self) -> List[Dict]:
        """查询集群中的master ip和name"""
        node_client = Node(self.ctx_cluster)
        # NOTE: 返回节点出现异常，直接报错
        cluster_nodes = node_client.list(is_format=False)
        # 过滤 master 信息
        masters = []
        for node in cluster_nodes.items:
            labels = node.labels
            # 排除非master节点
            if labels.get(node_constants.K8S_NODE_ROLE_MASTER) != "true":
                continue
            masters.append({"inner_ip": node.inner_ip, "host_name": node.name})
        return masters

    def _get_ip_map_hosts(self, inner_ips: List[str]) -> Dict[str, Dict]:
        """通过 IP 查询主机信息
        包含: 机房、机架、机型
        """
        host_property_filter = {
            "condition": "OR",
            "rules": [{"field": "bk_host_innerip", "operator": "equal", "value": inner_ip} for inner_ip in inner_ips],
        }
        try:
            hosts = cc.HostQueryService(
                self.username, self.biz_id, host_property_filter=host_property_filter
            ).fetch_all()
        except Exception as e:
            logger.error("查询主机信息失败，%s", e)
            # 忽略异常，直接返回为空
            return {}
        # 组装机房、机架、机型数据
        default_cloud_id = 0
        return {
            host["bk_host_innerip"]: {
                "inner_ip": host["bk_host_innerip"],
                "idc": host.get("idc_name"),
                "rack": host.get("rack"),
                "device_class": host.get("svr_device_class"),
                "bk_cloud_id": host.get("bk_cloud_id", default_cloud_id),
            }
            for host in hosts
        }

    def _get_ip_map_agent_status(self, hosts: List[Dict]) -> Dict[str, Dict]:
        """通过 IP 查询主机 agent 状态"""
        # 主机为空时，直接返回
        if not hosts:
            return {}
        params = [{"ip": host["inner_ip"], "bk_cloud_id": host["bk_cloud_id"]} for host in hosts]
        try:
            agents = gse.get_agent_status(self.username, params)
        except Exception as e:
            logger.error("查询主机agent信息失败，%s", e)
            return {}
        # 如果返回状态字段缺失，则认为agent状态异常，其中0表示agent不在线
        return {
            agent["ip"]: {"agent": agent.get("bk_agent_alive", node_constants.DEFAULT_BK_AGENT_ALIVE)}
            for agent in agents
        }
