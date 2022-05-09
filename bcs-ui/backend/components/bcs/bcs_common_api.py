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
from django.conf import settings

from backend.components.bcs import BCSClientBase
from backend.components.utils import http_get

CLUSTERKEEP_ENDPOINT = "{host_prefix}/v4/clusterkeeper"
# 针对特定接口的超时时间
DEFAULT_TIMEOUT = 20
DEFAULT_K8S_VERSION = "1.8.3"


class BCSClient(BCSClientBase):
    """Mesos和K8S共有的API"""

    @property
    def cluster_keeper_host(self):
        return CLUSTERKEEP_ENDPOINT.format(host_prefix=self.api_host)

    def get_events(self, params):
        """获取事件
        注意需要针对不同的环境进行查询
        """
        url = f"{settings.BCS_APIGW_DOMAIN[self._bcs_server_stag]}/bcsapi/v4/storage/events"
        resp = http_get(url, params=params, headers=self.headers)
        return resp
