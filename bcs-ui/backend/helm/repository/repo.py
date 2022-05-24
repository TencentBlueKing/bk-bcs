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
from typing import Dict, Tuple

from django.conf import settings

from backend.components.bk_repo import BkRepoClient, BkRepoTokenError
from backend.helm.helm.models.repo import Repository
from backend.helm.repository.constants import RepoCategory
from backend.utils.basic import getitems
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def get_repo_addr(project_code: str, repo_name: str) -> str:
    """获取仓库地址"""
    # 仓库地址: helm-repo-domain/project_code/repo_name
    return f"{settings.HELM_REPO_DOMAIN}/{project_code}/{repo_name}"


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


def get_repo(access_token: str, username: str, project_code: str, repo_name: str) -> Dict:
    """获取单个repo"""
    # 通过接口列表匹配单个仓库，并返回仓库信息
    # 获取仓库列表
    repos = BkRepoClient(access_token=access_token, username=username).list_project_repos(project_code)
    # 匹配仓库，并判断仓库是否为纳管仓库
    repo_info = {}
    for repo in repos:
        if repo["name"] != repo_name:
            continue
        repo_info = repo
    # 没有查到是抛出异常
    if not repo_info:
        raise error_codes.ResNotFoundError(f"repo: {repo_name} not found")
    repo_info["is_imported"] = is_imported(repo_info)
    repo_info["addr"] = get_repo_addr(project_code, repo_info["name"])
    # 添加访问制品库的用户名和token
    repo_info["username"], repo_info["password"] = get_repo_auth(username, project_code)

    return repo_info


def is_imported(repo: Dict) -> bool:
    if repo["category"] == RepoCategory.REMOTE:
        return True
    if repo["category"] == RepoCategory.COMPOSITE and getitems(repo, ["configuration", "proxy", "channelList"]):
        return True
    return False


def is_imported_repo(access_token: str, username: str, project_code: str, repo_name: str) -> bool:
    """判断是否为纳管的仓库

    NOTE: 这里通过仓库列表进行匹配，没有单个仓库接口
    """
    repo = get_repo(access_token, username, project_code, repo_name)
    # 判断条件: category为COMPOSITE和REMOTE，当为COMPOSITE类型时，channelList(代理源)不为空
    return is_imported(repo)


class RepoDBActions:
    def __init__(self, project_id: str, project_code: str, repo_name: str):
        self.project_id = project_id
        self.project_code = project_code
        self.repo_name = repo_name

    def get_or_create(self) -> Repository:
        # url 的组装格式: helm-repo-domain/project_code/
        repo_params = {
            "url": get_repo_addr(self.project_code, self.repo_name),
            "provider": "bkrepo",
            "storage_info": {},
        }
        return Repository.objects.get_or_create(
            name=self.repo_name,
            project_id=self.project_id,
            defaults=repo_params,
        )[0]

    def delete(self) -> None:
        """删除 repo"""
        Repository.objects.filter(project_id=self.project_id, name=self.repo_name).delete()
        return
