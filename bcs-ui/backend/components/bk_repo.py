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
import json
import logging
from typing import Any, Dict, List

from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from backend.components.base import (
    BaseCompError,
    BaseHttpClient,
    BkApiClient,
    response_handler,
    update_request_body,
    update_url_parameters,
)
from backend.utils.errcodes import ErrorCode

logger = logging.getLogger(__name__)


class BkRepoApigwConfig:
    """BK Repo Apigw 请求地址"""

    def __init__(self):
        # 请求域名
        self.host_for_apigw = getattr(settings, "BK_REPO_URL_PREFIX", "")

        # 经过apigw的请求地址
        self.create_project = f"{self.host_for_apigw}/repository/api/project"
        self.create_chart_repo = f"{self.host_for_apigw}/repository/api/repo"
        self.set_user_auth = f"{self.host_for_apigw}/auth/api/user/create/project"


class BkRepoRawConfig:
    """Bk Repo 原生API地址"""

    def __init__(self):
        self.host_for_raw_svc = getattr(settings, "HELM_MERELY_REPO_URL", "")

        # 针对chart相关的接口，直接访问 repo 服务的地址
        self.list_charts = f"{self.host_for_raw_svc}/api/{{project_name}}/{{repo_name}}/charts"
        self.get_chart_versions = f"{self.host_for_raw_svc}/api/{{project_name}}/{{repo_name}}/charts/{{chart_name}}"
        self.get_chart_version_detail = (
            f"{self.host_for_raw_svc}/api/{{project_name}}/{{repo_name}}/charts/{{chart_name}}/{{version}}"
        )
        self.delete_chart_version = (
            f"{self.host_for_raw_svc}/api/{{project_name}}/{{repo_name}}/charts/{{chart_name}}/{{version}}"
        )


try:
    from .bk_repo_ext import BkRepoRawConfig  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)


class BkRepoAuth(AuthBase):
    """用于调用注册到APIGW的BK Repo 系统接口的鉴权"""

    def __init__(self, access_token: str, username: str, password: str):
        self.access_token = access_token
        self.username = username
        self.password = password

    def __call__(self, r: PreparedRequest):
        # 添加auth参数到headers中
        r.headers.update(
            {
                "X-BKREPO-UID": self.username,
                "authorization": getattr(settings, "HELM_REPO_PLATFORM_AUTHORIZATION", ""),
                "Content-Type": "application/json",
                "X-BKAPI-AUTHORIZATION": json.dumps({"access_token": self.access_token}),
            }
        )
        # 当存在用户名和密码时，需要传递auth = (username, password)
        if self.username and self.password:
            r.prepare_auth((self.username, self.password))
        return r


class BkRepoRawAuth(AuthBase):
    """用于调用原生BK Repo 系统接口的鉴权"""

    def __init__(self, username: str, password: str):
        self.username = username
        self.password = password

    def __call__(self, r: PreparedRequest):
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


class BkRepoClient(BkApiClient):
    """访问注册到apigw的 Api"""

    PROJECT_EXIST_CODE = 251005  # 项目已经存在
    REPO_EXIST_CODE = 251007  # 仓库已经存在

    def __init__(self, username: str, access_token: str = None, password: str = None):
        self._config = BkRepoApigwConfig()
        self._bk_repo_raw_config = BkRepoRawConfig()
        self._client = BaseHttpClient(
            BkRepoAuth(access_token, username, password),
        )
        self._raw_client = BaseHttpClient(
            BkRepoRawAuth(username, password),
        )

    def create_project(self, project_code: str, project_name: str, description: str) -> Dict:
        """创建仓库所属项目

        :param project_code: BCS项目code
        :param project_name: BCS项目名称
        :param description: BCS项目描述
        :returns: 返回项目
        """
        data = {"name": project_code, "displayName": project_name, "description": description}
        resp = self._client.request_json("POST", self._config.create_project, json=data, raise_for_status=False)
        if resp.get("code") not in [ErrorCode.NoError, self.PROJECT_EXIST_CODE]:
            raise BkRepoCreateProjectError(f"create project error, {resp.get('message')}")
        return resp

    def create_chart_repo(self, project_code: str) -> Dict:
        """创建chart 仓库

        :param project_code: BCS项目code
        :returns: 返回仓库
        """
        data = {
            "projectId": project_code,
            "name": project_code,
            "type": "HELM",
            "category": "LOCAL",
            "public": False,  # 容器服务项目自己的仓库
            "configuration": {"type": "local"},
        }
        resp = self._client.request_json("POST", self._config.create_chart_repo, json=data, raise_for_status=False)
        if resp.get("code") not in [ErrorCode.NoError, self.REPO_EXIST_CODE]:
            raise BkRepoCreateRepoError(f"create repo error, {resp.get('message')}")
        return resp

    @response_handler()
    def set_auth(self, project_code: str, repo_admin_user: str, repo_admin_pwd: str) -> bool:
        """设置权限

        :param project_code: BCS项目code
        :param repo_admin_user: 仓库admin用户
        :param repo_admin_pwd: 仓库admin密码
        :returns: 返回auth信息
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

    def list_charts(self, project_name: str, repo_name: str, start_time: str = None) -> Dict:
        """获取项目下的chart

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param start_time: 增量查询的起始时间
        :returns: 返回项目下的chart列表
        """
        url = self._bk_repo_raw_config.list_charts.format(project_name=project_name, repo_name=repo_name)
        return self._raw_client.request_json("GET", url, params={"startTime": start_time})

    def get_chart_versions(self, project_name: str, repo_name: str, chart_name: str) -> List:
        """获取项目下指定chart的版本列表

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :returns: 返回chart版本列表
        """
        url = self._bk_repo_raw_config.get_chart_versions.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name
        )
        return self._raw_client.request_json("GET", url)

    def get_chart_version_detail(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        """获取指定chart版本的详情

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :param version: chart 版本
        :returns: 返回chart版本详情，包含名称、创建时间、版本、url等
        """
        url = self._bk_repo_raw_config.get_chart_version_detail.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name, version=version
        )
        return self._raw_client.request_json("GET", url)

    def delete_chart_version(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        """删除chart版本

        :param project_name: 项目名称
        :param repo_name: 仓库名称
        :param chart_name: chart 名称
        :param version: chart 版本
        :returns: 返回删除信息，格式: {"deleted": True}
        """
        url = self._bk_repo_raw_config.delete_chart_version.format(
            project_name=project_name, repo_name=repo_name, chart_name=chart_name, version=version
        )
        resp = self._raw_client.request_json("DELETE", url, raise_for_status=False)
        if not (resp.get("deleted") or "no such file or directory" in resp.get("error", "")):
            raise BkRepoDeleteVersionError(f"delete chart version error, {resp}")
        return resp
