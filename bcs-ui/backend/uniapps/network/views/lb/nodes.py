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
import json
from typing import Dict, List

from backend.container_service.clusters.base import utils as cluster_utils
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.node.client import Node
from backend.uniapps.network.constants import K8S_LB_LABEL, LBLabelOp


class LoadbalancerLabels:
    def __init__(self, ctx_cluster: CtxCluster):
        self.node_client = Node(ctx_cluster)

    def add_labels(self, ip_list: List[str]):
        """添加lb的标签"""
        node_label_list = self._construct_node_label_list(ip_list)
        self.node_client.set_labels_for_multi_nodes(node_label_list)

    def delete_labels(self, ip_list: List[str]):
        node_label_list = self._construct_node_label_list(ip_list, LBLabelOp.DELETE)
        self.node_client.set_labels_for_multi_nodes(node_label_list)

    def _construct_node_label_list(self, ip_list: List[str], op: str = LBLabelOp.ADD) -> List:
        """查询节点的标签"""
        node_list = self.node_client.list(is_format=False)
        # 获取节点标签
        node_label_list = []
        for node in node_list.items:
            if node.inner_ip not in ip_list:
                continue
            labels = self._construct_labels_by_op(node.labels, op)
            node_label_list.append({"node_name": node.name, "labels": labels})
        return node_label_list

    def _construct_labels_by_op(self, labels: Dict, op: str = LBLabelOp.ADD) -> Dict:
        if op == LBLabelOp.ADD:
            labels.update(K8S_LB_LABEL)
        else:
            for key in K8S_LB_LABEL:
                labels[key] = None
        return labels


def convert_ips(access_token: str, project_id: str, cluster_id: str, lb_ips: Dict) -> Dict:
    """转换为ip信息
    NOTE: 历史数据可能为{ip_id: 是否使用}，需要把ip_id转换为ip
    """
    nodes = cluster_utils.get_cluster_nodes(access_token, project_id, cluster_id)
    node_id_ip = {info["id"]: info["inner_ip"] for info in nodes}
    # 通过ip id, 获取ip
    _ips = {}
    for ip, used in lb_ips.items():
        if not ip.isdigit():
            _ips[ip] = used
            continue
        # 如果为字符串数字，则为ID，需要转换为ip
        if node_id_ip.get(int(ip)):
            _ips[node_id_ip[int(ip)]] = used
    return _ips
