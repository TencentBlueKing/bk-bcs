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
from typing import Dict

from backend.components import ssm
from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.iam.permissions.filter import ProjectFilter

logger = logging.getLogger(__name__)


def list_auth_projects(access_token: str, username: str = '') -> Dict:
    """获取用户有权限(project_view)的所有项目"""
    if not username:
        authorization = ssm.get_authorization_by_access_token(access_token)
        username = authorization['identity']['username']

    perm_filter = ProjectFilter().make_view_perm_filter(username)
    if not perm_filter:
        return {'code': 0, 'data': []}

    # TODO 通过分页方式, 支持 any 用户查看有权限的项目
    # 如果是 any, 表示所有项目. 由于项目量过大, 优化前仅返回空列表
    if ProjectFilter.op_is_any(perm_filter):
        logger.error(f'{username} project filter match any!')
        return {'code': 0, 'data': []}

    project_id_list = perm_filter.get('value')
    if not project_id_list:
        return {'code': 0, 'data': []}

    client = PaaSCCClient(auth=ComponentAuth(access_token))
    projects = client.list_projects_by_ids(project_id_list)
    for p in projects:
        p['project_code'] = p['english_name']

    return {'code': 0, 'data': projects}
