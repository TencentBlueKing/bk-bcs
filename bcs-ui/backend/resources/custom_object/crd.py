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
from typing import List, Optional, Type

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.constants import K8sResourceKind
from backend.utils.basic import getitems

from ..resource import ResourceClient, ResourceObj
from .constants import PREFERRED_CRD_API_VERSION
from .formatter import CRDFormatter


class CrdObj(ResourceObj):
    @property
    def additional_columns(self) -> List:
        """获取 资源新增列 信息"""
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
