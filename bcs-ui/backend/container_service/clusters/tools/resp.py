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

from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.constants import K8S_RESERVED_KEY_WORDS
from backend.resources.node.client import Node


class NodeRespBuilder:
    """构造节点 API 返回

    TODO: 现阶段返回先方便前端处理，拆分后，调整返回为{manifest: xxx, manifest_ext: xxx}
          manifest中放置原始node数据, manifest_ext中存放处理的状态等数据
    """

    def __init__(self, ctx_cluster: CtxCluster):
        self.client = Node(ctx_cluster)

    def list_nodes(self) -> Dict:
        """查询节点列表"""
        nodes = self.client.list(is_format=False)
        return {
            "manifest": nodes.data.to_dict(),
            "manifest_ext": {
                node.metadata["uid"]: {
                    "status": node.node_status,
                    "labels": {key: "readonly" for key in filter_label_keys(node.labels.keys())},
                }
                for node in nodes.items
            },
        }

    def query_labels(self, node_names: List[str]) -> Dict[str, Dict]:
        """查询节点标签
        TODO: 这里是兼容处理，方便前端使用，后续前端直接通过列表获取数据
        """
        node_labels = self.client.filter_nodes_field_data("labels", filter_node_names=node_names, default_data={})
        return node_labels


def filter_label_keys(label_keys: List) -> List:
    """过滤满足条件的标签key"""
    return list(filter(is_reserved_label_key, label_keys))


def is_reserved_label_key(label_key: str) -> bool:
    """判断label是否匹配
    NOTE: 现阶段包含指定字符串的label，认为是预留的label，不允许编辑
    """
    for key_word in K8S_RESERVED_KEY_WORDS:
        # k8s预留的标签key的格式: xxx.key_word/xxx
        if label_key.split("/")[0].endswith(key_word):
            return True
    return False
