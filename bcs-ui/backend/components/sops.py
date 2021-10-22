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
from dataclasses import asdict, dataclass, field
from typing import Dict

from django.conf import settings
from requests import PreparedRequest
from requests.auth import AuthBase

from backend.components.base import (
    BaseHttpClient,
    BkApiClient,
    response_handler,
    update_request_body,
    update_url_parameters,
)


class SopsConfig:
    """标准运维系统配置信息，提供后续使用的host， url等"""

    def __init__(self, host: str):
        # 请求域名
        self.host = host

        # 请求地址
        self.create_task_url = f"{host}/prod/create_task/{{template_id}}/{{bk_biz_id}}/"
        self.start_task_url = f"{host}/prod/start_task/{{task_id}}/{{bk_biz_id}}/"
        self.get_task_status_url = f"{host}/prod/get_task_status/{{task_id}}/{{bk_biz_id}}/"
        self.get_task_node_data_url = f"{host}/prod/get_task_node_data/{{bk_biz_id}}/{{task_id}}/"


@dataclass
class CreateTaskParams:
    name: str
    constants: Dict = field(default_factory=dict)


class SopsAuth(AuthBase):
    """用于调用 BK OP 系统接口的鉴权校验"""

    def __init__(self):
        self.app_code = settings.BCS_APP_CODE
        self.app_secret = settings.BCS_APP_SECRET
        self.bk_username = settings.ADMIN_USERNAME  # 模板所属业务的运维

    def __call__(self, r: PreparedRequest):
        # 针对get请求，添加auth参数到url中; 针对post请求，添加auth参数到body体中
        auth_params = {"app_code": self.app_code, "app_secret": self.app_secret, "bk_username": self.bk_username}
        if r.method in ["GET"]:
            r.url = update_url_parameters(r.url, auth_params)
        elif r.method in ["POST"]:
            r.body = update_request_body(r.body, auth_params)
        return r


class SopsClient(BkApiClient):
    def __init__(self):
        self._config = SopsConfig(host=settings.SOPS_API_HOST)
        self._client = BaseHttpClient(SopsAuth())

    @response_handler(default=dict)
    def create_task(self, bk_biz_id: str, template_id: str, data: CreateTaskParams) -> Dict:
        """通过业务流程创建任务

        :param bk_biz_id: 作业模板所属的业务ID
        :param template_id: 作业模板ID
        :param data: 创建任务实例需要的参数，包含名称、申请人、业务等
        :returns: 返回任务详情，包含任务连接、步骤名称、任务ID
        """
        url = self._config.create_task_url.format(template_id=template_id, bk_biz_id=bk_biz_id)
        return self._client.request_json("POST", url, json=asdict(data))

    @response_handler(default=dict)
    def start_task(self, bk_biz_id: str, task_id: str) -> Dict:
        """启动任务

        :param bk_biz_id: 作业模板所属的业务ID
        :param task_id: 任务ID
        :returns: 返回的数据中，包含任务连接
        """
        url = self._config.start_task_url.format(task_id=task_id, bk_biz_id=bk_biz_id)
        return self._client.request_json("POST", url)

    @response_handler(default=dict)
    def get_task_status(self, bk_biz_id: str, task_id: str) -> Dict:
        """获取任务状态

        :param bk_biz_id: 作业模板所属的业务ID
        :param task_id: 任务ID
        :returns: 返回任务执行状态，包含启动时间、任务状态、子步骤名称、子步骤状态等
        """
        url = self._config.get_task_status_url.format(task_id=task_id, bk_biz_id=bk_biz_id)
        return self._client.request_json("GET", url)

    @response_handler(default=dict)
    def get_task_node_data(self, bk_biz_id: str, task_id: str, node_id: str) -> Dict:
        """获取任务步骤的详情

        :param bk_biz_id: 作业模板所属的业务ID
        :param task_id: 任务ID
        :param node_id: 子步骤的ID
        :returns: 返回子步骤输出，便于解析输出，从而得到对应的value
        """
        url = self._config.get_task_node_data_url.format(task_id=task_id, bk_biz_id=bk_biz_id)
        return self._client.request_json("GET", url, params={"node_id": node_id})
