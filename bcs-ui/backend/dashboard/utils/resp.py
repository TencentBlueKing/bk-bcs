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
from typing import Dict, Optional, Union

from django.utils.translation import ugettext_lazy as _

from backend.container_service.clusters.constants import ClusterType
from backend.dashboard.exceptions import ResourceNotExist
from backend.resources.constants import K8sResourceKind
from backend.resources.resource import ResourceClient, ResourceList
from backend.resources.utils.format import ResourceFormatter


class ListApiRespBuilder:
    """构造 Dashboard 资源列表 Api 响应内容逻辑"""

    def __init__(
        self,
        client: ResourceClient,
        formatter: Optional[ResourceFormatter] = None,
        cluster_type: ClusterType = ClusterType.SINGLE,
        project_code: str = None,
        **kwargs
    ):
        """
        构造器初始化

        :param client: 资源客户端
        :param formatter: 资源格式化器（默认使用 client.formatter）
        :param cluster_type: 集群类型（独立/共享/联邦）
        :param project_code: 集群所属项目英文名
        """
        self.client = client
        self.formatter = formatter if formatter else self.client.formatter
        # 命名空间类资源需要根据集群类型做特殊处理
        if self.client.kind == K8sResourceKind.Namespace.value:
            self.resources = self.client.list(
                is_format=False, cluster_type=cluster_type, project_code=project_code, **kwargs
            )
        else:
            self.resources = self.client.list(is_format=False, **kwargs)
        # 兼容处理，若为 ResourceList 需要将其 data (ResourceInstance) 转换成 dict
        if isinstance(self.resources, ResourceList):
            self.resources = self.resources.data.to_dict()

    def build(self) -> Dict:
        """组装 Dashboard Api 响应内容"""
        result = {'manifest': {}, 'manifest_ext': {}}
        if not self.resources:
            return result

        result['manifest'] = self.resources
        result['manifest_ext'] = {
            item['metadata']['uid']: self.formatter.format_dict(item) for item in self.resources['items']
        }
        return result


class RetrieveApiRespBuilder:
    """构造 Dashboard 资源详情 Api 响应内容逻辑"""

    def __init__(
        self,
        client: ResourceClient,
        namespace: Union[str, None],
        name: str,
        formatter: Optional[ResourceFormatter] = None,
        **kwargs
    ):
        """
        构造器初始化

        :param client: 资源客户端
        :param namespace: 资源命名空间
        :param name: 资源名称
        :param formatter: 资源格式化器（默认使用 client.formatter）
        """
        self.client = client
        self.formatter = formatter if formatter else self.client.formatter
        raw_resource = self.client.get(namespace=namespace, name=name, is_format=False, **kwargs)
        if not raw_resource:
            raise ResourceNotExist(_('资源 {}/{} 不存在').format(namespace, name))
        self.resource = raw_resource.data.to_dict()

    def build(self) -> Dict:
        """组装 Dashboard Api 响应内容"""
        return {
            'manifest': self.resource,
            'manifest_ext': self.formatter.format_dict(self.resource),
        }
