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
from backend.resources.resource import ResourceClient

from .constants import KIND_RESOURCE_CLIENT_MAP


def is_native_kind(kind: str) -> bool:
    """
    是否为 K8s 原生资源类型（包含 CRD）

    :param kind: 资源类型名称
    :return: True / False
    """
    return kind in KIND_RESOURCE_CLIENT_MAP


def is_cobj_kind(kind: str) -> bool:
    """
    是否为 CRD 定义的 自定义资源 类型

    :param kind: 资源类型名称
    :return: True / False
    """
    return not is_native_kind(kind)


def get_native_kind_resource_client(kind: str) -> ResourceClient:
    """
    获取 K8s 原生资源对应的 ResourceClient

    :param kind: 资源类型名称
    :return: ResourceClient
    """
    return KIND_RESOURCE_CLIENT_MAP[kind]
