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
from django.conf import settings

from backend.components.bk_repo import BkRepoClient
from backend.helm.helm.models.repo import Repository
from backend.helm.repository.constants import RepoCategory
from backend.utils.basic import getitems
from backend.utils.error_codes import error_codes


def is_imported_repo(access_token: str, username: str, project_code: str, repo_name: str) -> bool:
    """判断是否为纳管的仓库

    NOTE: 这里通过仓库列表进行匹配，没有单个仓库接口
    """
    # 获取仓库列表
    repos = BkRepoClient(access_token=access_token, username=username).list_project_repos(project_code)
    # 匹配仓库，并判断仓库是否为纳管仓库
    repo_info = {}
    for repo in repos:
        if repo["name"] != repo_name:
            continue
        repo_info = repo
    if not repo_info:
        raise error_codes.ResNotFoundError(f"repo: {repo_name} not found")
    # 判断条件: category为COMPOSITE和REMOTE，当为COMPOSITE类型时，channelList(代理源)不为空
    if repo["category"] == RepoCategory.REMOTE:
        return True
    if repo["category"] == RepoCategory.COMPOSITE and getitems(repo, ["configuration", "proxy", "channelList"]):
        return True
    return False


class RepoDBActions:
    def __init__(self, project_id: str, project_code: str, repo_name: str):
        self.project_id = project_id
        self.project_code = project_code
        self.repo_name = repo_name

    def get_or_create(self) -> Repository:
        # url 的组装格式: helm-repo-domain/project_code/
        repo_params = {
            "url": f"{settings.HELM_REPO_DOMAIN}/{self.project_code}/{self.repo_name}",
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
