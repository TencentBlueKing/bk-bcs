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
from typing import Dict, List

from django.conf import settings

from backend.components.base import BaseHttpClient, BkApiClient, ComponentAuth, response_handler


class DataBusApiConfig:
    """databus API 配置"""

    def __init__(self, host: str):
        self.host = host

        # 日志采集规则 API 地址
        prefix_path = 'api/c/compapi/v2/bk_log'
        self.create_collect_config_url = f'{self.host}/{prefix_path}/create_bcs_collector'
        self.update_collect_config_url = f'{self.host}/{prefix_path}/update_bcs_collector/{{rule_id}}'
        self.list_collect_configs_url = f'{self.host}/{prefix_path}/list_bcs_collector?bcs_cluster_id={{cluster_id}}'
        self.delete_collect_config_url = f'{self.host}/{prefix_path}/delete_bcs_collector/{{rule_id}}'


class LogCollectorClient(BkApiClient):
    """日志采集规则管理 Client"""

    def __init__(self, auth: ComponentAuth):
        self._config = DataBusApiConfig(settings.COMPONENT_HOST)
        self._client = BaseHttpClient(auth.to_header_api_auth())

    @response_handler()
    def create_collect_config(self, config: Dict) -> Dict:
        return self._client.request_json('POST', self._config.create_collect_config_url, json=config)

    @response_handler()
    def update_collect_config(self, config_id: int, config: Dict) -> Dict:
        url = self._config.update_collect_config_url.format(rule_id=config_id)
        return self._client.request_json('POST', url, json=config)

    @response_handler(default=list)
    def list_collect_configs(self, cluster_id: str) -> List[Dict]:
        url = self._config.list_collect_configs_url.format(cluster_id=cluster_id)
        return self._client.request_json('GET', url)

    @response_handler()
    def delete_collect_config(self, config_id: int) -> Dict:
        url = self._config.delete_collect_config_url.format(rule_id=config_id)
        return self._client.request_json('DELETE', url)
