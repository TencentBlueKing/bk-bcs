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

from django.utils.translation import ugettext_lazy as _

from backend.components.enterprise.harbor import HarborClient
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def get_jfrog_account(access_token, project_code, project_id, is_bk=False):
    """
    获取项目的镜像账号
    """
    client = HarborClient(access_token, project_id, project_code)
    resp = client.create_account()

    # api调用失败
    if resp.get("code") != 0:
        message = error_codes.DepotError.f(_("创建项目仓库账号失败"), replace=True)
        error_message = f'{message}, {resp.get("message", "")}'
        logger.error(error_message)
        raise error_codes.ComponentError(error_message)

    return resp.get("data")


def trans_paging_query(query):
    """
    将前端的分页信息转换为harbor API需要的:page\pageSize
    该方法只在本文件内调用
    """
    start = query.get("start")
    limit = query.get("limit")
    if start is not None and limit is not None:
        page_size = limit
        page = (start // page_size) + 1
        query["page"] = page
        query["pageSize"] = page_size
    else:
        # 不分页的地方给默认值
        query["page"] = 1
        query["pageSize"] = 100000
    return query


def get_harbor_client(query):
    """镜像相关第一批API统一方法
    该方法只在本文件内调用
    """
    project_id = query.get("projectId")
    project_code = query.get("project_code", "")
    access_token = query.pop("access_token")
    client = HarborClient(access_token, project_id, project_code)
    return client


def get_public_image_list(query):
    """
    获取公共镜像列表
    """
    query = trans_paging_query(query)
    client = get_harbor_client(query)
    return client.get_public_image(**query)


def get_project_image_list(query):
    """
    获取项目镜像列表
    """
    query = trans_paging_query(query)
    client = get_harbor_client(query)
    return client.get_project_image(**query)


def get_image_tags(access_token, project_id, project_code, offset, limit, **query_params):
    """获取镜像信息和tag列表"""
    client = HarborClient(access_token, project_id, project_code)
    resp = client.get_image_tags(**query_params)
    # 处理返回数据(harbor 的tag列表没有做分页)
    data = resp.get("data") or {}
    data["has_previous"] = False
    data["has_next"] = False
    tags = data.get("tags") or []
    for tag in tags:
        # 外部版本只有一套仓库
        tag["artifactorys"] = ["PROD"]
    return resp


def get_pub_image_info(query):
    """公共获镜像详情（tag列表信息)"""
    client = get_harbor_client(query)
    resp = client.get_image_tags(**query)
    data = resp.get("data") or {}
    resp["data"] = [data]
    return resp


def get_project_image_info(query):
    """
    获取项目镜像详情（tag列表信息）
    """
    client = get_harbor_client(query)
    resp = client.get_image_tags(**query)
    data = resp.get("data") or {}
    resp["data"] = [data]
    return resp


def create_project_path_by_api(access_token, project_id, project_code):
    """调用仓库API创建项目仓库路径
    {
        "result": true,
        "message": "success",
        "data": {
            "project_id": 6,
            "name": "test",
            "creation_time": "2018-12-25 16:13:10",
            "update_time": "2018-12-25 16:13:10",
            "repo_count": 2
        },
        "code": 0
    }
    """
    client = HarborClient(access_token, project_id, project_code)
    resp = client.create_project_path()
    # api调用失败
    if resp.get("code") != 0:
        error_message = "%s, %s" % (
            error_codes.DepotError.f(_("创建项目仓库路径失败"), replace=True),
            resp.get("message", ""),
        )
        logger.error(error_message)
        raise error_codes.ComponentError(error_message)
    return True


try:
    from .api_ext import *  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
