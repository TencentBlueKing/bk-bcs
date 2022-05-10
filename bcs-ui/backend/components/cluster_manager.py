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
from typing import Any, Dict, List, Optional

from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from .base import BaseHttpClient, BkApiClient, response_handler


class ClusterManagerConfig:
    def __init__(self, host: str):
        self.host = host

        self.get_nodes_url = f"{self.host}/bcsapi/v4/clustermanager/v1/cluster/{{cluster_id}}/node"
        self.get_shared_clusters_url = f"{self.host}/bcsapi/v4/clustermanager/v1/sharedclusters"


class ClusterManagerAuth(AuthBase):
    """用于调用 clustermanager 系统的鉴权对象"""

    def __init__(self, access_token: Optional[str] = None):
        self.access_token = access_token

    def __call__(self, r: PreparedRequest):
        # 从配置文件读取访问系统的 admin token, 放置到请求头中
        r.headers["Authorization"] = f"Bearer {getattr(settings, 'BCS_APIGW_TOKEN', '')}"
        r.headers["Content-Type"] = "application/json"
        if self.access_token:
            r.headers['X-BKAPI-AUTHORIZATION'] = json.dumps({"access_token": self.access_token})
        return r


class ClusterManagerClient(BkApiClient):
    """访问 clustermanager 服务的 Client 对象
    :param auth: 包含校验信息的对象
    """

    def __init__(self, access_token: Optional[str] = None):
        self._config = ClusterManagerConfig(host=settings.CLUSTER_MANAGER_DOMAIN)
        self._client = BaseHttpClient(ClusterManagerAuth(access_token))

    @response_handler(default=list)
    def get_nodes(self, cluster_id: str) -> Dict[str, Any]:
        """查询集群下的节点
        :param cluster_id: 集群ID
        :return: 返回节点列表
        """
        url = self._config.get_nodes_url.format(cluster_id=cluster_id)
        return self._client.request_json("GET", url)

    @response_handler(default=list)
    def get_shared_clusters(self) -> Dict[str, Any]:
        """查询共享集群信息

        :return: 返回共享集群列表
        """
        # TODO 功能同步后去除
        if settings.REGION == 'ce':
            return {'code': 0, 'data': []}

        url = self._config.get_shared_clusters_url
        return self._client.request_json("GET", url)


def get_shared_clusters() -> List[Dict[str, str]]:
    """获取共享集群，仅包含集群ID、名称、集群环境"""
    clusters = ClusterManagerClient().get_shared_clusters()
    return [
        {
            "cluster_id": c["clusterID"],
            "name": c["clusterName"],
            "environment": c["environment"],
            "creator": c["creator"],
        }
        for c in clusters
    ]
