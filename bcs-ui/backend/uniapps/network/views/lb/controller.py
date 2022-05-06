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
from typing import Dict, List

from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.node.client import Node
from backend.uniapps.network.constants import K8S_LB_LABEL, LBLabelOp


class LBController:
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


# TODO: 下个迭代重构LB的相关功能，注意去掉不相关的信息
def convert_ip_used_data(
    access_token: str, project_id: str, cluster_id: str, ip_used_data: Dict[str, bool]
) -> Dict[str, bool]:
    """转换为ip信息
    NOTE: 历史数据可能为{ip_id: 是否使用}，需要把ip_id转换为ip

    :param access_token: access_token
    :param project_id: 项目ID
    :param cluster_id: 集群ID
    :param ip_used_data: IP信息，格式为{ip: True}

    :return: 返回IP信息
    """
    nodes = PaaSCCClient(auth=ComponentAuth(access_token)).get_node_list(project_id, cluster_id)
    node_id_ip = {info["id"]: info["inner_ip"] for info in nodes["results"] or []}
    # 通过ip id, 获取ip
    _ip_used_data = {}
    for ip, used in ip_used_data.items():
        try:
            # 临时数据中IP为节点ID，需要转换为对应的IP
            ip_id = int(ip)
            if node_id_ip.get(ip_id):
                _ip_used_data[node_id_ip[ip_id]] = used
        except ValueError:
            _ip_used_data[ip] = used
    return _ip_used_data
