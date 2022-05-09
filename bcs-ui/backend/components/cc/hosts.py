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
import functools
import logging
from typing import Dict, List

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components.base import CompParseBkCommonResponseError
from backend.components.cc import constants
from backend.components.cc.business import get_app_maintainers
from backend.components.cc.client import BkCCClient, PageData
from backend.utils.async_run import AsyncRunException, async_run

logger = logging.getLogger(__name__)


class HostQueryService:
    """主机查询相关服务"""

    def __init__(
        self,
        username: str,
        bk_biz_id: int,
        bk_set_ids: List = None,
        bk_module_ids: List = None,
        host_property_filter: Dict = None,
        bk_supplier_account: str = settings.BKCC_DEFAULT_SUPPLIER_ACCOUNT,
    ):
        """
        :param username: 查询者用户名
        :param bk_biz_id: 业务 ID
        :param bk_set_ids: 集群 ID 列表
        :parma bk_module_ids: 模块 ID 列表
        :param host_property_filter: 主机属性组合查询条件
        :param bk_supplier_account: 供应商
        """
        self.cc_client = BkCCClient(username)
        self.bk_biz_id = bk_biz_id
        self.bk_set_ids = bk_set_ids
        self.bk_module_ids = bk_module_ids
        self.host_property_filter = host_property_filter
        self.bk_supplier_account = bk_supplier_account

    def _fetch_count(self) -> int:
        """查询指定条件下主机数量"""
        resp_data = self.cc_client.list_biz_hosts(
            self.bk_biz_id,
            PageData(start=constants.DEFAULT_START_AT, limit=constants.LIMIT_FOR_COUNT),
            self.bk_set_ids,
            self.bk_module_ids,
            ['bk_host_innerip'],
            self.host_property_filter,
            self.bk_supplier_account,
        )
        return resp_data['count']

    def fetch_all(self) -> List[Dict]:
        """
        并发查询 CMDB，获取符合条件的全量主机信息

        :return: 主机列表
        """
        total = self._fetch_count()
        tasks = []
        for start in range(constants.DEFAULT_START_AT, total, constants.CMDB_LIST_HOSTS_MAX_LIMIT):
            # 组装并行任务配置信息
            tasks.append(
                functools.partial(
                    self.cc_client.list_biz_hosts,
                    self.bk_biz_id,
                    PageData(
                        start=start,
                        limit=constants.CMDB_LIST_HOSTS_MAX_LIMIT,
                    ),
                    self.bk_set_ids,
                    self.bk_module_ids,
                    constants.DEFAULT_HOST_FIELDS,
                    self.host_property_filter,
                    self.bk_supplier_account,
                )
            )

        try:
            results = async_run(tasks)
        except AsyncRunException as e:
            raise CompParseBkCommonResponseError(None, _('根据条件查询业务全量主机失败：{}').format(e))

        # 所有的请求结果合并，即为全量数据
        return [host for r in results for host in r.ret['info']]


class BizTopoQueryService:
    """业务拓扑信息查询"""

    def __init__(self, username: str, bk_biz_id: int):
        """
        :param username: 用户名
        :param bk_biz_id: 业务 ID
        """
        self.cc_client = BkCCClient(username)
        self.bk_biz_id = bk_biz_id

    def _fetch_biz_inst_topo(self) -> List:
        """
        查询业务拓扑

        :return: 业务，集群，模块拓扑信息
        """
        return self.cc_client.search_biz_inst_topo(self.bk_biz_id)

    def _fetch_biz_internal_module(self) -> Dict:
        """
        查询业务的内部模块

        :return: 业务的空闲机/故障机/待回收模块
        """
        return self.cc_client.get_biz_internal_module(self.bk_biz_id)

    def fetch(self) -> List:
        """
        查询全量业务拓扑

        :return: 全量业务拓扑（包含普通拓扑，内部模块）
        """
        biz_inst_topo = self._fetch_biz_inst_topo()
        raw_inner_mod_topo = self._fetch_biz_internal_module()
        # topo 最外层为业务，如果首个业务存在即为查询结果
        if biz_inst_topo and raw_inner_mod_topo:
            # 将内部模块补充到业务下属集群首位
            inner_mod_topo = {
                'bk_obj_id': 'set',
                'bk_obj_name': _('集群'),
                'bk_inst_id': raw_inner_mod_topo['bk_set_id'],
                'bk_inst_name': raw_inner_mod_topo['bk_set_name'],
                'child': [
                    {
                        'bk_obj_id': 'module',
                        'bk_obj_name': _('模块'),
                        'bk_inst_id': mod['bk_module_id'],
                        'bk_inst_name': mod['bk_module_name'],
                        'child': [],
                    }
                    for mod in raw_inner_mod_topo['module']
                ],
            }
            biz_inst_topo[0]['child'].insert(0, inner_mod_topo)

        return biz_inst_topo


def get_has_perm_hosts(bk_biz_id: int, username: str) -> List:
    """查询业务下有权限的主机"""
    all_maintainers = get_app_maintainers(username, bk_biz_id)
    if username in all_maintainers:
        # 如果是业务运维，查询全量主机
        return HostQueryService(username, bk_biz_id).fetch_all()
    # 否则查询有主机负责人权限的主机
    return _get_hosts_by_operator(bk_biz_id, username)


def _get_hosts_by_operator(bk_biz_id: int, username: str) -> List:
    """获取 指定用户 在业务下为 主备负责人 的机器"""
    host_list = HostQueryService(username, bk_biz_id).fetch_all()
    return [h for h in host_list if username in [h.get('operator'), h.get('bk_bak_operator')]]
