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

from django.conf import settings

from backend.container_service.misc.depot.api import get_jfrog_account

from ..helm.models.repo import Repository
from ..helm.providers.repo_provider import add_plain_repo

logger = logging.getLogger(__name__)


# TODO: 针对harbor的先保留，待合并后，删除这一部分功能
def get_or_create_private_repo(user, project):
    # 通过harbor api创建一次项目账号，然后存储在auth中
    project_id = project.project_id
    project_code = project.project_code
    private_repos = Repository.objects.filter(name=project_code, project_id=project_id)
    repo = private_repos.first()
    if repo:
        return repo
    account = get_jfrog_account(user.token.access_token, project_code, project_id)
    repo_auth = {
        "type": "basic",
        "role": "admin",
        "credentials": {"username": account.get("user"), "password": account.get("password")},
    }
    url = f"{settings.HELM_MERELY_REPO_URL}/chartrepo/{project_code}/"
    private_repo = add_plain_repo(target_project_id=project_id, name=project_code, url=url, repo_auth=repo_auth)
    return private_repo


# 替换get_or_create_private_repo功能
try:
    from .repo_ext import get_or_create_private_repo  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
