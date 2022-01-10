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
import json
from typing import List

from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from .base import BaseHttpClient, BkApiClient, ComponentAuth, response_handler


class ClusterManagerConfig:
    def __init__(self, host: str):
        self.host = host

        self.get_nodes_url = f"{self.host}/bcsapi/v4/clustermanager/v1/cluster/{{cluster_id}}/node"


class ClusterManagerAuth(AuthBase):
    """用于调用 clustermanager 系统的鉴权对象"""

    def __init__(self, access_token: str):
        self.access_token = access_token

    def __call__(self, r: PreparedRequest):
        # 从配置文件读取访问系统的 admin token, 放置到请求头中
        r.headers["Authorization"] = f"Bearer {getattr(settings, 'BCS_API_GW_AUTH_TOKEN', '')}"
        r.headers["Content-Type"] = "application/json"
        r.headers['X-BKAPI-AUTHORIZATION'] = json.dumps({"access_token": self.access_token})
        return r


class ClusterManagerClient(BkApiClient):
    """访问 clustermanager 服务的 Client 对象
    :param auth: 包含校验信息的对象
    """

    def __init__(self, auth: ComponentAuth):
        self._config = ClusterManagerConfig(host=settings.CLUSTER_MANAGER_DOMAIN)
        self._client = BaseHttpClient(ClusterManagerAuth(auth.access_token))

    @response_handler(default=list)
    def get_nodes(self, cluster_id: str) -> List:
        """查询集群下的节点
        :param cluster_id: 集群ID
        :return: 返回节点列表
        """
        url = self._config.get_nodes_url.format(cluster_id=cluster_id)
        return self._client.request_json("GET", url)
