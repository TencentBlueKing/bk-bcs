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

from backend.templatesets.legacy_apps.configuration.models import Template


class TemplatesetProvider(ResourceProvider):
    """模板集 Provider"""

    def list_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """
        获取模板集列表

        :param filter_obj: 查询参数。 以下为必传如: {"parent": {"id": 1}}
        :param page_obj: 分页对象
        :return: ListResult 类型的实例列表
        """
        template_qset = Template.objects.filter(project_id=filter_obj.parent['id']).values('id', 'name')
        results = [
            {'id': template['id'], 'display_name': template['name']}
            for template in template_qset[page_obj.slice_from : page_obj.slice_to]
        ]
        return ListResult(results=results, count=template_qset.count())

    def fetch_instance_info(self, filter_obj: FancyDict, **options) -> ListResult:
        """
        批量获取模板集属性详情

        :param filter_obj: 查询参数
        :return: ListResult 类型的实例列表
        """
        template_qset = Template.objects.filter(id__in=filter_obj.ids).values('id', 'name', 'creator')
        results = [
            {'id': template['id'], 'display_name': template['name'], '_bk_iam_approver_': [template['creator']]}
            for template in template_qset
        ]
        return ListResult(results=results, count=template_qset.count())

    def list_instance_by_policy(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr(self, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr_value(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def search_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """支持模糊搜索模板集名称"""
        template_qset = Template.objects.filter(
            project_id=filter_obj.parent['id'], name__icontains=filter_obj.keyword
        ).values('id', 'name')
        results = [
            {'id': template['id'], 'display_name': template['name']}
            for template in template_qset[page_obj.slice_from : page_obj.slice_to]
        ]
        return ListResult(results=results, count=template_qset.count())
