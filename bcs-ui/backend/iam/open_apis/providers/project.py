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
from iam.collection import FancyDict
from iam.resource.provider import ListResult, ResourceProvider
from iam.resource.utils import Page

from backend.container_service.projects.base import list_projects, page_query_projects

from .utils import get_system_token


class ProjectProvider(ResourceProvider):
    """项目 Provider"""

    def list_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        data = page_query_projects(get_system_token(), limit=page_obj.limit, offset=page_obj.offset)
        results = [{'id': p['project_id'], 'display_name': p['project_name']} for p in data['results']]
        return ListResult(results=results, count=data['count'])

    def fetch_instance_info(self, filter_obj: FancyDict, **options) -> ListResult:
        query_params = {'project_ids': ','.join(filter_obj.ids)}
        projects = list_projects(get_system_token(), query_params)
        results = [
            {
                'id': p['project_id'],
                'display_name': p['project_name'],
                '_bk_iam_approver_': [p['creator'], p.get('updater')],
            }
            for p in projects
        ]
        return ListResult(results=results, count=len(results))

    def list_instance_by_policy(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        # TODO 确认基于实例的查询是不是就是id的过滤查询
        return ListResult(results=[], count=0)

    def list_attr(self, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr_value(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def search_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """支持模糊搜索项目名"""
        data = page_query_projects(
            get_system_token(),
            limit=page_obj.limit,
            offset=page_obj.offset,
            query_params={'search_name': filter_obj.keyword},
        )
        if data and data['results']:
            results = [{'id': p['project_id'], 'display_name': p['project_name']} for p in data['results']]
            return ListResult(results=results, count=data['count'])
        return ListResult(results=[], count=0)
