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
import os
from typing import Optional

from django.conf import settings
from django.contrib.auth.decorators import login_required
from django.http import Http404, HttpResponseRedirect
from django.utils.decorators import method_decorator
from django.utils.translation import ugettext_lazy as _
from django.views.decorators.clickjacking import xframe_options_exempt
from rest_framework.exceptions import (
    AuthenticationFailed,
    MethodNotAllowed,
    NotAuthenticated,
    ParseError,
    PermissionDenied,
    ValidationError,
)
from rest_framework.renderers import BrowsableAPIRenderer, JSONRenderer, TemplateHTMLRenderer
from rest_framework.response import Response
from rest_framework.views import APIView, exception_handler, set_rollback

from backend.bcs_web.middleware import get_cookie_domain_by_host
from backend.components import paas_cc
from backend.components.base import (
    BaseCompError,
    CompInternalError,
    CompParseBkCommonResponseError,
    CompRequestError,
    CompResponseError,
)
from backend.container_service.projects.base.constants import ProjectKindID
from backend.dashboard.exceptions import DashboardBaseError
from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.packages.blue_krill.web.std_error import APIError
from backend.utils import cache
from backend.utils import exceptions as backend_exceptions
from backend.utils.basic import str2bool
from backend.utils.error_codes import error_codes
from backend.utils.local import local

logger = logging.getLogger(__name__)


def one_line_error(detail):
    """Extract one line error from error dict"""
    try:
        for field, errmsg in detail.items():
            if field == "non_field_errors":
                return errmsg[0]
            else:
                return "%s: %s" % (field, errmsg[0])
    except Exception:
        return _("参数格式错误")


class CompErrorFormatter:
    """格式化 components 相关错误"""

    # 错误代码来自： utils.exceptions::ComponentError
    error_code = 4001

    def __init__(self, exc: BaseCompError):
        self.exc = exc

    def format(self) -> Response:
        """格式化 components 异常"""
        if self._use_vague_message():
            message = _("数据请求失败，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG)
        else:
            # 注意：向客户端暴露原始的异常信息，可能有敏感信息
            message = _('{}，{}').format(_("第三方请求失败"), str(self.exc))

        # 拼装 Response 响应
        data = {
            "code": self.error_code,
            "message": message,
            "data": None,
            "request_id": local.request_id,
        }
        return Response(data)

    @staticmethod
    def _use_vague_message() -> bool:
        """是否向客户端展示模糊的错误信息，避免内部配置信息泄露"""
        if not settings.DEBUG and settings.IS_COMMON_EXCEPTION_MSG:
            return True
        return False


# 规则：异常类型 -> 格式化工具类
exc_resp_formatter_map = {
    CompRequestError: CompErrorFormatter,
    CompResponseError: CompErrorFormatter,
    CompInternalError: CompErrorFormatter,
    CompParseBkCommonResponseError: CompErrorFormatter,
}


def custom_exception_handler(exc: Exception, context):
    """自定义异常处理，将不同异常转换为不同的错误返回"""
    # 轮询查找异常类型与异常处理表
    for exc_type, formatter_cls in exc_resp_formatter_map.items():
        if isinstance(exc, exc_type):
            set_rollback()
            return formatter_cls(exc).format()

    if isinstance(exc, (NotAuthenticated, AuthenticationFailed)):
        data = {
            "code": error_codes.Unauthorized.code_num,
            "data": {"login_url": {"full": settings.LOGIN_FULL, "simple": settings.LOGIN_SIMPLE}},
            "message": error_codes.Unauthorized.message,
            "request_id": local.request_id,
        }
        return Response(data, status=error_codes.Unauthorized.status_code)

    elif isinstance(exc, (ValidationError, ParseError)):
        detail = exc.detail
        if "non_field_errors" in exc.detail:
            message = detail["non_field_errors"]
        else:
            message = detail
        data = {"code": 400, "message": message, "data": None, "request_id": local.request_id}
        set_rollback()
        return Response(data, status=200, headers={})

    # 对 Dashboard 类异常做特殊处理
    elif isinstance(exc, DashboardBaseError):
        data = {"code": exc.code, "message": exc.message, "data": None, "request_id": local.request_id}
        set_rollback()
        return Response(data, status=200)

    # iam 权限校验
    elif isinstance(exc, PermissionDeniedError):
        data = {"code": exc.code, "message": "%s" % exc, "data": exc.data, "request_id": local.request_id}
        set_rollback()
        return Response(data, status=200)

    elif isinstance(exc, APIError):
        # 更改返回的状态为为自定义错误类型的状态码
        data = {"code": exc.code_num, "message": exc.message, "data": None, "request_id": local.request_id}
        set_rollback()
        return Response(data)
    elif isinstance(exc, (MethodNotAllowed, PermissionDenied)):
        data = {"code": 400, "message": exc.detail, "data": None, "request_id": local.request_id}
        set_rollback()
        return Response(data, status=200)
    elif isinstance(exc, Http404):
        data = {"code": 404, "message": _("资源未找到"), "data": None}
        set_rollback()
        return Response(data, status=200)
    elif isinstance(exc, backend_exceptions.APIError):
        data = {"code": exc.code, "message": "%s" % exc, "data": exc.data, "request_id": local.request_id}
        set_rollback()
        return Response(data, status=exc.status_code)

    # Call REST framework's default exception handler to get the standard error response.
    response = exception_handler(exc, context)
    # Use a default error code
    if response is not None:
        response.data.update(code="ERROR")

    # catch all exception, if in prod/stag mode
    if settings.IS_COMMON_EXCEPTION_MSG and not settings.DEBUG and not response:
        logger.exception("restful api unhandle exception")

        data = {
            "code": 500,
            "message": _("数据请求失败，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG),
            "data": None,
            "request_id": local.request_id,
        }
        return Response(data)

    return response


class FinalizeResponseMixin:
    def finalize_response(self, request, response, *args, **kwargs):
        if "code" not in response.data:
            code = response.status_code // 100
            response.data = {
                "code": 0 if code == 2 else response.status_code,
                "message": response.status_text,
                "data": response.data,
            }
        return super(FinalizeResponseMixin, self).finalize_response(request, response, *args, **kwargs)


class ProjectMixin:
    """从 url 中提取 project_id，使用该 mixin 前请确保 url 参数中有 project_id 这个 key"""

    @property
    def project_id(self):
        project_id = self.request.parser_context["kwargs"].get("project_id")
        if not project_id:
            return self.request.project.project_id
        return project_id


class FilterByProjectMixin(ProjectMixin):
    def get_queryset(self):
        return self.queryset.filter(project_id=self.project_id)


class AppMixin:
    """从 url 中提取 app_id，使用该 mixin 前请确保 url 参数中有 app_id 这个 key"""

    @property
    def app_id(self):
        return self.request.parser_context["kwargs"]["app_id"]


class AccessTokenMixin:
    """从 url 中获取 access_token"""

    @property
    def access_token(self):
        # FIXME maybe we should raise 401 when not login
        return self.request.user.token.access_token


class ActionSerializerMixin:
    action_serializers = {}

    def get_serializer_class(self):
        if self.action in self.action_serializers:
            return self.action_serializers.get(self.action, None)
        else:
            return super().get_serializer_class()


class CodeJSONRenderer(JSONRenderer):
    """
    采用统一的结构封装返回内容
    """

    def render(self, data, accepted_media_type=None, renderer_context=None):
        response_data = {"data": data, "code": 0, "message": "success"}
        if isinstance(data, dict):
            # helm app 的变更操作结果通过 `transitioning_result` 字段反应
            # note: 不要对GET操作的返回结果进行处理
            if (
                renderer_context is not None
                and renderer_context["request"].method in ["POST", "PUT", "DELETE"]  # noqa
                and "transitioning_result" in data  # noqa
            ):
                response_data = {
                    "data": data,
                    "code": 0 if data["transitioning_result"] is True else 400,
                    "message": data["transitioning_message"],
                }
            elif "code" not in data or "message" not in data:
                code = data.get("code", 0)
                try:
                    code = int(code)
                except Exception:
                    code = 500

                message = data.get("message", "")

                response_data = {"data": data, "code": code, "message": message}
            else:
                response_data = data

        if renderer_context:
            if renderer_context.get('web_annotations'):
                response_data['web_annotations'] = renderer_context.get('web_annotations')

        response = super(CodeJSONRenderer, self).render(response_data, accepted_media_type, renderer_context)
        return response


def with_code_wrapper(func):
    func.renderer_classes = (BrowsableAPIRenderer, CodeJSONRenderer)
    return func


def make_bkmonitor_url(project: dict) -> str:
    """蓝鲸监控跳转链接"""
    if not project:
        return ""

    url = f"{getattr(settings,'BKMONITOR_HOST', '')}/?bizId={project['cc_app_id']}#/k8s"
    return url


def make_bklog_url(project: dict) -> str:
    """蓝鲸日志平台跳转链接"""
    if not project:
        return ""

    url = f"{getattr(settings, 'BKLOG_HOST', '')}/#/retrieve/?bizId={project['cc_app_id']}"
    return url


class VueTemplateView(APIView):
    """
    # TODO 重构优化逻辑
    """

    template_name = f"{settings.EDITION}/index.html"

    container_orchestration = ""
    request_url_suffix = ""
    renderer_classes = [TemplateHTMLRenderer]
    # 去掉权限控制
    permission_classes = ()

    def initial(self, request, *args, **kwargs):
        """
        获取去除后 project_code, mesos 的路径
        """
        request_paths = self.request.get_full_path_info().lstrip('/').split("/")

        if self.container_orchestration == "mesos":
            paths = request_paths[2:]
        else:
            paths = request_paths[1:]

        self.request_url_suffix = '/'.join(paths)

        super().initial(request, *args, **kwargs)

    def is_orchestration_match(self, kind: str) -> bool:
        """是否应该跳转"""
        # URL和项目类型匹配
        if self.container_orchestration == kind:
            return True

        # 未开启BCS, 且当前是不带 mesos 的连接
        if not kind and self.container_orchestration == "k8s":
            return True

        # 未开启, mesos链接, 需要跳转
        # 2个不匹配，需要跳转
        return False

    def make_redirect_url(self, project_code: str, kind: str) -> str:
        """跳转连接"""

        if kind == "mesos":
            redirect_url = os.path.join(
                settings.DEVOPS_BCS_HOST, settings.SITE_URL, project_code, "mesos", self.request_url_suffix
            )
        else:
            redirect_url = os.path.join(
                settings.DEVOPS_BCS_HOST, settings.SITE_URL, project_code, self.request_url_suffix
            )

        logger.info(
            "vue page orchestration: %s, kind: %s, request_url: %s, redirect_url: %s",
            self.container_orchestration,
            kind,
            self.request.get_full_path_info(),
            redirect_url,
        )

        return redirect_url

    def get_project_kind(self, project: dict) -> str:
        """获取项目类型"""
        if not project:
            return ""

        # 未开启容器服务
        if project['kind'] == 0:
            return ""

        # mesos
        if project['kind'] != ProjectKindID:
            return "mesos"

        # 包含 k8s, tke
        return "k8s"

    @xframe_options_exempt
    @method_decorator(login_required(redirect_field_name="c_url"))
    def get(self, request, project_code: Optional[str] = None):

        # 缓存项目类型
        @cache.region.cache_on_arguments(expiration_time=60 * 60)
        def cached_project_info(project_code):
            """缓存项目类型"""
            result = paas_cc.get_project(request.user.token.access_token, project_code)
            if result['code'] != 0:
                return {}

            return result['data']

        project = cached_project_info(project_code)
        kind = self.get_project_kind(project)

        if not self.is_orchestration_match(kind):
            return HttpResponseRedirect(redirect_to=self.make_redirect_url(project_code, kind))

        request_domain = request.get_host().split(':')[0]
        session_cookie_domain = get_cookie_domain_by_host(settings.SESSION_COOKIE_DOMAIN, request_domain)
        context = {
            "DEVOPS_HOST": settings.DEVOPS_HOST,
            "DEVOPS_BCS_HOST": settings.DEVOPS_BCS_HOST,
            "DEVOPS_BCS_API_URL": settings.DEVOPS_BCS_API_URL,
            "DEVOPS_ARTIFACTORY_HOST": settings.DEVOPS_ARTIFACTORY_HOST,
            "BKMONITOR_URL": make_bkmonitor_url(project),  # 蓝鲸监控跳转链接
            "BKLOG_URL": make_bklog_url(project),  # 日志平台跳转链接
            "LOGIN_FULL": settings.LOGIN_FULL,
            "RUN_ENV": settings.RUN_ENV,
            # 去除末尾的 /, 前端约定
            "STATIC_URL": settings.SITE_STATIC_URL,
            # 去除开头的 . document.domain需要
            "SESSION_COOKIE_DOMAIN": session_cookie_domain.lstrip("."),
            "REGION": settings.EDITION,
            "BK_CC_HOST": settings.BK_CC_HOST,
            "SITE_URL": settings.SITE_URL[:-1],
            "BK_IAM_APP_URL": settings.BK_IAM_APP_URL,
            "SUPPORT_MESOS": str2bool(settings.SUPPORT_MESOS),
            "CONTAINER_ORCHESTRATION": "",  # 前端路由, 默认地址不变
            "BCS_API_HOST": settings.BCS_API_HOST,
        }

        # mesos 需要修改 API 和静态资源路径
        if kind == "mesos":
            context["DEVOPS_BCS_API_URL"] = os.path.join(context["DEVOPS_BCS_API_URL"], "mesos")
            context["STATIC_URL"] = os.path.join(context["STATIC_URL"], "mesos")
            context["CONTAINER_ORCHESTRATION"] = kind

        # 特定版本多域名的支持
        try:
            from .views_ext import replace_host
        except ImportError:
            pass
        else:
            context['DEVOPS_HOST'] = replace_host(context['DEVOPS_HOST'], request_domain)
            context['DEVOPS_BCS_API_URL'] = replace_host(context['DEVOPS_BCS_API_URL'], request_domain)
            context['DEVOPS_BCS_HOST'] = replace_host(context['DEVOPS_BCS_HOST'], request_domain)

        # 增加扩展的字段渲染前端页面，用于多版本
        ext_context = getattr(settings, 'EXT_CONTEXT', {})
        if ext_context:
            context.update(ext_context)

        headers = {"X-Container-Orchestration": kind.upper()}
        ext_headers = getattr(settings, 'EXT_HEADERS', {})
        if ext_headers:
            headers.update(ext_headers[session_cookie_domain])

        return Response(context, headers=headers)


class LoginSuccessView(APIView):
    template_name = f"{settings.EDITION}/login_success.html"
    renderer_classes = [TemplateHTMLRenderer]
    # 去掉权限控制
    permission_classes = ()

    @xframe_options_exempt
    @method_decorator(login_required(redirect_field_name="c_url"))
    def get(self, request):
        # 去除开头的 . document.domain需要
        request_domain = request.get_host().split(':')[0]
        session_cookie_domain = get_cookie_domain_by_host(settings.SESSION_COOKIE_DOMAIN, request_domain)
        context = {"SESSION_COOKIE_DOMAIN": session_cookie_domain.lstrip(".")}
        return Response(context)
