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

用户登录验证，及用户信息
"""
import logging

from django.conf import settings

from backend.components.utils import http_get
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


BK_LOGIN_PREFIX_PATH = "api/c/compapi/v2/bk_login"


def common_base_request(url, data):
    """请求"""
    data.update({"bk_app_code": settings.APP_ID, "bk_app_secret": settings.APP_TOKEN})
    resp = http_get(url, params=data)
    if not resp.get("result"):
        logger.error(
            """{error_code} ESB errorcode: {esb_code}\ncurl -X GET -d "{data}" {url}\nresp:{resp}""".format(
                error_code=error_codes.BkLoginError, esb_code=resp.get("code"), data=data, url=url, resp=resp
            )
        )
    return resp


def get_all_users():
    url = f"{settings.BK_PAAS_INNER_HOST}/{BK_LOGIN_PREFIX_PATH}/get_all_users/"
    data = {
        "bk_username": "100",
    }
    resp = common_base_request(url, data)
    return resp
