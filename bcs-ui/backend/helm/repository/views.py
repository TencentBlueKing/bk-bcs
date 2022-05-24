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

from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.bk_repo import BkRepoClient, RepoData
from backend.helm.repository import repo as bkrepo_utils
from backend.helm.repository.constants import RepoType
from backend.helm.repository.serializers import RepoParamsSLZ


class RepositoryViewSet(SystemViewSet):
    def list_repos(self, request, project_id):
        """查询项目下的仓库列表"""
        client = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username)

        # NOTE: 主动添加一次当前用户的授权，避免用户在列表页面进行操作
        project_code = request.project.project_code
        client.set_token(project_code)

        repos = client.list_project_repos(project_code)
        # 添加 是否为第三方纳管仓库 标识
        helm_repos = []
        for repo in repos:
            # 排除Docker仓库
            if repo["type"] == RepoType.DOCKER:
                continue
            repo["is_imported"] = bkrepo_utils.is_imported(repo)
            repo["addr"] = bkrepo_utils.get_repo_addr(project_code, repo["name"])
            helm_repos.append(repo)
        return Response(helm_repos)

    def create_repo(self, request, project_id):
        """创建仓库，纳管第三方仓库"""
        params = self.params_validate(RepoParamsSLZ)
        # 组装请求参数
        req_data = RepoData.from_dict(params)
        # 纳管仓库
        client = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username)
        client.create_repo(request.project.project_code, req_data)

        # 存储到平台DB，目的是兼容现阶段的关联关系
        bkrepo_utils.RepoDBActions(project_id, request.project.project_code, params["name"]).get_or_create()

        return Response()

    def update_repo(self, request, project_id, repo_name):
        """更新纳管的仓库"""
        # 必须为纳管仓库，才允许编辑
        if not bkrepo_utils.is_imported_repo(
            request.user.token.access_token, request.user.username, request.project.project_code, repo_name
        ):
            raise ValidationError(_("非纳管仓库，不允许编辑操作"))
        params = self.params_validate(RepoParamsSLZ)
        # 组装请求参数
        req_data = RepoData.from_dict(params)
        # 更新仓库
        client = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username)
        client.update_repo(request.project.project_code, req_data)

        return Response()

    def repo_detail(self, request, project_id, repo_name):
        """查询单个仓库详情"""
        # 添加访问制品库的用户名和token
        return Response(
            bkrepo_utils.get_repo(
                request.user.token.access_token, request.user.username, request.project.project_code, repo_name
            )
        )

    def delete_repo(self, request, project_id, repo_name):
        """删除纳管的仓库"""
        # 必须为纳管仓库，才允许编辑
        if not bkrepo_utils.is_imported_repo(
            request.user.token.access_token, request.user.username, request.project.project_code, repo_name
        ):
            raise ValidationError(_("非纳管仓库，不允许删除操作"))

        client = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username)
        client.delete_repo(request.project.project_code, repo_name)

        # 删除记录
        bkrepo_utils.RepoDBActions(project_id, request.project.project_code, repo_name).get_or_create()

        return Response()

    def refresh_index(self, request, project_id, repo_name):
        """刷新纳管的仓库，完成index等的更新"""
        if not bkrepo_utils.is_imported_repo(
            request.user.token.access_token, request.user.username, request.project.project_code, repo_name
        ):
            raise ValidationError(_("非纳管仓库，不允许刷新操作"))

        token = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username).get_token()
        client = BkRepoClient(
            access_token=request.user.token.access_token, username=request.user.username, password=token["id"]
        )
        client.refresh_index(request.project.project_code, repo_name)

        return Response()
