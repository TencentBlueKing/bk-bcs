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

业务相关：business / application / app
"""
import functools
import logging
import re
from typing import Dict, List

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components.base import CompParseBkCommonResponseError
from backend.components.cc import constants
from backend.components.cc.client import BkCCClient, PageData
from backend.components.exceptions import BaseCompUtilError, ResourceNotFoundError
from backend.utils.async_run import AsyncRunException, async_run

logger = logging.getLogger(__name__)


class AppQueryService:
    """业务查询相关服务"""

    def __init__(
        self,
        username: str,
        fields: List = None,
        condition: Dict = None,
        bk_supplier_account: str = settings.BKCC_DEFAULT_SUPPLIER_ACCOUNT,
    ):
        """
        :param username: 查询者用户名
        :param fields: 指定字段
        :param condition: 查询条件
        :param bk_supplier_account: 供应商
        """
        self.cc_client = BkCCClient(username)
        self.fields = fields
        self.condition = condition
        self.bk_supplier_account = bk_supplier_account

    def _fetch_count(self) -> int:
        """查询指定条件下业务数量，用于后续并发查询用"""
        page = PageData(start=constants.DEFAULT_START_AT, limit=constants.LIMIT_FOR_COUNT)
        resp_data = self.cc_client.search_business(page, ['bk_biz_id'], self.condition, self.bk_supplier_account)
        return resp_data['count']

    def fetch_all(self) -> List[Dict]:
        """并发查询 CMDB，获取符合条件的全量业务信息"""
        total = self._fetch_count()
        tasks = []
        for start in range(constants.DEFAULT_START_AT, total, constants.CMDB_MAX_LIMIT):
            # 组装并行任务配置信息
            tasks.append(
                functools.partial(
                    self.cc_client.search_business,
                    PageData(start=start, limit=constants.CMDB_MAX_LIMIT),
                    self.fields,
                    self.condition,
                    self.bk_supplier_account,
                )
            )

        try:
            results = async_run(tasks)
        except AsyncRunException as e:
            raise CompParseBkCommonResponseError(None, _('根据条件查询全量业务失败：{}').format(e))

        # 所有的请求结果合并，即为全量数据
        return [app for r in results for app in r.ret['info']]

    def get(self, bk_biz_id: int) -> Dict:
        """获取单个业务信息"""
        resp_data = self.cc_client.search_business(
            PageData(),
            fields=self.fields,
            condition={'bk_biz_id': bk_biz_id},
            bk_supplier_account=self.bk_supplier_account,
        )
        if not (resp_data['count'] and resp_data['info']):
            raise ResourceNotFoundError(_('ID 为 {} 的业务不存在').format(bk_biz_id))
        return resp_data['info'][0]


def fetch_has_maintain_perm_apps(username: str) -> List[Dict]:
    """获取有运维权限的业务信息"""
    # NOTE: CMDB 建议查询方式: 以 admin 用户身份跳过资源查询权限，然后CMDB接口根据传递的 condition 中用户，返回过滤的业务
    maintainers_resp = AppQueryService(settings.ADMIN_USERNAME, condition={'bk_biz_maintainer': username}).fetch_all()
    return [{'id': item['bk_biz_id'], 'name': item['bk_biz_name']} for item in maintainers_resp]


def get_application_name(bk_biz_id: int) -> str:
    """通过业务ID，获取业务名称"""
    try:
        # 直接使用 admin 用户身份查询业务名称
        app_info = AppQueryService(settings.ADMIN_USERNAME).get(bk_biz_id)
    except BaseCompUtilError:
        return ''
    return app_info.get('bk_biz_name', '')


def get_app_maintainers(username: str, bk_biz_id: int) -> List[str]:
    """获取业务下的所有运维"""
    try:
        app_info = AppQueryService(username).get(bk_biz_id)
    except BaseCompUtilError:
        return []
    maintainers = app_info.get('bk_biz_maintainer', '')
    return re.findall(r"[^,;]+", maintainers)
