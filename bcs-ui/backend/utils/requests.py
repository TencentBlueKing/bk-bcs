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
import time
from urllib.parse import urlparse

import requests

from backend.utils.exceptions import ComponentError
from backend.utils.local import local

from .decorators import requests_curl_log, response

rpool = requests.Session()
logger = logging.getLogger(__name__)

TIMEOUT = 30
SSL_VERIFY = False


def request_factory(method, handle_resp=False):
    """http请求封装"""

    @response(f="json", handle_resp=handle_resp)
    def _request(url, params=None, data=None, json=None, **kwargs):
        kwargs.setdefault("timeout", TIMEOUT)
        kwargs.setdefault("verify", SSL_VERIFY)

        # 如果第三方返回400是正常请求等，添加raise_for_status=False参数，默认非2xx类都会异常
        raise_for_status = kwargs.pop("raise_for_status", True)
        # 自定义第三接口错误信息
        err_msg = kwargs.pop("err_msg", None)

        # request_id 往下层服务透传
        headers = kwargs.pop("headers", {})
        headers["X-Request-Id"] = local.request_id

        try:
            # 记录日志改进, 404, 500等也会记录
            st = time.time()
            resp = rpool.request(method, url, params=params, data=data, json=json, headers=headers, **kwargs)
            requests_curl_log(resp, st, params)

            if raise_for_status:
                resp.raise_for_status()

            return resp
        except Exception as error:
            e_msg = f"第三方请求异常，url: {url}, {error}"
            logger.exception(e_msg)
            raise ComponentError(err_msg or error)

    return _request


def headers_for_apigw(access_token, jwt):
    if not jwt:
        return None
    return {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token, "jwt": jwt})}


def relay_request_to_mesos(request):
    """复制请求到mesos
    请求规则: url路径转换backend/mesos, 请求参数, body, 头部不动
    返回规则: 返回json
    """
    params = dict(request.GET.copy())
    data = request.body
    headers = get_headers(request.META)
    base_url = request.build_absolute_uri()
    parsed_url = urlparse(base_url)
    path = parsed_url.path
    if path.startswith("/bcs"):
        path = parsed_url.path[4:]
    relay_url = parsed_url._replace(path='/backend/mesos' + path).geturl()

    resp = rpool.request(request.method, relay_url, params=params, data=data, headers=headers)
    logger.info("relay_request_to_mesos base_url: %s, %s, %s, %s, %s", base_url, relay_url, data, headers, resp.text)
    return resp.json()


def get_headers(environ):
    """
    Retrieve the HTTP headers from a WSGI environment dictionary.  See
    https://docs.djangoproject.com/en/dev/ref/request-response/#django.http.HttpRequest.META
    """
    headers = {}
    for key, value in environ.items():
        # Sometimes, things don't like when you send the requesting host through.
        if key.startswith('HTTP_') and key != 'HTTP_HOST':
            headers[key[5:].replace('_', '-')] = value
        elif key in ('CONTENT_TYPE', 'CONTENT_LENGTH'):
            headers[key.replace('_', '-')] = value

    return headers


# raw http request
http_get = request_factory("get")
http_post = request_factory("post")
http_patch = request_factory("patch")
http_put = request_factory("put")
http_delete = request_factory("delete")

# bk http request which only return data fields in response
bk_get = request_factory("get", handle_resp=True)
bk_post = request_factory("post", handle_resp=True)
bk_patch = request_factory("patch", handle_resp=True)
bk_put = request_factory("put", handle_resp=True)
bk_delete = request_factory("delete", handle_resp=True)
