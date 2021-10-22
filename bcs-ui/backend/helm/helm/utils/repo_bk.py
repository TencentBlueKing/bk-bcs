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

import requests
import yaml

from backend.utils.basic import md5_digest

logger = logging.getLogger(__name__)


def get_incremental_charts_and_hash_value(project_code, username, password, start_time):
    return None, None


def make_requests_auth(auth):
    if auth["type"].lower() == "basic":
        return requests.auth.HTTPBasicAuth(
            username=auth["credentials"]["username"],
            password=auth["credentials"]["password"],
        )

    raise NotImplementedError(auth["type"])


def get_charts_info(url, auths):
    url = url.rstrip("/")
    req_charts_url = "{url}/index.yaml".format(url=url)
    try:
        if not auths:
            resp = requests.get(req_charts_url, verify=False)
        else:
            for auth in auths:
                resp = requests.get(req_charts_url, auth=make_requests_auth(auth), verify=False)
                if resp.status_code != 401:
                    break

        content = resp.text
        charts_info = yaml.load(content)["entries"]

    except Exception as e:
        logger.error("get charts info fail: [url=%s], error: %s", req_charts_url, str(e))
        return (False, None, None)

    # 生成MD5，主要是便于后续校验是否变动
    charts_info_hash = md5_digest(str(charts_info))
    return (True, charts_info, charts_info_hash)


try:
    # 替换get_incremental_charts_and_hash_value, get_charts_info函数功能
    from .repo_ext import get_charts_info, get_incremental_charts_and_hash_value  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
