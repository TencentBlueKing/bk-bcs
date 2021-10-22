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
from functools import wraps
from urllib import parse

import six
from django.conf import settings
from django.utils.encoding import force_str
from django.utils.translation import ugettext_lazy as _
from requests.models import Response

from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.exceptions import APIError, ComponentError

logger = logging.getLogger(__name__)

# 格式化方法
FORMAT_FUNC = {
    "json": json.loads,
}

# 最大返回字符数 10KB
MAX_RESP_TEXT_SIZE = 1024 * 10

# 马赛克
MOSAIC_CHAR = "*"
MOSAIC_WORD = MOSAIC_CHAR * 3

# 敏感单词
SENSITIVE_KEYWORD = ["access_token", "bk_app_secret", "X-BKAPI-AUTHORIZATION", "X-BK-APP-SECRET"]


def get_desensitive_url(request, params):
    """获取脱敏URL"""
    if not (isinstance(params, dict) and params):
        return request.url

    desensitive_params = {}
    for key, value in params.items():
        if key in SENSITIVE_KEYWORD:
            desensitive_params[key] = MOSAIC_WORD
        else:
            desensitive_params[key] = value

    raw_url = parse.urlparse(request.url)
    desensitive_url = raw_url._replace(query=parse.urlencode(desensitive_params, safe=MOSAIC_CHAR)).geturl()
    return desensitive_url


def requests_curl_log(resp, st, params):
    """记录requests curl log"""
    if not isinstance(resp, Response):
        raise ValueError(_("返回值[{}]必须是Respose对象").format(resp))

    # 添加日志信息
    curl_req = "REQ: curl -X {method} '{url}'".format(
        method=resp.request.method, url=get_desensitive_url(resp.request, params)
    )

    if resp.request.body:
        try:
            body = json.loads(resp.request.body.decode("utf-8"))
            for s_key in SENSITIVE_KEYWORD:
                if s_key in body:
                    body[s_key] = MOSAIC_WORD
            curl_req += " -d '{body}'".format(body=json.dumps(body))
        except Exception:
            curl_req += " -d '{body}'".format(body=force_str(resp.request.body))

    if resp.request.headers:
        for key, value in resp.request.headers.items():
            # ignore headers
            if key in ["User-Agent", "Accept-Encoding", "Connection", "Accept", "Content-Length"]:
                continue
            if key == "Cookie" and value.startswith("x_host_key"):
                continue

            # 去除敏感信息, key保留, 表示鉴权信息有传递
            if key in SENSITIVE_KEYWORD:
                value = MOSAIC_WORD

            curl_req += " -H '{k}: {v}'".format(k=key, v=value)

    if len(resp.text) > MAX_RESP_TEXT_SIZE:
        resp_text = f"{resp.text[:MAX_RESP_TEXT_SIZE]}...(total {len(resp.text)} Bytes)"
    else:
        resp_text = resp.text

    curl_resp = "RESP: [%s] %.2fms %s" % (resp.status_code, (time.time() - st) * 1000, resp_text)

    logger.info("%s\n \t %s", curl_req, curl_resp)


def response(f=None, handle_resp=False):
    """返回值格式化handle_resp"""

    def decorator(func):
        @wraps(func)
        def _wrapped_func(*args, **kwargs):
            # 判断是否抛异常
            raise_exception = kwargs.pop("raise_exception", True)
            try:
                resp = func(*args, **kwargs)
                format_func = FORMAT_FUNC.get(f)
                if format_func:
                    # 获取内容
                    if isinstance(resp, Response):
                        content = resp.text
                    elif isinstance(resp, six.string_types):
                        content = resp
                    else:
                        raise ValueError(_("返回值[{}]必须是字符串或者Response对象").format(resp))

                    # 解析格式
                    err_msg = kwargs.get("err_msg", None)
                    try:
                        resp = format_func(content)
                    except Exception as e:
                        logger.exception("请求第三方失败，使用【%s】格式解析 %s 异常，%s", f, content, e)
                        error = ComponentError(err_msg or e)
                        # 设置response, 上层使用可进一步判断
                        error.response = resp
                        raise error

                if handle_resp:
                    if resp.get("code") != ErrorCode.NoError:
                        raise APIError(resp.get("message"))
                    return resp.get("data")

                return resp
            except Exception as e:
                if raise_exception:
                    raise
                return {"message": str(e)}

        return _wrapped_func

    return decorator


def parse_response_data(default_data=None, err_msg_prefix=None):
    def decorate(func):
        @wraps(func)
        def parse(*args, **kwargs):
            resp = func(*args, **kwargs)
            # NOTE: 支持code和result标识，code为0或者result为True，表示接口成功，返回data数据
            if resp.get("code") == ErrorCode.NoError or resp.get("result") is True:
                return resp.get("data") or default_data

            prefix = err_msg_prefix or f"{func.__name__} error"
            raise error_codes.APIError(f"{prefix}: {resp.get('message')}")

        return parse

    return decorate


def handle_api_not_implemented(keyword, module):
    """处理mesos API 为上线, 返回404 page not found的情况"""

    def decorate(func):
        @wraps(func)
        def handle(*args, **kwargs):
            try:
                resp = func(*args, **kwargs)
                return resp
            except ComponentError as error:
                response = getattr(error, "response", None)
                if response and keyword and keyword in response.text:
                    raise error_codes.APIError(
                        _("当前集群API版本过低, 还不支持{}，请{}").format(module, settings.COMMON_CUSTOMER_SUPPORT_MSG)
                    )
                else:
                    raise error

        return handle

    return decorate
