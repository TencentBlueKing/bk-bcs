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

from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from .base import BaseHttpClient, BkApiClient, ComponentAuth, update_url_parameters


class BcsApiConfig:
    """BcsApi 系统配置对象，为 Client 提供地址等信息"""

    def __init__(self, host: str):
        self.host = host

        # BCS API 系统接口地址
        self.query_cluster_id_url = f"{host}/{{env_name}}/rest/clusters/bcs/query_by_id/"
        self.get_cluster_credentials_url = f"{host}/{{env_name}}/rest/clusters/{{bcs_cluster_id}}/client_credentials"


class BcsApiAuth(AuthBase):
    """用于调用 bcs-api 系统的鉴权对象"""

    def __init__(self, access_token: str):
        self.access_token = access_token

    def __call__(self, r: PreparedRequest):
        # 从配置文件读取访问系统的通用 Token，置入请求头中
        auth_token = f'Bearer {getattr(settings, "BCS_APIGW_TOKEN", "")}'
        r.headers['Authorization'] = auth_token
        r.headers['Content-Type'] = 'application/json'

        # 在 GET 请求参数中追加 access_token
        r.url = update_url_parameters(r.url, {'access_token': self.access_token})
        return r


class BcsApiClient(BkApiClient):
    """访问 BCS API 服务的 Client 对象

    :param auth: 包含校验信息的对象

    API 方法常用请求参数说明
    ===

    :param env_name: 集群环境，比如 stag/prod
    :param project_id: 项目 ID
    :param cluster_id: 集群 ID
    :param bcs_cluster_id: 集群在 BCS 系统中的唯一 ID
    """

    def __init__(self, auth: ComponentAuth):
        self._config = BcsApiConfig(host=settings.BCS_API_PRE_URL)
        self._client = BaseHttpClient(BcsApiAuth(auth.access_token))

    def query_cluster_id(self, env_name: str, project_id: str, cluster_id: str) -> str:
        """查询集群在 BCS-Api 中的 ID

        :returns: 集群 ID 字符串
        """
        url = self._config.query_cluster_id_url.format(env_name=env_name)
        resp = self._client.request_json(
            'GET', url, params={'project_id': project_id, 'cluster_id': cluster_id}, raise_for_status=False
        )
        return resp['id']

    def get_cluster_credentials(self, env_name: str, bcs_cluster_id: str) -> Dict:
        """
        获取访问集群 apiserver 所需的鉴权信息，比如证书、user_token、server_address_path 等

        :returns: 包含集群鉴权信息的字典
        """
        url = self._config.get_cluster_credentials_url.format(env_name=env_name, bcs_cluster_id=bcs_cluster_id)
        return self._client.request_json('GET', url, raise_for_status=False)
