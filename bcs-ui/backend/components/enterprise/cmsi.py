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

消息管理，用于支持向用户发送多种类型的消息，包括邮件、短信、语音通知等
"""
import base64
import logging

from django.conf import settings
from django.utils.encoding import smart_bytes, smart_str

from backend.components.utils import http_post
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


CSMI_PREFIX_PATH = "api/c/compapi/cmsi"


def common_base_request(url, data):
    """请求"""
    data.update({"bk_app_code": settings.APP_ID, "bk_app_secret": settings.APP_TOKEN, "bk_username": "100"})
    resp = http_post(url, json=data)
    if not resp.get("result"):
        logger.error(
            """{error_code} ESB errorcode: {esb_code}\ncurl -X POST -d "{data}" {url}\nresp:{resp}""".format(
                error_code=error_codes.CmsiError, esb_code=resp.get("code"), data=data, url=url, resp=resp
            )
        )
    return resp


def send_mail(title, content, receiver__username):
    url = f"{settings.COMPONENT_HOST}/{CSMI_PREFIX_PATH}/send_mail/"
    content = smart_str(base64.b64encode(smart_bytes(content)))
    data = {
        "receiver__username": receiver__username,  # 多个以逗号分隔
        "title": title,
        "content": content,
        "is_content_base64": True,
    }
    resp = common_base_request(url, data)
    return resp


def send_weixin(heading, message, receiver__username):
    url = f"{settings.COMPONENT_HOST}/{CSMI_PREFIX_PATH}/send_weixin/"
    data = {"receiver__username": receiver__username, "data": {"heading": heading, "message": message}}  # 多个以逗号分隔
    resp = common_base_request(url, data)
    return resp


def send_sms(content, receiver__username):
    url = f"{settings.COMPONENT_HOST}/{CSMI_PREFIX_PATH}/send_sms/"
    content = smart_str(base64.b64encode(smart_bytes(content)))
    data = {"receiver__username": receiver__username, "content": content, "is_content_base64": True}  # 多个以逗号分隔
    resp = common_base_request(url, data)
    return resp
