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

聚合所有 components 系统
"""
from backend.components.base import ComponentAuth
from backend.components.bcs_api import BcsApiClient

from .paas_cc import PaaSCCClient


class ComponentCollection:
    """可供使用的所有 Component 系统组合

    :param auth: 校验对象
    """

    def __init__(self, auth: ComponentAuth):
        self._auth = auth

    @property
    def paas_cc(self):
        return PaaSCCClient(auth=self._auth)

    @property
    def bcs_api(self):
        return BcsApiClient(auth=self._auth)
