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
import hashlib
import logging

import requests

from backend.components.bk_repo import BkRepoClient
from backend.components.utils import http_get

logger = logging.getLogger(__name__)


def get_incremental_charts_and_hash_value(project_code, username, password, start_time):
    """获取chart，并且计算chart对应的hash值"""
    client = BkRepoClient(username=username, password=password)
    charts_info = client.list_charts(project_code, project_code, start_time=start_time)
    if not charts_info:
        return None, None

    return charts_info, _md5(str(charts_info))


def make_requests_auth(auth):
    if auth["type"].lower() == "basic":
        return requests.auth.HTTPBasicAuth(
            username=auth["credentials"]["username"],
            password=auth["credentials"]["password"],
        )

    raise NotImplementedError(auth["type"])


def _md5(content):
    h = hashlib.md5()
    h.update(content.encode("utf-8"))
    return h.hexdigest()


def get_charts_info(url, auths):
    url = url.rstrip("/")
    # 更改为直接获取chart list，不解析index.yaml
    req_charts_url = "{url}/api/charts".format(url=url)
    try:
        if not auths:
            charts_info = http_get(req_charts_url)
        else:
            for auth in auths:
                charts_info = http_get(req_charts_url, auth=make_requests_auth(auth))
                if charts_info:
                    break

        # 兼容处理
        if "not found" in charts_info.get("error", ""):
            return (False, None, None)

    except Exception as e:
        logger.error("get charts info fail: [url=%s], error: %s", req_charts_url, str(e))
        return (False, None, None)

    # 生成MD5，主要是便于后续校验是否变动
    charts_info_hash = _md5(str(charts_info))
    return (True, charts_info, charts_info_hash)
