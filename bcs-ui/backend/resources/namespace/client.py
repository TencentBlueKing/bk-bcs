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
from typing import Dict

from backend.resources.constants import K8sResourceKind
from backend.resources.namespace.formatter import NamespaceFormatter
from backend.resources.namespace.utils import create_cc_namespace, get_namespaces_by_cluster_id
from backend.resources.resource import ResourceClient


class Namespace(ResourceClient):
    kind = K8sResourceKind.Namespace.value
    formatter = NamespaceFormatter()

    def get_or_create_cc_namespace(self, name: str, username: str) -> Dict:
        """
        尝试在 PaaSCC 中查询指定命名空间，若不存在则创建

        :param name: 命名空间名称
        :param username: 操作者
        :return: Namespace 信息
        """
        # 假定cc中有，集群中也存在
        cc_namespaces = get_namespaces_by_cluster_id(
            self.ctx_cluster.context.auth.access_token, self.ctx_cluster.project_id, self.ctx_cluster.id
        )
        for ns in cc_namespaces:
            if ns["name"] == name:
                return self._extract_namespace_info(ns)

        return self._create_namespace(username, name)

    def _create_namespace(self, creator: str, name: str) -> Dict:
        """
        在 PaaSCC 与 集群 中创建 Namespace

        :param creator: 创建者
        :param name: 新命名空间名称
        :return: Namespace 信息
        """
        # TODO 补充 imagepullsecrets 和命名空间变量的创建?
        # TODO 操作审计
        # 先在集群中创建命名空间（可能存在 PaasCC不存在但是集群存在的情况，需要预先检查），再同步至 PaaSCC
        if not self.get(name=name):
            self.create(body={"apiVersion": "v1", "kind": "Namespace", "metadata": {"name": name}}, name=name)
        namespace = create_cc_namespace(
            self.ctx_cluster.context.auth.access_token, self.ctx_cluster.project_id, self.ctx_cluster.id, name, creator
        )
        return self._extract_namespace_info(namespace)

    def _extract_namespace_info(self, namespace: Dict) -> Dict:
        """
        提取 Namespace 需要的信息

        :param namespace: Namespace 对象(来源于 PaaSCC)
        :return: 仅需要的 Namespace 信息
        """
        return {'name': namespace['name'], 'namespace_id': namespace['id']}
