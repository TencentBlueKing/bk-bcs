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

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.custom_object import CustomResourceDefinition


def get_crd_info(crd_name: str, ctx_cluster: CtxCluster) -> Dict:
    """
    获取 CRD 基础信息

    :param crd_name: CRD 名称
    :param ctx_cluster: 集群 Context
    :return: CRD 信息，包含 kind，scope 等
    """
    return CustomResourceDefinition(ctx_cluster).get(crd_name) or {}
