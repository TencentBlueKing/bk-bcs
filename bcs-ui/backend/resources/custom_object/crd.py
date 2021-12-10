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
from typing import Dict, List, Optional, Type, Union

from django.conf import settings

from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.constants import ClusterType
from backend.resources.constants import K8sResourceKind
from backend.resources.resource import ResourceClient, ResourceList, ResourceObj
from backend.resources.utils.format import ResourceFormatter
from backend.utils.basic import getitems

from .constants import PREFERRED_CRD_API_VERSION
from .formatter import CRDFormatter


class CrdObj(ResourceObj):
    @property
    def additional_columns(self) -> List:
        """ 获取 资源新增列 信息 """
        manifest = self.data.to_dict()
        additional_columns = getitems(manifest, 'spec.additionalPrinterColumns', [])
        # 存在时间会统一处理，因此此处直接过滤掉
        return [col for col in additional_columns if col['name'].lower() != 'age']


class CustomResourceDefinition(ResourceClient):
    kind = K8sResourceKind.CustomResourceDefinition.value
    result_type: Type['ResourceObj'] = CrdObj
    formatter = CRDFormatter()

    def __init__(self, ctx_cluster: CtxCluster, api_version: Optional[str] = PREFERRED_CRD_API_VERSION):
        super().__init__(ctx_cluster, api_version)

    def list(
        self,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        cluster_type: ClusterType = ClusterType.SINGLE,
        **kwargs,
    ) -> Union[ResourceList, Dict]:
        """
        获取 CRD 列表

        :param is_format: 是否进行格式化
        :param formatter: 额外指定的格式化器
        :param cluster_type: 集群类型（共享/联邦/独立）
        :return: 命名空间列表
        """
        crds = super().list(is_format, formatter, **kwargs)
        # 共享集群只支持部分 CRD，配置在 settings 中
        if cluster_type == ClusterType.SHARED:
            crds = self._filter_shared_cluster_enabled_crds(crds)
        return crds

    def _filter_shared_cluster_enabled_crds(self, crds: ResourceList) -> Dict:
        """
        根据配置信息，过滤出当前共享集群支持的 CRD 列表

        :param crds: 集群总 CRD 列表
        :return: 过滤后的 CRD 列表
        """
        crds = crds.data.to_dict()
        crds['items'] = [
            crd for crd in crds['items'] if getitems(crd, 'metadata.name') in settings.SHARED_CLUSTER_ENABLED_CRDS
        ]
        return crds
