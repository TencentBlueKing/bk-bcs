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

Harbor 仓库API
"""
import json
import logging

from django.conf import settings

from backend.utils.requests import http_get, http_post

logger = logging.getLogger(__name__)

# 镜像API在apigw上注册的地址
DEPOT_API_PREFIX = "{host}{stag}".format(host=settings.DEPOT_API, stag=settings.DEPOT_STAG)

DEPOT_IMAGE_API_PREFIX = f"{DEPOT_API_PREFIX}/image"


class HarborClient:
    def __init__(self, access_token, project_id, project_code):
        self.access_token = access_token
        self.project_id = project_id
        self.project_code = project_code

        # 设置header
        bkapi_headers = {"access_token": access_token}
        # 查询公共镜像时，不需要传项目信息
        if project_id:
            bkapi_headers["project_id"] = project_id
            bkapi_headers["project_code"] = project_code
        self.headers = {"X-BKAPI-AUTHORIZATION": json.dumps(bkapi_headers)}
        self.kwargs = {"headers": self.headers, "timeout": 25}
        self.query = None

    def handle_error_msg(self, resp):
        """
        code 统一返回 0
        """
        if resp.get("code") == "00":
            resp["code"] = 0
        return resp

    def get_image_tags(self, **query_params):
        """
        获取镜像tag列表
        查询参数:imageRepo=library/k8s/kubeops/hyperkube
        """
        self.query = query_params
        self.url = f"{DEPOT_IMAGE_API_PREFIX}/listImageTags/"
        resp = http_get(self.url, params=self.query, **self.kwargs)
        self.method = "GET"
        self.handle_error_msg(resp)
        # 将tag数据放到 data['tags']中
        tags = resp.get("data") or []
        tag_count = len(tags)
        resp["data"] = {"tags": tags, "tagCount": tag_count}
        return resp

    def get_public_image(self, **query_params):
        """
        获取公共镜像列表
        参数：searchKey=&page=1&pageSize=100
        列表中已经返回了镜像详情
        """
        self.query = query_params
        self.url = f"{DEPOT_IMAGE_API_PREFIX}/listPublicImages/"
        resp = http_get(self.url, params=self.query, **self.kwargs)
        self.method = "GET"
        self.handle_error_msg(resp)
        return resp

    def get_project_image(self, **query_params):
        """
        获取项目镜像列表
        参数：searchKey=&page=1&pageSize=100
        """
        self.query = query_params
        # projectId用于传给harbor校验权限
        self.query["projectId"] = self.project_id
        self.url = f"{DEPOT_IMAGE_API_PREFIX}/{self.project_code}/listImages/"
        resp = http_get(self.url, params=self.query, **self.kwargs)
        self.method = "GET"
        self.handle_error_msg(resp)
        return resp

    def create_account(self):
        """
        创建项目账号
        """
        self.url = f"{DEPOT_IMAGE_API_PREFIX}/{self.project_code}/createAccount/"
        resp = http_post(self.url, **self.kwargs)
        self.method = "POST"
        self.handle_error_msg(resp)
        return resp

    def create_project_path(self):
        """创建项目仓库路径"""
        self.url = f"{DEPOT_API_PREFIX}/project/{self.project_code}"
        resp = http_post(self.url, **self.kwargs)
        self.method = "POST"
        self.handle_error_msg(resp)
        return resp
