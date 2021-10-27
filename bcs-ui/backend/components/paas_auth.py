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

from backend.iam import legacy_perms as permissions

from .ssm import get_client_access_token

logger = logging.getLogger(__name__)


def get_access_token():
    """获取非用户态access_token"""
    return get_client_access_token()


def get_role_list(access_token, project_id, need_user=False):
    """获取角色列表(权限中心暂时没有角色的概念，先获取所有用户)"""
    project_perm = permissions.ProjectPermission()
    users = project_perm.query_authorized_users(project_id, permissions.ProjectActions.VIEW.value)

    role_list = []
    for _u in users:
        # 所有用户都设置为项目成员
        role_list.append(
            {
                "display_name": "项目成员",
                "role_id": 0,
                "role_name": "manager",
                "user_id": _u.get("id"),
                "user_type": "user",
            }
        )
    return role_list


try:
    from .paas_auth_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
