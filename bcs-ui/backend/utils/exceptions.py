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

backend 自定义错误
"""
from django.conf import settings
from django.utils.translation import ugettext_lazy as _


def get_auth_url(perms=None):
    return f'{settings.BK_IAM_APP_URL}/perm-apply/'


try:
    from .exceptions_ext import get_auth_url  # noqa
except ImportError:
    pass


class APIError(Exception):
    """所有API继承的基础"""

    # 返回的http状态
    status_code = 200

    # 返回数据中的code
    code = 400

    data = []
    # 英文编码，现在都是数字编码返回，已经没用
    # 异常名称即可以当为code_name
    # code_name = xxx

    # 错误消息前缀
    msg_prefix = ""

    def __str__(self):
        if self.args and self.msg_prefix:
            msg = "%s，%s" % (self.msg_prefix, self.args[0])
        elif self.args:
            msg = str(self.args[0])
        else:
            msg = self.msg_prefix
        return msg


class ResNotFoundError(APIError):
    """资源未找到，可以让前端显示404页面"""

    code = 404


class ComponentError(APIError):
    """第三放接口返回非200等异常
    - 前端需要提示
    - 需要自己在代码中处理
    - 如果没有处理，同时返回错误信息
    """

    # 状态是400 + x
    code = 4001
    msg_prefix = _("第三方请求失败")
    msg_in_prod = _("数据请求失败，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG)

    def __str__(self):
        # 正式环境不需要提示后端信息
        # 如果不是自定义的消息, 全部替换成 msg_in_prod
        if settings.IS_COMMON_EXCEPTION_MSG and not settings.DEBUG:
            if not self.args:
                return self.msg_in_prod

            if isinstance(self.args[0], Exception):
                # 异常替换成msg_in_prod
                return self.msg_in_prod
            else:
                # 返回自定义消息
                return str(self.args[0])

        return super(ComponentError, self).__str__()


class ValidateError(APIError):
    """参数请求错误， 需要提示前端
    和django restful的区别: 返回状态码是200
    """

    code = 400


class Rollback(APIError):
    """API回滚使用,需要捕获使用"""


class ConfigError(APIError):
    """配置文件异常,需要捕获使用"""


class NoAuthPermError(APIError):
    code = 4003

    @property
    def data(self):
        perms = self.args[1]

        d = {"perms": perms, "apply_url": get_auth_url(perms=perms)}
        return d


class PermissionDeniedError(APIError):
    # TODO: 后续根据场景再考虑是否支持 status_code == 403
    code = 4003

    @property
    def data(self):
        d = {"apply_url": self.args[1] if len(self.args) > 1 else ''}
        return d


class VerifyAuthPermError(NoAuthPermError):
    code = 4005


class VerifyAuthPermErrorWithNoRaise(NoAuthPermError):
    code = 0
