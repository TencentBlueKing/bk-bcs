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

Base utilities for components module
"""
import functools
import json
import logging
import time
from abc import ABC
from types import MethodType
from typing import Any, Callable, Dict, Optional
from urllib import parse

import requests
from requests import HTTPError, PreparedRequest, RequestException, Response
from requests.auth import AuthBase

from backend.utils.decorators import requests_curl_log
from backend.utils.local import local

logger = logging.getLogger(__name__)


class BaseCompError(Exception):
    """components 异常基类"""


class CompRequestError(BaseCompError):
    """请求异常错误

    :param url: 请求 URL 地址
    :param exc: 原异常对象
    """

    def __init__(self, url: str, exc: Exception):
        self.url = url
        self.exc = exc
        super().__init__(f'request error "{url}": {exc}')


class CompResponseError(BaseCompError):
    """请求响应错误

    :param url: 请求 URL 地址
    :param response: 请求响应对象
    :param exc: 原异常对象
    """

    def __init__(self, url: str, response: Response, exc: Exception):
        self.url = url
        self.response = response
        self.exc = exc
        super().__init__(f'response error "{url}": {exc}')


class CompInternalError(BaseCompError):
    """Component 模块内部错误（未知）错误"""

    def __str__(self):
        s = super().__str__()
        return f'comp internal error: {s}'


class BkApiClient(ABC):
    """抽象类：调用蓝鲸体系 API 的 Client"""


class ComponentAuth:
    """用于调用 Comp 系统的鉴权对象"""

    def __init__(self, access_token: str):
        self.access_token = access_token

    def to_header_api_auth(self) -> AuthBase:
        """转换为 requests 使用的 Header 鉴权对象"""
        return HeaderApiAuth(self.access_token)


class HeaderApiAuth(AuthBase):
    """用于调用蓝鲸体系 Api 系统的鉴权对象

    :param access_token: 通过当前用户获取的 access_token
    """

    def __init__(self, access_token: str):
        self.access_token = access_token

    def __call__(self, r: PreparedRequest):
        r.headers['X-BKAPI-AUTHORIZATION'] = json.dumps({"access_token": self.access_token})
        return r


# 使用全局的 requests 请求池
POOL_MAXSIZE = 100
_pool_adapter = requests.adapters.HTTPAdapter(pool_connections=POOL_MAXSIZE, pool_maxsize=POOL_MAXSIZE)


class BaseHttpClient:
    """发送 HTTP 请求的 Client 模块，基于 requests 模块包装而来。提供常见的工具函数，并对异常进行封装。

    :param auth: 默认 requests 身份校验对象
    """

    _default_timeout = 30
    _ssl_verify = False

    def __init__(self, auth: Optional[AuthBase] = None):
        self._auth = auth

    def request(self, method: str, url: str, **kwargs) -> Response:
        """发送 HTTP 请求

        :param method: 请求类型，GET / POST / ...
        :param url: 请求地址
        :param raise_for_status: 是否在响应状态码非 2xx 时抛出异常
        :param **kwargs: 其他请求参数
        :raises: CompRequestError（请求错误），CompInternalError（内部错误）
        """
        raise_for_status = kwargs.pop('raise_for_status', True)

        self.set_defaults_kwargs(kwargs)
        # 设置请求 request-id
        kwargs['headers']["X-Request-Id"] = local.request_id

        session = requests.session()
        session.mount('http://', _pool_adapter)
        session.mount('https://', _pool_adapter)
        try:
            started_at = time.time()
            resp = session.request(method, url, **kwargs)

            # TODO：记录 curl log 出错时，是否不应该影响主流程？
            requests_curl_log(resp, started_at, kwargs.get('params'))
        except RequestException as e:
            logger.exception('requests error when requesting {}: {}'.format(url, e))
            raise CompRequestError(url, e)
        except Exception as e:
            logger.exception('internal error when requesting {}: {}'.format(url, e))
            raise CompInternalError(str(e)) from e

        if raise_for_status:
            try:
                resp.raise_for_status()
            except HTTPError as e:
                raise CompRequestError(url, e)
        return resp

    def request_json(self, method: str, url: str, **kwargs) -> Any:
        """请求并尝试返回 Json 结果

        :raises: CompResponseError（响应错误，无法正常以 Json 解析）
        """
        resp = self.request(method, url, **kwargs)
        try:
            return resp.json()
        except Exception as e:
            raise CompResponseError(url, resp, e)

    def set_defaults_kwargs(self, kwargs: Dict[str, Any]):
        """修改请求 kwargs，设置默认的请求参数

        :param kwargs: 用于 requests.request 请求参数
        """
        kwargs.setdefault('auth', self._auth)
        kwargs.setdefault('timeout', self._default_timeout)
        kwargs.setdefault('verify', self._ssl_verify)
        kwargs.setdefault('headers', {})


def update_url_parameters(url: str, parameters: Dict) -> str:
    """更新请求地址里的 GET 参数，并返回新的地址

    :param url: 原始地址
    :param parameters: 需要追加的参数字典
    :returns: 新地址
    """
    parsed_url = parse.urlparse(url)
    orig_parameters = parse.parse_qs(parsed_url.query)
    orig_parameters.update(parameters)
    new_query = parse.urlencode(orig_parameters, doseq=True)
    # Build a new namedtuple object using new query string
    return parse.ParseResult(
        parsed_url.scheme, parsed_url.netloc, parsed_url.path, parsed_url.params, new_query, parsed_url.fragment
    ).geturl()


def update_request_body(body: Optional[bytes], params: Dict) -> bytes:
    """更新请求 body 体参数
    :param body: 原始的body
    :param params: 需要添加的参数
    :returns: 返回新的body
    """
    # body体为None时，需要设置为空字典，方便添加参数
    if not body:
        body_dict = {}
    else:
        body_dict = json.loads(bytes.decode(body))
    body_dict.update(params)
    return str.encode(json.dumps(body_dict))


class CompParseBkCommonResponseError(BaseCompError):
    """解析返回数据错误

    :param resp_json: 请求返回的数据
    :param exc: 原异常对象
    """

    def __init__(self, resp_json: Any, message: str):
        self.resp_json = json.dumps(resp_json)
        self.message = message
        super().__init__(f"parse response content error, response: {resp_json}, message: {message}")


class BkCommonResponseHandler:
    """提供解析response函数和原始response函数

    :param default_data: 当resp中data为空时，可以通过设置此字段，标识期望返回的内容
    :param func: 调用的函数
    """

    def __init__(self, default: Optional[Any] = None, func: Callable = None):
        self.default = default
        self.func = func

    def __get__(self, instance, cls):
        if instance is None:
            return self
        else:
            return MethodType(self, instance)

    def __call__(self, *args, **kwargs) -> Any:
        resp = self.func(*args, **kwargs)
        if resp.get("code") == 0 or resp.get("result") is True:
            return resp.get("data") or self._get_default()
        raise CompParseBkCommonResponseError(resp, resp.get("message"))

    def _get_default(self):
        """参考 drf Serializer Field get_default 方法，兼容可调用对象，如 dict, list"""
        return self.default() if callable(self.default) else self.default

    def raw_request(self, *args, **kwargs) -> Any:
        return self.func(*args, **kwargs)


def response_handler(default: Optional[Any] = None):
    return functools.partial(BkCommonResponseHandler, default)
