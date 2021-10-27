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
from collections import OrderedDict

from django.core.paginator import Paginator
from rest_framework.pagination import LimitOffsetPagination as _LimitOffsetPagination

from backend.utils.response import APIResult


class LimitOffsetPagination(_LimitOffsetPagination):
    """backend的分页封装"""

    def paginate_queryset(self, queryset, request, view=None):
        """赋值view，用于下面获取message"""
        self.view = view
        return super(LimitOffsetPagination, self).paginate_queryset(queryset, request, view=view)

    def get_paginated_response(self, data):
        data = OrderedDict(
            [
                ('count', self.count),
                ('next', self.get_next_link()),
                ('previous', self.get_previous_link()),
                ('results', data),
            ]
        )

        # 按约定返回数据格式, message需要对象指定赋值
        return APIResult(data, message=getattr(self.view, 'message', ''))


# 默认每页数量
DEFAULT_PAGE_LIMIT = 5
# 单次拉取最大上限
MAX_LIMIT_COUNT = 200


def custom_paginator(raw_data, offset, limit=None):
    """使用django paginator进行分页处理"""
    limit = limit or DEFAULT_PAGE_LIMIT

    # 每页查询数量不可超过上限
    if limit > MAX_LIMIT_COUNT:
        limit = MAX_LIMIT_COUNT

    page_cls = Paginator(raw_data, limit)
    curr_page = 1
    if offset or offset == 0:
        curr_page = (offset // limit) + 1
    # 如果当前页大于总页数，返回为空
    count = page_cls.count
    if curr_page > page_cls.num_pages:
        return {"count": count, "results": []}

    # 获取当前页的数据
    curr_page_info = page_cls.page(curr_page)
    curr_page_list = curr_page_info.object_list
    return {"count": count, "results": curr_page_list}
