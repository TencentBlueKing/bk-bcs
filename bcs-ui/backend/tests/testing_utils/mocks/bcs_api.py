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

from .utils import mockable_function


class StubBcsApiClient:
    """使用假数据的 BCS-Api client 对象"""

    def __init__(self, *args, **kwargs):
        pass

    @mockable_function
    def query_cluster_id(self, env_name: str, project_id: str, cluster_id: str) -> str:
        return {'id': 'faked-bcs-cluster-id-100'}

    @mockable_function
    def get_cluster_credentials(self, env_name: str, bcs_cluster_id: str) -> Dict:
        return {'server_address_path': '/foo', 'user_token': 'foo-token'}
