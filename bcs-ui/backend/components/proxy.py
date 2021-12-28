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
from urllib.parse import urljoin

import attr
from requests import PreparedRequest, Request
from requests.auth import AuthBase

from backend.components.base import BaseHttpClient, BkApiClient

DEFAULT_CONTENT_TYPE = "application/json"


class ProxyAuth(AuthBase):
    """代理接口需要的权限"""

    def __init__(self, token: Optional[str] = None, content_type: Optional[str] = DEFAULT_CONTENT_TYPE):
        self.token = token
        self.content_type = content_type

    def __call__(self, r: PreparedRequest):
        if self.token:
            r.headers['Authorization'] = f"Bearer {self.token}"
        r.headers['Content-Type'] = self.content_type
        return r


@attr.dataclass
class ProxyConfig:
    host: str
    request: Request
    prefix_path: Optional[str] = None
    token: Optional[str] = None
    content_type: Optional[str] = DEFAULT_CONTENT_TYPE
    verify_ssl: bool = False


class ProxyClient(BkApiClient):
    """访问代理服务的client

    :param proxy_data: 访问代理需要的内容，包含host、token等
    """

    def __init__(self, proxy_config: Optional[ProxyConfig] = None):
        self.proxy_config = proxy_config
        self.request = proxy_config.request
        self._client = BaseHttpClient(ProxyAuth(proxy_config.token, proxy_config.content_type))

    @property
    def source_url(self):
        """获取真实服务的url
        去除proxy配置中添加的前缀，获取真正的路径，再添加域名组装到真正路径
        """
        source_path = self.request.path
        if self.proxy_config:
            source_path = self.request.path.split(self.proxy_config.prefix_path)[-1]
        return urljoin(self.proxy_config.host, source_path)

    @property
    def request_params(self):
        """获取请求的params数据
        TODO: 后续可能会有要删除的字段
        """
        return self.request.query_params

    @property
    def request_data(self):
        """获取请求的data数据"""
        return self.request.data

    def proxy(self, **kwargs):
        return self._client.request_json(
            self.request.method,
            self.source_url,
            params=self.request_params,
            json=self.request_data,
            verify=self.proxy_config.verify_ssl,
            **kwargs,
        )
