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

from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.bk_repo import BkRepoClient, RepoConfig, RepoData
from backend.helm.repository.serializers import RepoParamsSLZ

HELM_REPO_TYPE = "HELM"
HELM_REPO_CATEGORY = "COMPOSITE"


class RepositoryViewSet(SystemViewSet):
    def list_repos(self, request, project_id):
        """查询项目下的仓库列表"""
        client = BkRepoClient(access_token=request.user.token.access_token)

        return Response(client.list_project_repos(request.project.project_code))

    def create_repo(self, request, project_id):
        """创建仓库，纳管第三方仓库"""
        params = self.params_validate(RepoParamsSLZ)
        # 组装请求参数
        req_data = RepoData(
            name=params["name"],
            type=HELM_REPO_TYPE,
            category=HELM_REPO_CATEGORY,
            public=False,  # 纳管的仓库不允许 public
            configuration=RepoConfig(
                type=HELM_REPO_CATEGORY.lower(),
                proxy={
                    "channelList": [
                        {
                            "public": params["is_public"],
                            "name": params["name"],
                            "url": params["url"],
                            "username": params["username"],
                            "password": params["password"],
                        }
                    ]
                },
            ),
        )
        # 纳管仓库
        client = BkRepoClient(access_token=request.user.token.access_token, username=request.user.username)
        client.create_repo(request.project.project_code, req_data)

        return Response()

    def update_repo(self, request, project_id):
        """更新纳管的仓库"""
        params = self.params_validate(RepoParamsSLZ)

    def delete_repo(self, request, project_id):
        """删除纳管的仓库"""
        pass
