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
from operator import itemgetter
from typing import Dict, List

from iam.collection import FancyDict
from iam.resource.provider import ListResult, ResourceProvider
from iam.resource.utils import Page

from backend.components.cluster_manager import ClusterManagerClient

from .utils import get_system_token


class CloudAccountProvider(ResourceProvider):
    """云账号资源"""

    def list_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """获取云账号列表"""
        accounts = self._list_cloud_accounts(filter_obj.parent['id'])
        return ListResult(results=accounts[page_obj.slice_from : page_obj.slice_to], count=len(accounts))

    def fetch_instance_info(self, filter_obj: FancyDict, **options) -> ListResult:
        client = ClusterManagerClient(get_system_token())
        accounts = client.list_cloud_accounts_by_ids(filter_obj.ids)
        results = [
            {
                'id': acct['accountID'],
                'display_name': acct['accountName'],
                '_bk_iam_approver_': [acct['creator'], acct['updater']],
            }
            for acct in accounts
        ]
        return ListResult(results=results, count=len(results))

    def list_instance_by_policy(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr(self, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr_value(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def search_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        accounts = self._list_cloud_accounts(filter_obj.parent['id'])
        accounts = [acct for acct in accounts if filter_obj.keyword in acct['display_name']]

        return ListResult(results=accounts[page_obj.slice_from : page_obj.slice_to], count=len(accounts))

    def _list_cloud_accounts(self, project_id: str) -> List[Dict[str, str]]:
        """根据项目 ID, 查询项目下的云账号"""
        client = ClusterManagerClient(get_system_token())
        accounts = client.list_cloud_accounts(project_id)

        if not accounts:
            return []

        accounts = sorted(accounts, key=itemgetter('updateTime'), reverse=True)

        return [{'id': acct['accountID'], 'display_name': acct['accountName']} for acct in accounts]
