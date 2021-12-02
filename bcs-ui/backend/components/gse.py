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

from django.conf import settings

from backend.components.utils import http_post
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)

GSE_HOST = settings.COMPONENT_HOST
BK_APP_CODE = settings.APP_ID
BK_APP_SECRET = settings.APP_TOKEN
PREFIX_PATH = "/api/c/compapi"
FUNCTION_PATH_MAP = {"agent_status": "/v2/gse/get_agent_status"}


def get_agent_status(username, hosts, bk_supplier_id=0):
    url = "{host}{prefix_path}{path}".format(
        host=GSE_HOST, prefix_path=PREFIX_PATH, path=FUNCTION_PATH_MAP["agent_status"]
    )

    data = {"bk_app_code": BK_APP_CODE, "bk_app_secret": BK_APP_SECRET, "bk_username": username, "hosts": hosts}
    if bk_supplier_id is not None:
        data["bk_supplier_id"] = bk_supplier_id
    resp = http_post(url, json=data)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError.f(resp.get("message"))
    return resp.get("data", {}).values()


try:
    from .gse_ext import get_agent_status  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
