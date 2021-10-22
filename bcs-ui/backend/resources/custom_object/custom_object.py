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
from typing import Optional

from django.utils.translation import ugettext_lazy as _

from backend.container_service.clusters.base.models import CtxCluster
from backend.utils.error_codes import error_codes

from ..resource import ResourceClient
from .crd import CustomResourceDefinition
from .formatter import CustomObjectFormatter
from .utils import parse_cobj_api_version


class CustomObject(ResourceClient):
    formatter = CustomObjectFormatter()

    def __init__(self, ctx_cluster: CtxCluster, kind: str, api_version: Optional[str] = None):
        self.kind = kind
        super().__init__(ctx_cluster, api_version)


def get_cobj_client_by_crd(ctx_cluster: CtxCluster, crd_name: str) -> CustomObject:
    crd_client = CustomResourceDefinition(ctx_cluster)
    crd = crd_client.get(name=crd_name, is_format=False)
    if crd:
        return CustomObject(
            ctx_cluster, kind=crd.data.spec.names.kind, api_version=parse_cobj_api_version(crd.data.to_dict())
        )
    raise error_codes.ResNotFoundError(_("集群({})中未注册自定义资源({})").format(ctx_cluster.id, crd_name))
