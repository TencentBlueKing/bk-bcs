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
from functools import partial
from typing import Optional

from django.utils.translation import gettext as _
from rest_framework import status

from backend.packages.blue_krill.web.std_error import APIError
from backend.packages.blue_krill.web.std_error import ErrorCode as StdErrorCode


class CallableAPIError(APIError):
    """extends `APIError` type to support callable interface for backward compatibility"""

    def __call__(self, message: Optional[str] = None, *args, **kwargs) -> 'APIError':
        """
        Replace error message entirely, the reason why the method exists is backwad compatibility,
        If you need to customize error message, use `.format(message, replace=True)` instead.

        WARNING: this method uses legacy string templating interface(with %)
        """
        if args:
            message = message % args
        elif kwargs:
            message = message % kwargs
        return self.format(message=message, replace=True)


ErrorCode = partial(StdErrorCode, error_cls=CallableAPIError)


class ErrorCodes:
    # From base_views.py
    RecordNotFound = ErrorCode(_("记录不存在"), status_code=404)
    JSONParseError = ErrorCode(_("解析异常"))
    DBOperError = ErrorCode(_("DB操作异常"))

    # TODO 禁用 APIError，该 ErrorCode 定义过于模糊，容易误用，考虑后续去除
    APIError = ErrorCode(_('请求失败'), code_num=40001)
    NoBCSService = ErrorCode(_('该项目没有使用蓝鲸容器服务'), code_num=416)
    ValidateError = ErrorCode(_('参数不正确'), code_num=40002)
    ComponentError = ErrorCode(_('第三方接口调用失败'), code_num=40003)
    JsonFormatError = ErrorCode(_('json格式错误'), code_num=40004)
    ParamMissError = ErrorCode(_('参数缺失'), code_num=40005)
    CheckFailed = ErrorCode(_('校验失败'), code_num=40006)
    ExpiredError = ErrorCode(_('资源已过期'), code_num=40007)
    # 集群资源相关异常
    ResourceError = ErrorCode(_('资源异常'), code_num=40008)
    # Helm相关错误码(前端已经在使用)
    HelmNoRegister = ErrorCode(_('集群未注册'), code_num=40031)
    HelmNoNode = ErrorCode(_('集群下没有节点'), code_num=40032)
    # 功能暂未开放
    NotOpen = ErrorCode(_('功能正在建设中'), code_num=40040)
    # 未登入, 只是定义，一般不需要使用
    Unauthorized = ErrorCode(
        _('用户未登录或登录态失效，请使用登录链接重新登录'),
        code_num=40101,
        status_code=status.HTTP_401_UNAUTHORIZED,
    )
    # 资源未找到
    ResNotFoundError = ErrorCode(_('资源未找到'), code_num=40400, status_code=status.HTTP_404_NOT_FOUND)

    ######################################
    # 打印日志使用, 1402是分配给BCS SaaS使用
    ######################################

    ConfigError = ErrorCode(_('配置{}错误'), code_num=1402400)
    # 权限中心API调用错误
    IAMError = ErrorCode(_('权限中心接口调用失败'), code_num=1402100)
    # 仓库API调用错误
    DepotError = ErrorCode(_('仓库接口调用失败'), code_num=1402101)
    # 消息管理API调用错误，请按ESB的错误码指引排查
    CmsiError = ErrorCode(_('消息管理CMSI接口调用失败'), code_num=1402102)
    # 蓝鲸登录平台API调用错误，请按ESB的错误码指引排查
    BkLoginError = ErrorCode(_('蓝鲸登录平台接口调用失败'), code_num=1402103)


error_codes = ErrorCodes()
