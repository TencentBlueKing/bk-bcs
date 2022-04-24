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
from typing import Dict, List, Optional, Union

from backend.container_service.clusters.constants import ClusterType
from backend.resources.constants import K8sResourceKind
from backend.resources.namespace.constants import PROJ_CODE_ANNO_KEY
from backend.resources.namespace.formatter import NamespaceFormatter
from backend.resources.namespace.utils import create_cc_namespace, get_namespaces_by_cluster_id
from backend.resources.resource import ResourceClient, ResourceList
from backend.resources.utils.format import ResourceFormatter
from backend.utils.basic import getitems


class Namespace(ResourceClient):
    kind = K8sResourceKind.Namespace.value
    formatter = NamespaceFormatter()

    def list(
        self,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        cluster_type: ClusterType = ClusterType.SINGLE,
        project_code: str = None,
        **kwargs,
    ) -> Union[ResourceList, Dict]:
        """
        获取命名空间列表

        :param is_format: 是否进行格式化
        :param formatter: 额外指定的格式化器
        :param cluster_type: 集群类型（共享/联邦/独立）
        :param project_code: 项目英文名
        :return: 命名空间列表
        """
        namespaces = super().list(is_format, formatter, **kwargs)
        # 共享集群中的命名空间可能来自不同项目，需要根据 project_code 过滤
        if cluster_type == ClusterType.SHARED and project_code:
            namespaces = self._filter_shared_cluster_ns(namespaces, project_code)
        return namespaces

    def watch(
        self,
        formatter: Optional[ResourceFormatter] = None,
        cluster_type: ClusterType = ClusterType.SINGLE,
        project_code: str = None,
        **kwargs,
    ) -> List:
        """
        获取较指定的 ResourceVersion 更新的资源状态变更信息

        :param formatter: 指定的格式化器（自定义资源用）
        :param cluster_type: 集群类型（共享/联邦/独立）
        :param project_code: 项目英文名
        :return: 指定资源 watch 结果
        """
        events = super().watch(formatter, **kwargs)
        # 共享集群中的命名空间可能来自不同项目，需要根据 project_code 过滤
        if cluster_type == ClusterType.SHARED and project_code:
            events = [e for e in events if self.is_project_ns_in_shared_cluster(e['manifest'], project_code)]
        return events

    def get_or_create_cc_namespace(
        self, name: str, username: str, labels: Optional[Dict] = None, annotations: Optional[Dict] = None
    ) -> Dict:
        """
        尝试在 PaaSCC 中查询指定命名空间，若不存在则创建

        :param name: 命名空间名称
        :param username: 操作者
        :param labels: 标签
        :param annotations: 注解
        :return: Namespace 信息
        """
        # 假定cc中有，集群中也存在
        namespace_info = self.get_cc_namespace_info(name)
        if namespace_info:
            return namespace_info

        return self._create_namespace(name, username, labels, annotations)

    def get_cc_namespace_info(self, name: str) -> Dict:
        """
        获取 CC 中命名空间信息

        :param name: 命名空间名称
        :return: Namespace 信息
        """
        cc_namespaces = get_namespaces_by_cluster_id(
            self.ctx_cluster.context.auth.access_token, self.ctx_cluster.project_id, self.ctx_cluster.id
        )
        for ns in cc_namespaces:
            if ns['name'] == name:
                return self._extract_namespace_info(ns)

    def _create_namespace(
        self, name: str, creator: str, labels: Optional[Dict] = None, annotations: Optional[Dict] = None
    ) -> Dict:
        """
        在 PaaSCC 与 集群 中创建 Namespace

        :param name: 新命名空间名称
        :param creator: 创建者
        :param labels: 标签
        :param annotations: 注解
        :return: Namespace 信息
        """
        # TODO 补充 imagepullsecrets 和命名空间变量的创建?
        # 先在集群中创建命名空间（可能存在 PaasCC不存在但是集群存在的情况，需要预先检查），再同步至 PaaSCC
        if not self.get(name=name):
            manifest = {
                'apiVersion': 'v1',
                'kind': 'Namespace',
                'metadata': {'name': name, 'labels': labels or {}, 'annotations': annotations or {}},
            }
            self.create(body=manifest, name=name)
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

    def _filter_shared_cluster_ns(self, namespaces: ResourceList, project_code: str) -> Dict:
        """
        根据共享集群命名空间规则，过滤出属于指定项目的命名空间

        :param namespaces: 集群总命名空间列表
        :param project_code: 项目英文名
        :return: 过滤后的命名空间列表
        """
        namespaces = namespaces.data.to_dict()
        namespaces['items'] = [
            ns for ns in namespaces['items'] if self.is_project_ns_in_shared_cluster(ns, project_code)
        ]
        return namespaces

    @staticmethod
    def is_project_ns_in_shared_cluster(ns: Dict, project_code: str) -> bool:
        """
        检查指定的命名空间是否属于项目
        如果 annotations 中包含 io.tencent.bcs.projectcode: {project_code} 和当前项目的 code 相同，则认为属于当前项目

        :param ns: 命名空间 manifest
        :param project_code: 项目英文名
        :return: True / False
        """
        return getitems(ns, ['metadata', 'annotations', PROJ_CODE_ANNO_KEY]) == project_code
