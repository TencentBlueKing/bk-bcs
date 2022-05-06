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
from typing import Dict, List, Optional

from attr import asdict, dataclass
from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from backend.components.base import BaseCompError, BaseHttpClient, BkApiClient, response_handler
from backend.utils.errcodes import ErrorCode

logger = logging.getLogger(__name__)


class BkRepoConfig:
    """Bk Repo 原生API地址"""

    def __init__(self):
        self.bk_repo_host = settings.BK_REPO_DOMAIN
        self.docker_repo_host = settings.DOCKER_REPO_DOMAIN
        self.helm_repo_host = settings.HELM_REPO_DOMAIN

        # 项目及仓库接口
        self.create_project = f"{self.bk_repo_host}/repository/api/project"
        self.create_repo = f"{self.bk_repo_host}/repository/api/repo"
        self.set_user_auth = f"{self.bk_repo_host}/auth/api/user/create/project"

        # 镜像相关
        self.list_images = f"{self.bk_repo_host}/docker/ext/repo/{{project_name}}/{{repo_name}}"
        self.list_image_tag = f"{self.bk_repo_host}/docker/ext/tag/{{project_name}}/{{repo_name}}/{{image_name}}"

        # 针对chart相关的接口，直接访问 repo 服务的地址
        self.list_charts = f"{self.helm_repo_host}/{{project_name}}/{{repo_name}}/api/charts"
        self.get_chart_versions = f"{self.helm_repo_host}/{{project_name}}/{{repo_name}}/api/charts/{{chart_name}}"
        self.get_chart_version_detail = (
            f"{self.helm_repo_host}/{{project_name}}/{{repo_name}}/api/charts/{{chart_name}}/{{version}}"
        )
        self.delete_chart_version = (
            f"{self.helm_repo_host}/{{project_name}}/{{repo_name}}/api/charts/{{chart_name}}/{{version}}"
        )


class BkRepoAuth(AuthBase):
    """用于调用注册到APIGW的BK Repo 系统接口的鉴权"""

    def __init__(
        self, access_token: Optional[str] = None, username: Optional[str] = None, password: Optional[str] = None
    ):
        self.access_token = access_token
        self.username = username
        self.password = password

    def __call__(self, r: PreparedRequest):
        # 添加auth参数到headers中
        r.headers.update(
            {
                "X-BKREPO-UID": self.username or settings.ADMIN_USERNAME,
                "authorization": settings.BK_REPO_AUTHORIZATION,
                "Content-Type": "application/json",
            }
        )
        # 当存在用户名和密码时，需要传递auth = (username, password)
        if self.username and self.password:
            r.prepare_auth((self.username, self.password))
        return r


class BaseRequestBkRepoError(BaseCompError):
    """Bk repo api异常基类"""

    def __str__(self):
        s = super().__str__()
        return f"request bk repo api error, {s}"


class BkRepoCreateProjectError(BaseRequestBkRepoError):
    """创建项目异常"""


class BkRepoCreateRepoError(BaseRequestBkRepoError):
    """创建仓库异常"""


class BkRepoDeleteVersionError(BaseRequestBkRepoError):
    """删除版本异常"""


@dataclass
class PageData:
    pageNumber: int = 0
    pageSize: int = 100000  # 沿用先前的默认数量


class BkRepoClient(BkApiClient):
    """访问注册到apigw的 Api"""

    PROJECT_EXIST_CODE = 251005  # 项目已经存在
    REPO_EXIST_CODE = 251007  # 仓库已经存在

    def __init__(
        self, access_token: Optional[str] = None, username: Optional[str] = None, password: Optional[str] = None
    ):
        self._config = BkRepoConfig()
        self._client = BaseHttpClient(BkRepoAuth(access_token, username, password))

    def create_project(self, project_code: str, project_name: str, description: str) -> Dict:
        """创建仓库所属项目

        :param project_code: BCS项目code
        :param project_name: BCS项目名称
        :param description: BCS项目描述
        :return: 返回项目
        """
        data = {"name": project_code, "displayName": project_name, "description": description}
        resp = self._client.request_json("POST", self._config.create_project, json=data, raise_for_status=False)
        if resp.get("code") not in [ErrorCode.NoError, self.PROJECT_EXIST_CODE]:
            raise BkRepoCreateProjectError(f"create project error, {resp.get('message')}")
        return resp

    def create_repo(self, project_code: str, repo_type: str = "HELM", is_public: bool = False) -> Dict:
        """创建仓库

        :param project_code: BCS项目code
        :param repo_type: 仓库类型，支持DOCKER, HELM, OCI
        :param is_public: 是否允许公开
        :return: 返回仓库
        """
        data = {
            "projectId": project_code,
            "name": project_code,
            "type": repo_type,
            "category": "LOCAL",
            "public": is_public,  # 容器服务项目自己的仓库
            "configuration": {"type": "local"},
        }
        resp = self._client.request_json("POST", self._config.create_repo, json=data, raise_for_status=False)
        if resp.get("code") not in [ErrorCode.NoError, self.REPO_EXIST_CODE]:
            raise BkRepoCreateRepoError(f"create repo error, {resp.get('message')}")
        return resp

    @response_handler()
    def set_auth(self, project_code: str, repo_admin_user: str, repo_admin_pwd: str) -> bool:
        """设置权限

        :param project_code: BCS项目code
        :param repo_admin_user: 仓库admin用户
        :param repo_admin_pwd: 仓库admin密码
        :return: 返回auth信息
        """
        data = {
            "admin": False,
            "name": repo_admin_user,
            "pwd": repo_admin_pwd,
            "userId": repo_admin_user,
            "asstUsers": [repo_admin_user],
            "group": True,
            "projectId": project_code,
        }
        return self._client.request_json("POST", self._config.set_user_auth, json=data, raise_for_status=False)

    def list_images(self, project_name: str, repo_name: str, page: PageData, name: Optional[str] = None) -> Dict:
        """获取镜像列表

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param page: 分页信息
        :param name: 要过来的镜像名称，默认为空，即查询所有
        """
        url = self._config.list_images.format(project_name=project_name, repo_name=repo_name)
        params = asdict(page)
        params["name"] = name
        try:
            resp = self._client.request_json("GET", url, params=params)
            return resp.get("data") or {}
        except Exception:
            return {}

    def list_image_tags(
        self, project_name: str, repo_name: str, image_name: str, page: PageData, tag: Optional[str] = None
    ) -> Dict:
        """获取镜像tag

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param image_name: 镜像名称
        :param page: 分页信息
        :param tag: 镜像tag
        """
        url = self._config.list_image_tag.format(project_name=project_name, repo_name=repo_name, image_name=image_name)
        params = asdict(page)
        params["tag"] = tag
        try:
            resp = self._client.request_json("GET", url, params=params)
            return resp.get("data") or {}
        except Exception:
            return {}

    def list_charts(self, project_name: str, repo_name: str, start_time: str = None) -> Dict:
        """获取项目下的chart

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param start_time: 增量查询的起始时间
        :return: 返回项目下的chart列表
        """
        url = self._config.list_charts.format(project_name=project_name, repo_name=repo_name)
        return self._client.request_json("GET", url, params={"startTime": start_time})

    def get_chart_versions(self, project_name: str, repo_name: str, chart_name: str) -> List:
        """获取项目下指定chart的版本列表

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :return: 返回chart版本列表
        """
        url = self._config.get_chart_versions.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name
        )
        return self._client.request_json("GET", url)

    def get_chart_version_detail(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        """获取指定chart版本的详情

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :param version: chart 版本
        :return: 返回chart版本详情，包含名称、创建时间、版本、url等
        """
        url = self._config.get_chart_version_detail.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name, version=version
        )
        return self._client.request_json("GET", url)

    def delete_chart_version(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        """删除chart版本

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :param version: chart 版本
        :return: 返回删除信息，格式: {"deleted": True}
        """
        url = self._config.delete_chart_version.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name, version=version
        )
        resp = self._client.request_json("DELETE", url, raise_for_status=False)
        if not (resp.get("deleted") or "no such file or directory" in resp.get("error", "")):
            raise BkRepoDeleteVersionError(f"delete chart version error, {resp}")
        return resp
