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
from django.contrib.auth.models import User

from backend.components import bk_repo
from backend.helm.helm.utils.auth import BasicAuthGenerator
from backend.utils import FancyDict
from backend.utils.error_codes import error_codes

from ..helm.models.repo import Repository, RepositoryAuth
from ..helm.providers.repo_provider import add_plain_repo

logger = logging.getLogger(__name__)


def get_or_create_private_repo(user: User, project: FancyDict):
    access_token = user.token.access_token
    username = user.username
    project_code = project.english_name
    project_id = project.project_id
    private_repos = Repository.objects.filter(name=project_code, project_id=project_id)
    # 当仓库和角色权限创建后，则直接返回
    repo = private_repos.first()
    if repo and RepositoryAuth.objects.filter(repo=repo).exists():
        return repo
    repo_client = bk_repo.BkRepoClient(username=username, access_token=access_token)
    if not repo:
        # 创建bkrepo项目
        try:
            repo_client.create_project(project_code, project.project_name, project.description)
        except bk_repo.BkRepoCreateProjectError as e:
            raise error_codes.APIError(f"create bk repo project error, {e}")
        # 创建helm repo
        try:
            repo_client.create_repo(project_code)
        except bk_repo.BkRepoCreateRepoError as e:
            raise error_codes.APIError(f"create bk repo error, {e}")
        # db中存储repo信息
        repo_params = {
            "url": f"{settings.HELM_REPO_DOMAIN}/{project_code}/{project_code}/",
            "provider": "bkrepo",
            "storage_info": {},
        }
        repo = Repository.objects.get_or_create(name=project_code, project_id=project_id, defaults=repo_params)[0]
    # 创建admin账号
    role_list = ["admin"]
    for role in role_list:
        basic_auth = BasicAuthGenerator().generate_basic_auth_by_role(role)
        repo_auth_params = {"type": "basic", "credentials": basic_auth}
        repo_client.set_auth(project_code, basic_auth["username"], basic_auth["password"])
        RepositoryAuth.objects.get_or_create(repo=repo, role=role, defaults=repo_auth_params)

    return repo
