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
import logging
from typing import Dict, List, Optional

from attr import asdict, dataclass
from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from backend.components.base import BaseHttpClient, BkApiClient, response_handler, update_request_body
from backend.components.cc import constants

logger = logging.getLogger(__name__)


@dataclass
class PageData:
    start: int = constants.DEFAULT_START_AT
    limit: int = constants.CMDB_MAX_LIMIT
    sort: str = ""  # 排序字段


class BkCCConfig:
    """蓝鲸配置平台配置信息，提供后续使用的host， url等"""

    def __init__(self, host: str):
        # 请求域名
        self.host = host
        self.prefix_path = 'api/c/compapi/v2/cc'

        # 请求地址
        # 查询业务信息
        self.search_business_url = f'{host}/{self.prefix_path}/search_business/'
        # 查询业务拓扑
        self.search_biz_inst_topo_url = f'{host}/{self.prefix_path}/search_biz_inst_topo/'
        # 查询内部模块拓扑
        self.get_biz_internal_module_url = f'{host}/{self.prefix_path}/get_biz_internal_module/'
        # 查询业务下主机
        self.list_biz_hosts_url = f'{host}/{self.prefix_path}/list_biz_hosts/'


class BkCCAuth(AuthBase):
    """用于蓝鲸配置平台接口的鉴权校验"""

    def __init__(self, username: str, bk_supplier_account: Optional[str] = settings.BKCC_DEFAULT_SUPPLIER_ACCOUNT):
        self.bk_app_code = settings.BCS_APP_CODE
        self.bk_app_secret = settings.BCS_APP_SECRET
        self.operator = username
        self.bk_username = username
        self.bk_supplier_account = bk_supplier_account

    def __call__(self, r: PreparedRequest):
        data = {
            "bk_app_code": self.bk_app_code,
            "bk_app_secret": self.bk_app_secret,
            "bk_username": self.bk_username,
            "operator": self.operator,
        }
        if self.bk_supplier_account:
            data["bk_supplier_account"] = self.bk_supplier_account
        r.body = update_request_body(r.body, data)
        return r


class BkCCClient(BkApiClient):
    """CMDB API SDK"""

    def __init__(self, username: str, bk_supplier_account: Optional[str] = settings.BKCC_DEFAULT_SUPPLIER_ACCOUNT):
        self._config = BkCCConfig(host=settings.COMPONENT_HOST)
        self._client = BaseHttpClient(BkCCAuth(username, bk_supplier_account=bk_supplier_account))

    @response_handler(default=dict)
    def search_business(
        self,
        page: PageData,
        fields: Optional[List] = None,
        condition: Optional[Dict] = None,
        bk_supplier_account: Optional[str] = None,
    ) -> Dict:
        """
        获取业务信息

        :param page: 分页条件
        :param fields: 返回的字段
        :param condition: 查询条件
        :parma bk_supplier_account: 供应商
        :return: 返回业务信息，格式:{'count': 1, 'info': [{'id': 1}]}
        """
        url = self._config.search_business_url
        params = {
            'page': asdict(page),
            'fields': fields,
            'condition': condition,
            'bk_supplier_account': bk_supplier_account,
        }
        return self._client.request_json("POST", url, json=params)

    @response_handler(default=list)
    def search_biz_inst_topo(self, bk_biz_id: int) -> List:
        """
        获取业务拓扑信息

        :param bk_biz_id: 业务 ID
        :return: 业务拓扑信息
        """
        url = self._config.search_biz_inst_topo_url
        params = {'bk_biz_id': bk_biz_id}
        return self._client.request_json('POST', url, json=params)

    @response_handler(default=dict)
    def get_biz_internal_module(self, bk_biz_id: int) -> Dict:
        """
        查询内部模块拓扑

        :param bk_biz_id: 业务 ID
        :return: 内部模块拓扑信息
        """
        url = self._config.get_biz_internal_module_url
        params = {'bk_biz_id': bk_biz_id}
        return self._client.request_json('POST', url, json=params)

    @response_handler(default=dict)
    def list_biz_hosts(
        self,
        bk_biz_id: int,
        page: PageData,
        bk_set_ids: List,
        bk_module_ids: List,
        fields: List,
        host_property_filter: Optional[Dict] = None,
        bk_supplier_account: str = None,
    ) -> Dict:
        """
        获取业务主机信息

        :param bk_biz_id: 业务 ID
        :param page: 分页配置
        :param bk_set_ids: 集群 ID 列表
        :param bk_module_ids: 模块 ID 列表
        :param fields: 指定的字段信息
        :param host_property_filter: 主机属性组合查询条件
        :param bk_supplier_account: 供应商
        :return: 主机信息，格式：{'count': 1, 'info': [{'bk_host_innerip': '127.0.0.1'}]}
        """
        url = self._config.list_biz_hosts_url
        params = {
            'bk_biz_id': bk_biz_id,
            'page': asdict(page),
            'bk_set_ids': bk_set_ids,
            'bk_module_ids': bk_module_ids,
            'fields': fields,
            'host_property_filter': host_property_filter,
            'bk_supplier_account': bk_supplier_account,
        }
        return self._client.request_json('POST', url, json=params)
