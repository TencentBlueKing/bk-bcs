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
from typing import List, Optional, Tuple

from backend.components.bk_repo import BkRepoClient, BkRepoTokenError

logger = logging.getLogger(__name__)


def get_repo_auth(username: str, project_code: str) -> Tuple[str, str]:
    """获取仓库的授权信息，包含 username 和 password"""
    client = BkRepoClient(username=username)
    # 获取用户对应的 token
    token = {}
    try:
        token = client.get_token()
    except BkRepoTokenError as e:
        logger.exception("获取token失败, %s", e)
    # 如果可以获取到token，则直接返回
    if token:
        return (username, token["id"])
    # 生成token
    token = client.set_token(project_code)
    return (username, token["id"])


def get_compatible_repo_auth(
    username: Optional[str] = None, project_code: Optional[str] = None, auth_conf: Optional[List] = None
) -> Tuple[str, str]:
    """获取仓库 auth，兼容先前生成的admin token"""
    if username and project_code:
        return get_repo_auth(username, project_code)
    credentials = auth_conf[0]["credentials"]
    return (credentials["username"], credentials["password"])
