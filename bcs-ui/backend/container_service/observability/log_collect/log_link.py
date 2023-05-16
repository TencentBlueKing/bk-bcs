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
import json
from typing import Dict, Optional
from urllib.parse import urlencode

from django.conf import settings

from .models import LogIndexSet


def get_log_links(project_id: str, bk_biz_id: int, **kwargs) -> Dict:
    try:
        log_index = LogIndexSet.objects.get(project_id=project_id, bk_biz_id=bk_biz_id)
    except LogIndexSet.DoesNotExist:
        std_index_set_id = file_index_set_id = None
    else:
        std_index_set_id = log_index.std_index_set_id
        file_index_set_id = log_index.file_index_set_id

    container_ids = kwargs.get('container_ids')
    if not container_ids:
        # 返回项目维度的日志链接地址
        return {
            'std_log_link': _generate_link(bk_biz_id, std_index_set_id),
            'file_log_link': _generate_link(bk_biz_id, file_index_set_id),
        }

    log_links = {}
    for container_id in container_ids:
        query = {
            'addition': json.dumps([{'field': '__ext.container_id', 'operator': 'is', 'value': container_id}]),
        }
        log_links[container_id] = {
            'std_log_link': _generate_link(bk_biz_id, std_index_set_id, query=query),
            'file_log_link': _generate_link(bk_biz_id, file_index_set_id, query=query),
        }

    return log_links


def _generate_link(bk_biz_id: int, index_set_id: Optional[int] = None, query: Optional[Dict] = None):
    if not index_set_id:
        index_set_id = ''

    log_link_format = f'{settings.BKLOG_HOST}/#/retrieve/{{index_set_id}}'
    log_link = log_link_format.format(index_set_id=index_set_id)

    if query:
        query['bizId'] = bk_biz_id
    else:
        query = {'bizId': bk_biz_id}

    return f'{log_link}?{urlencode(query)}'
